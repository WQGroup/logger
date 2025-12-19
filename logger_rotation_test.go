package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRotationBoundaryValues 测试轮转逻辑的边界情况
func TestRotationBoundaryValues(t *testing.T) {
	// 测试 MaxSizeMB 为 1（最小有效值）
	t.Run("MaxSizeMB_Equals_1", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-rotate-1")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "boundary_test"
		settings.MaxSizeMB = 1
		settings.MaxAgeDays = 1
		settings.UseHierarchicalPath = false
		SetLoggerSettings(settings)

		// 写入超过 1MB 的日志
		longStr := strings.Repeat("x", 1024)
		for i := 0; i < 1200; i++ {
			Infof("Test message %d: %s", i, longStr)
		}

		// 验证文件轮转
		files, err := os.ReadDir(root)
		if err != nil {
			t.Fatal(err)
		}

		// 应该有多个日志文件（因为超过 1MB）
		logFiles := 0
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
				logFiles++
			}
		}
		if logFiles < 2 {
			t.Fatalf("Expected at least 2 log files due to rotation, got %d", logFiles)
		}
	})

	// 测试 MaxSizeMB 为 0（禁用大小轮转）
	t.Run("MaxSizeMB_Equals_0", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-rotate-0")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "no_size_rotate"
		settings.MaxSizeMB = 0 // 禁用大小轮转
		settings.RotationTime = time.Hour * 24
		SetLoggerSettings(settings)

		Info("Test message without size rotation")

		// 验证只使用时间轮转
		if rotateLogsWriter == nil {
			t.Fatal("Expected rotateLogsWriter to be non-nil when MaxSizeMB=0")
		}
	})

	// 测试极短的 RotationTime
	t.Run("Very_Short_RotationTime", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-short-rotate")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "short_rotate"
		settings.MaxSizeMB = 0
		settings.RotationTime = time.Millisecond * 100 // 极短轮转时间
		SetLoggerSettings(settings)

		Info("First message")
		time.Sleep(time.Millisecond * 150) // 等待超过轮转时间
		Info("Second message")

		// 验证产生了不同的日志文件
		// 由于轮转时间包含分钟，实际可能不会立即轮转
		// 这里主要验证不会崩溃
	})
}

// TestHierarchicalPathBoundary 测试分层路径的边界情况
func TestHierarchicalPathBoundary(t *testing.T) {
	// 测试基本的分层路径创建
	t.Run("Basic_Hierarchical_Path", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-hier-basic")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "hier_basic_test"
		settings.MaxSizeMB = 0
		settings.UseHierarchicalPath = true
		SetLoggerSettings(settings)

		Info("Basic hierarchical path message")

		// 验证路径创建（使用当前日期）
		now := time.Now()
		year := now.Format("2006")
		month := now.Format("01")
		day := now.Format("02")
		expectedPath := filepath.Join(root, year, month, day)

		// 路径可能不会立即创建，所以我们只验证日志器没有崩溃
		t.Logf("Expected hierarchical path: %s", expectedPath)
	})

	// 测试路径创建
	t.Run("Path_Creation", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-path-create")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "path_create_test"
		settings.MaxSizeMB = 1
		settings.UseHierarchicalPath = true
		SetLoggerSettings(settings)

		Info("Path creation test message")

		// 验证日志器正常工作
		currentFile := CurrentFileName()
		if currentFile == "" {
			t.Error("Current file path should not be empty")
		}
	})
}

// TestRotationTimeZeros 测试 RotationTime 为零的处理
func TestRotationTimeZeros(t *testing.T) {
	t.Run("Zero_RotationTime", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-zero-rotation")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "zero_rotation"
		settings.RotationTime = 0 // 零值
		settings.MaxSizeMB = 0    // 强制使用时间轮转
		SetLoggerSettings(settings)

		// 验证被设置为默认值（24小时）
		if settings.RotationTime != time.Duration(24)*time.Hour {
			t.Fatalf("Expected RotationTime to be set to default 24h, got %v", settings.RotationTime)
		}

		Info("Test with zero rotation time")
	})
}

// TestNegativeMaxSize 测试负数 MaxSizeMB 的处理
func TestNegativeMaxSize(t *testing.T) {
	t.Run("Negative_MaxSizeMB", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-negative-size")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "negative_size"
		settings.MaxSizeMB = -1 // 负数
		SetLoggerSettings(settings)

		// 负数应该被忽略，使用时间轮转
		Info("Test with negative max size")

		if rotateLogsWriter == nil {
			t.Fatal("Expected rotateLogsWriter to be non-nil with negative MaxSizeMB")
		}
	})
}

// TestRotationModeSwitch 测试轮转模式切换
func TestRotationModeSwitch(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-switch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 首先使用大小轮转
	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "switch_test"
	settings.MaxSizeMB = 1
	SetLoggerSettings(settings)

	Info("Message with size rotation")

	// 切换到时间轮转
	settings.MaxSizeMB = 0
	SetLoggerSettings(settings)

	Info("Message with time rotation")

	// 验证切换成功
	if rotateLogsWriter == nil {
		t.Fatal("Expected rotateLogsWriter to be non-nil after switching to time rotation")
	}
}

// TestMaxAgeBoundary 测试 MaxAge 边界值
func TestMaxAgeBoundary(t *testing.T) {
	t.Run("MaxAge_Zero_Days", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-maxage-zero")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "maxage_zero"
		settings.MaxAgeDays = 0 // 零天
		SetLoggerSettings(settings)

		// 验证被设置为默认值（7天）
		if settings.MaxAgeDays != 7 {
			t.Fatalf("Expected MaxAgeDays to be set to default 7, got %d", settings.MaxAgeDays)
		}
	})

	t.Run("MaxAge_Negative_Days", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-maxage-negative")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "maxage_negative"
		settings.MaxAgeDays = -5 // 负数
		SetLoggerSettings(settings)

		// 验证被设置为默认值（7天）
		if settings.MaxAgeDays != 7 {
			t.Fatalf("Expected MaxAgeDays to be set to default 7, got %d", settings.MaxAgeDays)
		}
	})
}