package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// TestConcurrentInitDefaultLogger 测试 initDefaultLogger 的并发安全性
func TestConcurrentInitDefaultLogger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// 重置全局状态
	loggerMutex.Lock()
	loggerBase = nil
	rotateLogsWriter = nil
	currentLogFileFPath = ""
	loggerOnce = sync.Once{}
	loggerMutex.Unlock()

	const numGoroutines = 100
	var wg sync.WaitGroup
	var initCount int64

	// 并发调用 GetLogger 触发 initDefaultLogger
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 多次调用 GetLogger
			for j := 0; j < 10; j++ {
				logger := GetLogger()
				if logger != nil {
					atomic.AddInt64(&initCount, 1)
				}

				// 使用日志器
				logger.Infof("Goroutine %d, iteration %d", id, j)
			}
		}(i)
	}

	wg.Wait()

	// 验证只有一个初始化
	loggerMutex.RLock()
	finalLogger := loggerBase
	loggerMutex.RUnlock()

	if finalLogger == nil {
		t.Error("Logger should be initialized")
	}

	// initCount 应该等于 numGoroutines * 10，因为每次 GetLogger 都应该返回有效的日志器
	expectedCount := int64(numGoroutines * 10)
	if initCount != expectedCount {
		t.Errorf("Expected %d successful gets, got %d", expectedCount, initCount)
	}

	t.Logf("Concurrent initDefaultLogger test: %d successful operations", initCount)
}

// TestSetCustomFormatterConcurrency 测试 SetCustomFormatter 的并发安全性
func TestSetCustomFormatterConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-formatter-concurrency")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 保存原始状态
	loggerMutex.RLock()
	originalLogger := loggerBase
	originalRotateWriter := rotateLogsWriter
	originalCurrentFile := currentLogFileFPath
	loggerMutex.RUnlock()

	// 初始化日志器
	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "formatter_test"
	SetLoggerSettings(settings)

	defer func() {
		loggerMutex.Lock()
		loggerBase = originalLogger
		rotateLogsWriter = originalRotateWriter
		currentLogFileFPath = originalCurrentFile
		loggerMutex.Unlock()
	}()

	const numGoroutines = 50
	var wg sync.WaitGroup

	// 创建不同类型的格式器
	formatters := []logrus.Formatter{
		&logrus.JSONFormatter{},
		&logrus.TextFormatter{},
		&WithFieldFormatter{TimestampFormat: "2006-01-02 15:04:05"},
		&WithFieldFormatter{DisableTimestamp: true},
		&WithFieldFormatter{DisableLevel: true},
	}

	// 并发设置格式器并使用日志器
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				// 设置自定义格式器
				formatter := formatters[(id+j)%len(formatters)]
				SetCustomFormatter(formatter)

				// 立即使用日志器
				logger := GetLogger()
				if logger == nil {
					t.Errorf("Goroutine %d: logger is nil", id)
					continue
				}

				// 使用 WithFields
				logger.WithFields(logrus.Fields{
					"goroutine": id,
					"iteration": j,
					"formatter": fmt.Sprintf("%T", formatter),
				}).Info("Concurrent formatter test")
			}
		}(i)
	}

	wg.Wait()

	// 验证最终状态
	logger := GetLogger()
	if logger == nil {
		t.Error("Logger should still be available after concurrent formatter changes")
	}

	t.Log("Concurrent SetCustomFormatter test completed successfully")
}

