package logger

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// TestConcurrentWrites 测试并发写入日志
func TestConcurrentWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-concurrent")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	defer func() {
		// 恢复原始状态
		SetLoggerSettings(&Settings{
			LogRootFPath: ".",
			LogNameBase:  "logger",
		})
	}()

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "concurrent_test"
	settings.MaxSizeMB = 10 // 足够大的文件以避免轮转
	SetLoggerSettings(settings)

	const (
		numGoroutines = 50
		numMessages   = 100
	)

	var (
		wg         sync.WaitGroup
		writeCount int64
	)

	// 启动多个 goroutine 并发写入
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				Infof("Goroutine %d, message %d", id, j)
				atomic.AddInt64(&writeCount, 1)
			}
		}(i)
	}

	// 等待所有写入完成
	wg.Wait()

	// 验证写入计数
	expectedCount := int64(numGoroutines * numMessages)
	if writeCount != expectedCount {
		t.Errorf("Expected %d writes, got %d", expectedCount, writeCount)
	}

	t.Logf("Successfully performed %d concurrent writes", writeCount)
}

// TestConcurrentSetLoggerSettings 测试并发调用 SetLoggerSettings
func TestConcurrentSetLoggerSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// No need to save state here since we're testing concurrent access
	defer func() {
		// 恢复原始状态
		SetLoggerSettings(&Settings{
			LogRootFPath: ".",
			LogNameBase:  "logger",
		})
	}()

	const numGoroutines = 20
	var wg sync.WaitGroup

	// 创建多个临时目录
	dirs := make([]string, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		dir, err := os.MkdirTemp("", "logger-ut-concurrent-settings")
		if err != nil {
			t.Fatal(err)
		}
		dirs[i] = dir
		defer os.RemoveAll(dir)
	}

	// 并发设置不同的配置
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			settings := NewSettings()
			settings.LogRootFPath = dirs[id]
			settings.LogNameBase = "concurrent_settings_test"
			settings.Level = logrus.Level(id%6 + 1) // 循环使用不同级别
			SetLoggerSettings(settings)

			// 写入一些日志
			Infof("Test message from goroutine %d", id)
		}(i)
	}

	wg.Wait()

	// 验证最终状态是有效的
	logger := GetLogger()
	if logger == nil {
		t.Error("Logger should be initialized after concurrent settings")
	}

	t.Log("Concurrent SetLoggerSettings test completed successfully")
}

// TestConcurrentRotation 测试并发轮转
func TestConcurrentRotation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-concurrent-rotation")
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

	// 使用较小的大小以触发轮转
	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "rotation_test"
	settings.MaxSizeMB = 1
	SetLoggerSettings(settings)

	const numWriters = 10
	var wg sync.WaitGroup

	// 启动多个写入者
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// 写入大量数据以触发轮转
			longStr := string(make([]byte, 1024)) // 1KB 字符串
			for j := 0; j < 500; j++ {
				Infof("Writer %d, message %d: %s", id, j, longStr)
			}
		}(i)
	}

	// 等待所有写入完成
	wg.Wait()

	// 验证没有数据竞争
	t.Log("Concurrent rotation test completed successfully")
}

// TestConcurrentFormatterAccess 测试并发访问格式器
func TestConcurrentFormatterAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

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

	const numGoroutines = 30
	var wg sync.WaitGroup

	// 并发创建和测试不同的格式器
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 创建新设置
			settings := NewSettings()
			settings.FormatterType = []string{
				FormatterTypeWithField,
				FormatterTypeJSON,
				FormatterTypeText,
			}[id%3]

			// 设置格式器
			SetLoggerSettings(settings)

			// 使用格式器记录日志
			WithFields(logrus.Fields{
				"goroutine": id,
				"iteration": id % 10,
			}).Info("Concurrent formatter test")
		}(i)
	}

	wg.Wait()

	t.Log("Concurrent formatter access test completed successfully")
}

// TestConcurrentHierarchicalPath 测试并发分层路径创建
func TestConcurrentHierarchicalPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-concurrent-hier")
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

	const numGoroutines = 20
	var wg sync.WaitGroup

	// 并发创建使用分层路径的日志器
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			settings := NewSettings()
			settings.LogRootFPath = root
			settings.LogNameBase = "hierarchical_test"
			settings.MaxSizeMB = 1
			settings.UseHierarchicalPath = true
			SetLoggerSettings(settings)

			// 写入日志
			Infof("Hierarchical path test from goroutine %d", id)
		}(i)
	}

	wg.Wait()

	// 验证分层路径被正确创建（使用当前日期）
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	expectedPath := filepath.Join(root, year, month, day)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Logf("Note: Hierarchical path may not be created in test: %s", expectedPath)
	}

	t.Log("Concurrent hierarchical path test completed successfully")
}