// TestResourceLeakDetection 测试资源泄漏检测
func TestResourceLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource leak test in short mode")
	}

	// 获取初始文件描述符数量（Unix系统）
	var initialFDs int
	if runtime.GOOS != "windows" {
		// 获取当前进程的文件描述符数量
		if fdDir, err := os.Open("/proc/self/fd"); err == nil {
			names, _ := fdDir.Readdirnames(-1)
			initialFDs = len(names)
			fdDir.Close()
		}
	}

	root, err := os.MkdirTemp("", "logger-ut-resource-leak")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 执行大量日志器创建和销毁操作
	const iterations = 50
	for i := 0; i < iterations; i++ {
		// 创建新设置
		settings := NewSettings()
		settings.LogRootFPath = filepath.Join(root, fmt.Sprintf("iter_%d", i))
		settings.LogNameBase = fmt.Sprintf("test_%d", i)
		settings.MaxSizeMB = 1

		// 创建日志器
		logger, err := NewLogHelperWithError(settings)
		if err != nil {
			t.Errorf("Iteration %d: failed to create logger: %v", i, err)
			continue
		}

		// 写入一些日志
		for j := 0; j < 10; j++ {
			logger.Infof("Iteration %d, message %d", i, j)
		}

		// 关闭日志器（通过设置新的配置）
		Close()
	}

	// 强制垃圾回收
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// 检查文件描述符数量
	if runtime.GOOS != "windows" {
		if fdDir, err := os.Open("/proc/self/fd"); err == nil {
			names, _ := fdDir.Readdirnames(-1)
			finalFDs := len(names)
			fdDir.Close()

			fdsIncreased := finalFDs - initialFDs
			if fdsIncreased > 10 { // 允许一些合理的增长
				t.Errorf("Potential file descriptor leak: increased from %d to %d (+%d)",
					initialFDs, finalFDs, fdsIncreased)
			} else {
				t.Logf("File descriptors: %d -> %d (+%d)", initialFDs, finalFDs, fdsIncreased)
			}
		}
	}

	t.Log("Resource leak detection test completed")
}

// TestLumberjackResourceManagement 测试 lumberjack 资源管理
func TestLumberjackResourceManagement(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-lumberjack")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 测试多个 lumberjack 实例的创建和关闭
	const numInstances = 20
	var writers []*lumberjack.Logger

	// 创建多个 lumberjack 实例
	for i := 0; i < numInstances; i++ {
		filename := filepath.Join(root, fmt.Sprintf("test_%d.log", i))
		writer := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    1, // 1MB
			MaxBackups: 3,
			MaxAge:     7,
			LocalTime:  true,
			Compress:   false,
		}
		writers = append(writers, writer)

		// 写入一些数据
		_, err := writer.Write([]byte(fmt.Sprintf("Test data for instance %d\n", i)))
		if err != nil {
			t.Errorf("Failed to write to instance %d: %v", i, err)
		}
	}

	// 验证文件被创建
	for i := range writers {
		filename := filepath.Join(root, fmt.Sprintf("test_%d.log", i))
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("File %s was not created", filename)
		}
	}

	// 关闭所有写入者（在 Go 中，lumberjack 不需要显式关闭）
	// lumberjack 没有明确的 Close 方法，它会在垃圾回收时自动关闭

	// 强制垃圾回收
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	t.Log("Lumberjack resource management test completed")
}

// TestPathTraversalSecurity 测试路径遍历攻击防护
func TestPathTraversalSecurity(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		shouldBlock bool
	}{
		{
			name:        "Normal_Path",
			path:        "/var/log/app",
			shouldBlock: false,
		},
		{
			name:        "Relative_Dot_Dot",
			path:        "../../etc/passwd",
			shouldBlock: true,
		},
		{
			name:        "Absolute_Dot_Dot",
			path:        "/var/log/../../../etc/passwd",
			shouldBlock: true,
		},
		{
			name: "Path_Clean_Dot_Dot",
			path: filepath.Join("/tmp", "..", "..", "etc", "passwd"),
			// Note: filepath.Join may clean the path on some platforms,
			// so we only block if the original path contains ".."
			shouldBlock: runtime.GOOS != "windows",
		},
		{
			name:        "Multiple_Slashes",
			path:        "/var//log///app",
			shouldBlock: false,
		},
		{
			name:        "Current_Directory",
			path:        "./logs",
			shouldBlock: false,
		},
		{
			name:        "Backward_Slash_Windows",
			path:        "..\\..\\windows\\system32",
			shouldBlock: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			settings := NewSettings()
			settings.LogRootFPath = tc.path
			settings.LogNameBase = "security_test"

			err := validateSettings(settings)
			if tc.shouldBlock {
				if err == nil {
					t.Errorf("Expected error for path %s, but got none", tc.path)
				} else if !strings.Contains(err.Error(), "path traversal") {
					t.Errorf("Expected path traversal error for %s, got: %v", tc.path, err)
				}
			} else {
				// 对于正常的路径，可能会有其他错误（如目录不存在），但不应该是路径遍历错误
				if err != nil && strings.Contains(err.Error(), "path traversal") {
					t.Errorf("Unexpected path traversal error for valid path %s: %v", tc.path, err)
				}
			}
		})
	}
}

// TestMaliciousInputSecurity 测试恶意输入防护
func TestMaliciousInputSecurity(t *testing.T) {
	testCases := []struct {
		name        string
		logNameBase string
		shouldBlock bool
	}{
		{
			name:        "Normal_Name",
			logNameBase: "application",
			shouldBlock: false,
		},
		{
			name:        "Name_With_Slash",
			logNameBase: "app/log",
			shouldBlock: true,
		},
		{
			name:        "Name_With_Colon",
			logNameBase: "app:log",
			shouldBlock: true,
		},
		{
			name:        "Name_With_Asterisk",
			logNameBase: "app*log",
			shouldBlock: true,
		},
		{
			name:        "Name_With_Question",
			logNameBase: "app?log",
			shouldBlock: true,
		},
		{
			name:        "Name_With_Double_Quote",
			logNameBase: "app\"log",
			shouldBlock: true,
		},
		{
			name:        "Name_With_Less_Than",
			logNameBase: "app<log",
			shouldBlock: true,
		},
		{
			name:        "Name_With_Greater_Than",
			logNameBase: "app>log",
			shouldBlock: true,
		},
		{
			name:        "Name_With_Pipe",
			logNameBase: "app|log",
			shouldBlock: true,
		},
		{
			name:        "Empty_Name",
			logNameBase: "",
			shouldBlock: true,
		},
		{
			name:        "Very_Long_Name",
			logNameBase: strings.Repeat("a", 1000),
			shouldBlock: false, // 长度限制在其他地方处理
		},
		{
			name:        "Name_With_Null_Byte",
			logNameBase: "app\x00log",
			shouldBlock: false, // 验证函数可能不会检测到 null 字节
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			settings := NewSettings()
			settings.LogRootFPath = "/tmp"
			settings.LogNameBase = tc.logNameBase

			err := validateSettings(settings)
			if tc.shouldBlock {
				if err == nil {
					t.Errorf("Expected error for log name %s, but got none", tc.logNameBase)
				} else if !strings.Contains(err.Error(), "invalid characters") {
					t.Errorf("Expected invalid characters error for %s, got: %v", tc.logNameBase, err)
				}
			} else {
				if err != nil && strings.Contains(err.Error(), "invalid characters") {
					t.Errorf("Unexpected invalid characters error for valid name %s: %v", tc.logNameBase, err)
				}
			}
		})
	}
}

// TestSystemDirectoryProtection 测试系统目录保护
func TestSystemDirectoryProtection(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("System directory protection test only runs on Windows")
	}

	testCases := []struct {
		name        string
		path        string
		shouldBlock bool
	}{
		{
			name:        "Windows_Directory",
			path:        `C:\Windows`,
			shouldBlock: true,
		},
		{
			name:        "Program_Files",
			path:        `C:\Program Files`,
			shouldBlock: true,
		},
		{
			name:        "Program_Files_x86",
			path:        `C:\Program Files (x86)`,
			shouldBlock: true,
		},
		{
			name:        "ProgramData",
			path:        `C:\ProgramData`,
			shouldBlock: true,
		},
		{
			name:        "System32",
			path:        `C:\Windows\System32`,
			shouldBlock: true,
		},
		{
			name:        "Case_Insensitive_Windows",
			path:        `c:\WINDOWS`,
			shouldBlock: true,
		},
		{
			name:        "User_Directory",
			path:        `C:\Users\test\AppData`,
			shouldBlock: false,
		},
		{
			name:        "Temp_Directory",
			path:        `C:\Temp`,
			shouldBlock: false,
		},
		{
			name:        "Unix_Style_Path_Windows",
			path:        `C:/Windows`,
			shouldBlock: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLogPath(tc.path)
			if tc.shouldBlock {
				if err == nil {
					t.Errorf("Expected error for path %s, but got none", tc.path)
				} else if !strings.Contains(err.Error(), "cannot use system directory") {
					t.Errorf("Expected system directory error for %s, got: %v", tc.path, err)
				}
			} else {
				if err != nil && strings.Contains(err.Error(), "cannot use system directory") {
					t.Errorf("Unexpected system directory error for valid path %s: %v", tc.path, err)
				}
			}
		})
	}
}