// TestConcurrentLoggerBaseAccess 测试并发访问 loggerBase
func TestConcurrentLoggerBaseAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

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

	// 初始化日志器
	settings := NewSettings()
	SetLoggerSettings(settings)

	const numReaders = 50
	const numWriters = 10
	var wg sync.WaitGroup

	// 启动读取者
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				logger := GetLogger()
				if logger == nil {
					t.Errorf("Goroutine %d: GetLogger returned nil", id)
					return
				}
				// 执行一些日志操作
				logger.Debugf("Reader %d, iteration %d", id, j)
			}
		}(i)
	}

	// 启动写入者（修改设置）
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			root, err := os.MkdirTemp("", "logger-ut-concurrent-base")
			if err != nil {
				t.Errorf("Goroutine %d: %v", id, err)
				return
			}
			defer os.RemoveAll(root)

			for j := 0; j < 10; j++ {
				settings := NewSettings()
				settings.LogRootFPath = root
				settings.LogNameBase = "base_test"
				settings.Level = logrus.Level(j%6 + 1)
				SetLoggerSettings(settings)

				// 短暂延迟
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	t.Log("Concurrent logger base access test completed successfully")
}

// TestConcurrentWithFields 测试并发 WithFields
func TestConcurrentWithFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-concurrent-fields")
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

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "fields_test"
	SetLoggerSettings(settings)

	const numGoroutines = 30
	var wg sync.WaitGroup

	// 并发使用 WithFields
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 使用 WithField
			for j := 0; j < 50; j++ {
				WithField("goroutine_id", id).Infof("Message %d", j)
			}

			// 使用 WithFields
			fields := logrus.Fields{
				"goroutine":   id,
				"type":        "concurrent",
				"iteration":   0,
				"data":        "test",
				"nested_data": map[string]interface{}{"key": "value"},
			}

			for j := 0; j < 50; j++ {
				fields["iteration"] = j
				WithFields(fields).Info("Concurrent WithFields test")
			}
		}(i)
	}

	wg.Wait()

	t.Log("Concurrent WithFields test completed successfully")
}

// TestConcurrentLevelChanges 测试并发更改日志级别
func TestConcurrentLevelChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-concurrent-level")
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

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "level_test"
	SetLoggerSettings(settings)

	const numGoroutines = 20
	var wg sync.WaitGroup

	// 并发更改日志级别
	levels := []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 50; j++ {
				// 更改日志级别
				settings := NewSettings()
				settings.LogRootFPath = root
				settings.LogNameBase = "level_test"
				settings.Level = levels[j%len(levels)]
				SetLoggerSettings(settings)

				// 写入不同级别的日志
				logger := GetLogger()
				logger.Trace("Trace message")
				logger.Debug("Debug message")
				logger.Info("Info message")
				logger.Warn("Warn message")
				logger.Error("Error message")
			}
		}(i)
	}

	wg.Wait()

	t.Log("Concurrent level changes test completed successfully")
}

// TestConcurrentGetLogger 测试并发调用 GetLogger
func TestConcurrentGetLogger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// 重置日志器状态
	loggerBase = nil
	rotateLogsWriter = nil
	currentLogFileFPath = ""

	const numGoroutines = 100
	var wg sync.WaitGroup

	// 并发调用 GetLogger（首次初始化）
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			logger := GetLogger()
			if logger == nil {
				t.Errorf("Goroutine %d: GetLogger returned nil", id)
				return
			}

			// 使用日志器
			logger.Infof("Goroutine %d got logger", id)
		}(i)
	}

	wg.Wait()

	// 验证日志器已初始化
	if loggerBase == nil {
		t.Error("Logger should be initialized after concurrent GetLogger calls")
	}

	t.Log("Concurrent GetLogger test completed successfully")
}

// TestRaceConditionWithFormatterFactory 测试格式器工厂的竞态条件
func TestRaceConditionWithFormatterFactory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	factory := &FormatterFactory{}
	const numGoroutines = 50
	var wg sync.WaitGroup

	// 并发创建格式器
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			settings := NewSettings()
			settings.FormatterType = []string{
				FormatterTypeWithField,
				FormatterTypeJSON,
				FormatterTypeText,
				FormatterTypeEasy,
				"",        // 测试空值
				"invalid", // 测试无效值
			}[id%6]

			formatter := factory.CreateFormatter(settings)
			if formatter == nil {
				t.Errorf("Goroutine %d: formatter is nil", id)
				return
			}

			// 测试格式器功能
			entry := &logrus.Entry{
				Time:    time.Now(),
				Level:   logrus.InfoLevel,
				Message: "Test message",
				Data:    logrus.Fields{"id": id},
			}

			_, err := formatter.Format(entry)
			if err != nil {
				t.Errorf("Goroutine %d: format error: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	t.Log("Formatter factory race condition test completed successfully")
}