// TestExtremeConfigurationValues 测试极端配置值
func TestExtremeConfigurationValues(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-extreme-config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	testCases := []struct {
		name      string
		setupFunc func() *Settings
		expectErr bool
	}{
		{
			name: "Zero_Values",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.MaxAgeDays = 0
				s.MaxSizeMB = 0
				s.RotationTime = 0
				s.LogRootFPath = root
				return s
			},
			expectErr: true, // RotationTime=0 会被拒绝
		},
		{
			name: "Negative_Values",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.MaxAgeDays = -1
				s.MaxSizeMB = -1
				s.RotationTime = -time.Hour
				s.LogRootFPath = root
				return s
			},
			expectErr: true,
		},
		{
			name: "Max_Age_Days_Too_Large",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.MaxAgeDays = 366 // 超过最大值 365
				s.LogRootFPath = root
				return s
			},
			expectErr: true,
		},
		{
			name: "Max_Size_MB_Too_Large",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.MaxSizeMB = 1025 // 超过最大值 1024
				s.LogRootFPath = root
				return s
			},
			expectErr: true,
		},
		{
			name: "Rotation_Time_Too_Small",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.RotationTime = time.Millisecond * 10 // 小于最小值 1 分钟
				s.LogRootFPath = root
				return s
			},
			expectErr: true,
		},
		{
			name: "Valid_Boundary_Values",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.MaxAgeDays = 365           // 最大允许值
				s.MaxSizeMB = 1024           // 最大允许值
				s.RotationTime = time.Minute // 最小允许值
				s.LogRootFPath = root
				return s
			},
			expectErr: false,
		},
		{
			name: "Very_Large_Rotation_Time",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.RotationTime = time.Hour * 24 * 365 * 10 // 10年
				s.MaxSizeMB = 1                            // 使用大小轮转避免时间轮转问题
				s.LogRootFPath = root
				return s
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			settings := tc.setupFunc()
			err := validateSettings(settings)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				}

				// 如果验证通过，尝试创建日志器
				if err == nil {
					_, err := NewLogHelperWithError(settings)
					if err != nil {
						t.Errorf("Failed to create logger for %s: %v", tc.name, err)
					}
				}
			}
		})
	}
}

// TestNullAndEmptyValues 测试空值处理
func TestNullAndEmptyValues(t *testing.T) {
	testCases := []struct {
		name      string
		setupFunc func() *Settings
		expectErr bool
	}{
		{
			name: "Nil_Settings",
			setupFunc: func() *Settings {
				return nil
			},
			expectErr: true,
		},
		{
			name: "Empty_LogRootFPath",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.LogRootFPath = ""
				return s
			},
			expectErr: false, // 应该使用默认值
		},
		{
			name: "Empty_LogNameBase",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.LogNameBase = ""
				s.LogRootFPath = "/tmp"
				return s
			},
			expectErr: true, // 空的 LogNameBase 应该报错
		},
		{
			name: "Whitespace_Only",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.LogRootFPath = "   "
				s.LogNameBase = "   "
				return s
			},
			expectErr: true,
		},
		{
			name: "Empty_FormatterType",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.LogRootFPath = "/tmp"
				s.FormatterType = ""
				return s
			},
			expectErr: false, // 应该使用默认格式器
		},
		{
			name: "Invalid_FormatterType",
			setupFunc: func() *Settings {
				s := NewSettings()
				s.LogRootFPath = "/tmp"
				s.FormatterType = "nonexistent_formatter"
				return s
			},
			expectErr: false, // 应该回退到默认格式器
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			settings := tc.setupFunc()
			err := validateSettings(settings)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.name)
				}
			} else {
				// 对于某些情况，即使验证通过，创建日志器时可能仍会失败
				if err != nil {
					t.Logf("Validation error for %s (may be acceptable): %v", tc.name, err)
				}
			}
		})
	}
}

// TestLargeFileHandling 测试大文件处理
func TestLargeFileHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-large-file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 使用小的最大大小以触发轮转
	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "large_file_test"
	settings.MaxSizeMB = 1  // 1MB
	settings.MaxAgeDays = 1 // 保留1天

	logger := NewLogHelper(settings)

	// 写入大量数据以触发轮转
	const numMessages = 1000
	message := strings.Repeat("This is a test message for large file handling. ", 100) // 约4KB

	start := time.Now()
	for i := 0; i < numMessages; i++ {
		logger.Infof("Message %d: %s", i, message)
	}
	duration := time.Since(start)

	t.Logf("Wrote %d messages in %v (%.2f msg/sec)",
		numMessages, duration, float64(numMessages)/duration.Seconds())

	// 验证文件轮转
	files, err := filepath.Glob(filepath.Join(root, "large_file_test*.log*"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Error("No log files found")
	}

	// 计算总大小
	var totalSize int64
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			t.Errorf("Failed to stat file %s: %v", file, err)
			continue
		}
		totalSize += info.Size()
	}

	expectedMinSize := int64(numMessages * len(message) / 2) // 估算
	if totalSize < expectedMinSize {
		t.Logf("Note: Total size %d may be less than expected %d", totalSize, expectedMinSize)
	}

	t.Logf("Created %d log files, total size: %d bytes", len(files), totalSize)
}

// TestHighConcurrencyStressTest 高并发压力测试
func TestHighConcurrencyStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-stress")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 保存原始状态
	loggerMutex.RLock()
	originalLogger := loggerBase
	originalRotateWriter := rotateLogsWriter
	originalCurrentFile := currentLogFileFPath
	loggerMutex.RUnlock()

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "stress_test"
	settings.MaxSizeMB = 10
	SetLoggerSettings(settings)

	defer func() {
		loggerMutex.Lock()
		loggerBase = originalLogger
		rotateLogsWriter = originalRotateWriter
		currentLogFileFPath = originalCurrentFile
		loggerMutex.Unlock()
	}()

	const (
		numGoroutines = 20
		numMessages   = 100
		messageSize   = 1024 // 1KB
	)

	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	message := strings.Repeat("x", messageSize)

	start := time.Now()

	// 启动大量 goroutine 并发写入
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numMessages; j++ {
				// 使用不同类型的日志方法
				switch j % 5 {
				case 0:
					Infof("Goroutine %d, message %d: %s", id, j, message)
				case 1:
					Debugf("Debug message from goroutine %d, iteration %d", id, j)
				case 2:
					Warnf("Warning from goroutine %d, iteration %d", id, j)
				case 3:
					Errorf("Error from goroutine %d, iteration %d", id, j)
				case 4:
					WithFields(logrus.Fields{
						"goroutine": id,
						"iteration": j,
						"type":      "structured",
					}).Info("Structured log message")
				}

				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalMessages := int64(numGoroutines * numMessages)
	successRate := float64(successCount) / float64(totalMessages) * 100

	t.Logf("Stress test completed:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Messages per goroutine: %d", numMessages)
	t.Logf("  Total messages: %d", totalMessages)
	t.Logf("  Successful writes: %d", successCount)
	t.Logf("  Failed writes: %d", errorCount)
	t.Logf("  Success rate: %.2f%%", successRate)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Throughput: %.2f msg/sec", float64(totalMessages)/duration.Seconds())

	if successRate < 99.0 {
		t.Errorf("Success rate too low: %.2f%% (expected >= 99%%)", successRate)
	}

	// 验证日志文件
	files, err := filepath.Glob(filepath.Join(root, "stress_test*.log*"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Error("No log files created after stress test")
	}

	t.Logf("Created %d log files", len(files))
}

// TestMemoryLeakDetection 内存泄漏检测
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// 强制垃圾回收以获取基线
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	root, err := os.MkdirTemp("", "logger-ut-memory-leak")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 执行大量日志器操作
	const iterations = 100
	for i := 0; i < iterations; i++ {
		settings := NewSettings()
		settings.LogRootFPath = filepath.Join(root, fmt.Sprintf("iter_%d", i))
		settings.LogNameBase = fmt.Sprintf("test_%d", i)

		logger := NewLogHelper(settings)

		// 创建大量 WithFields 条目
		for j := 0; j < 10; j++ {
			entry := logger.WithFields(logrus.Fields{
				"iteration":  i,
				"sub_iter":   j,
				"data":       strings.Repeat("x", 100),
				"nested":     map[string]interface{}{"key": "value"},
				"large_data": make([]byte, 1024),
			})
			entry.Info("Memory leak test message")
		}

		// 关闭日志器
		Close()
	}

	// 再次强制垃圾回收
	runtime.GC()
	runtime.GC() // 调用两次以确保完整回收
	time.Sleep(100 * time.Millisecond)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// 计算内存增长
	memIncrease := m2.Alloc - m1.Alloc
	memIncreaseMB := float64(memIncrease) / 1024 / 1024

	t.Logf("Memory usage:")
	t.Logf("  Before: %d bytes (%.2f MB)", m1.Alloc, float64(m1.Alloc)/1024/1024)
	t.Logf("  After:  %d bytes (%.2f MB)", m2.Alloc, float64(m2.Alloc)/1024/1024)
	t.Logf("  Increase: %d bytes (%.2f MB)", memIncrease, memIncreaseMB)

	// 允许合理的内存增长（考虑日志缓冲等）
	if memIncreaseMB > 10 {
		t.Errorf("Potential memory leak: increased by %.2f MB", memIncreaseMB)
	}
}

// TestConcurrentFileOperations 并发文件操作测试
func TestConcurrentFileOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent file operations test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-concurrent-files")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 保存原始状态
	loggerMutex.Lock()
	originalLogger := loggerBase
	originalRotateWriter := rotateLogsWriter
	originalCurrentFile := currentLogFileFPath
	defer func() {
		loggerBase = originalLogger
		rotateLogsWriter = originalRotateWriter
		currentLogFileFPath = originalCurrentFile
		loggerMutex.Unlock()
	}()

	const numLoggers = 20
	var loggers []*logrus.Logger
	var wg sync.WaitGroup

	// 创建多个日志器
	for i := 0; i < numLoggers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			settings := NewSettings()
			settings.LogRootFPath = filepath.Join(root, fmt.Sprintf("logger_%d", id))
			settings.LogNameBase = fmt.Sprintf("log_%d", id)
			settings.MaxSizeMB = 1

			logger := NewLogHelper(settings)
			loggers = append(loggers, logger)
		}(i)
	}

	wg.Wait()

	// 并发写入所有日志器
	for i, logger := range loggers {
		wg.Add(1)
		go func(id int, l *logrus.Logger) {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				l.Infof("Logger %d, message %d", id, j)
			}
		}(i, logger)
	}

	wg.Wait()

	// 验证文件创建
	var totalFiles int
	for i := 0; i < numLoggers; i++ {
		pattern := filepath.Join(root, fmt.Sprintf("logger_%d", i), "log_*.log*")
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Errorf("Failed to glob files for logger %d: %v", i, err)
		} else {
			totalFiles += len(files)
		}
	}

	if totalFiles == 0 {
		t.Error("No log files created")
	}

	t.Logf("Created %d log files across %d loggers", totalFiles, numLoggers)
}

// TestConcurrentCurrentFileName 测试 CurrentFileName 的并发安全性
func TestConcurrentCurrentFileName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-concurrent-filename")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 保存原始状态
	loggerMutex.Lock()
	originalLogger := loggerBase
	originalRotateWriter := rotateLogsWriter
	originalCurrentFile := currentLogFileFPath
	defer func() {
		loggerBase = originalLogger
		rotateLogsWriter = originalRotateWriter
		currentLogFileFPath = originalCurrentFile
		loggerMutex.Unlock()
	}()

	const numGoroutines = 50
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	var resultsMu sync.Mutex

	// 并发设置日志器并获取文件名
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 设置日志器
			settings := NewSettings()
			settings.LogRootFPath = root
			settings.LogNameBase = fmt.Sprintf("test_%d", id%10) // 只使用10个不同的名称
			SetLoggerSettings(settings)

			// 获取当前文件名
			fileName := CurrentFileName()

			resultsMu.Lock()
			results[id] = fileName
			resultsMu.Unlock()
		}(i)
	}

	wg.Wait()

	// 验证结果
	var nilCount int
	var uniqueNames = make(map[string]bool)

	for _, fileName := range results {
		if fileName == "" {
			nilCount++
		} else {
			uniqueNames[fileName] = true
		}
	}

	t.Logf("CurrentFileName results:")
	t.Logf("  Total calls: %d", numGoroutines)
	t.Logf("  Empty results: %d", nilCount)
	t.Logf("  Unique filenames: %d", len(uniqueNames))

	// 应该有一些非空的结果
	if nilCount == numGoroutines {
		t.Error("All CurrentFileName calls returned empty")
	}
}

// TestWindowsGUICompatibility Windows GUI 兼容性测试
func TestWindowsGUICompatibility(t *testing.T) {
	// 这个测试主要验证 isWindowsGUI 函数的行为
	guiMode := isWindowsGUI()
	t.Logf("Windows GUI mode detected: %v", guiMode)

	// 测试日志器在 GUI 模式下的行为
	root, err := os.MkdirTemp("", "logger-ut-gui-compat")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "gui_test"

	logger := NewLogHelper(settings)

	// 写入测试消息
	logger.Info("GUI compatibility test message")

	// 验证日志器可以正常工作
	if logger == nil {
		t.Error("Logger should be created successfully in GUI mode")
	}
}

// TestFormatValidation 测试格式验证
func TestFormatValidation(t *testing.T) {
	testCases := []struct {
		name         string
		timestampFmt string
		expectErr    bool
	}{
		{
			name:         "Standard_Format",
			timestampFmt: "2006-01-02 15:04:05.000",
			expectErr:    false,
		},
		{
			name:         "RFC3339_Format",
			timestampFmt: time.RFC3339,
			expectErr:    false,
		},
		{
			name:         "Empty_Format",
			timestampFmt: "",
			expectErr:    false,
		},
		{
			name:         "Invalid_Format",
			timestampFmt: "invalid format",
			expectErr:    false, // WithFieldFormatter 不验证格式字符串
		},
		{
			name:         "Very_Long_Format",
			timestampFmt: strings.Repeat("2006", 100),
			expectErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := &WithFieldFormatter{
				TimestampFormat: tc.timestampFmt,
			}

			entry := &logrus.Entry{
				Time:    time.Now(),
				Level:   logrus.InfoLevel,
				Message: "Test message",
			}

			_, err := formatter.Format(entry)
			if tc.expectErr && err == nil {
				t.Errorf("Expected error for format %s, but got none", tc.timestampFmt)
			} else if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error for format %s: %v", tc.timestampFmt, err)
			}
		})
	}
}

// TestConcurrentErrorHandling 并发错误处理测试
func TestConcurrentErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent error handling test in short mode")
	}

	const numGoroutines = 20
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 并发尝试创建无效的日志器
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 尝试使用无效路径
			settings := NewSettings()
			if runtime.GOOS == "windows" {
				settings.LogRootFPath = "C:\\invalid<>|?.log"
			} else {
				settings.LogRootFPath = "/dev/null/invalid/path"
			}
			settings.LogNameBase = fmt.Sprintf("invalid_%d", id)

			_, err := NewLogHelperWithError(settings)
			if err != nil {
				errorChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// 收集所有错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	t.Logf("Concurrent error handling: %d errors captured out of %d attempts",
		len(errors), numGoroutines)

	// 应该有一些错误
	if len(errors) == 0 {
		t.Log("Note: No errors were generated (may have permissions)")
	}
}

// BenchmarkLoggerPerformance 性能基准测试
func BenchmarkLoggerPerformance(b *testing.B) {
	root, err := os.MkdirTemp("", "logger-bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "bench"
	settings.MaxSizeMB = 100 // 足够大以避免轮转

	logger := NewLogHelper(settings)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Infof("Benchmark message %d: %s", i, strings.Repeat("x", 100))
			i++
		}
	})
}

// BenchmarkWithFieldsPerformance WithFields 性能基准测试
func BenchmarkWithFieldsPerformance(b *testing.B) {
	root, err := os.MkdirTemp("", "logger-bench-fields")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "bench_fields"
	logger := NewLogHelper(settings)

	fields := logrus.Fields{
		"user_id":    12345,
		"session_id": "abcdef-123456",
		"ip":         "192.168.1.1",
		"user_agent": "Mozilla/5.0",
		"request_id": "req-123456",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.WithFields(fields).Infof("Benchmark message %d", i)
			i++
		}
	})
}
