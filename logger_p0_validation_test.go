package logger

import (
	"runtime"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestP0FixesValidation 验证所有P0修复是否正常工作
func TestP0FixesValidation(t *testing.T) {
	// 测试1: 验证lumberjack资源正确关闭
	t.Run("LumberjackResourceClosure", func(t *testing.T) {
		settings := NewSettings()
		settings.MaxSizeMB = 1
		settings.LogNameBase = "p0_test"

		err := SetLoggerSettingsWithError(settings)
		if err != nil {
			t.Fatalf("SetLoggerSettingsWithError failed: %v", err)
		}

		// 写入一些日志
		Info("P0 validation test message")

		// 关闭日志器
		err = Close()
		if err != nil {
			t.Errorf("Close() returned error: %v", err)
		}

		// 验证可以重新设置（资源已释放）
		settings2 := NewSettings()
		err = SetLoggerSettingsWithError(settings2)
		if err != nil {
			t.Errorf("SetLoggerSettingsWithError after close failed: %v", err)
		}
	})

	// 测试2: 验证并发访问安全性
	t.Run("ConcurrentAccess", func(t *testing.T) {
		const numGoroutines = 10
		const numMessages = 10

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()
				for j := 0; j < numMessages; j++ {
					Infof("Goroutine %d, message %d", id, j)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	// 测试3: 验证路径安全性
	t.Run("PathSecurity", func(t *testing.T) {
		testCases := []struct {
			path    string
			isValid bool
		}{
			{".", true},           // 当前目录允许
		{"./", true},         // 相对路径允许
			{"../etc", false},     // 路径遍历不允许
			{"C:\\Windows", false}, // 系统目录不允许
		}

		for _, tc := range testCases {
			err := validateLogPath(tc.path)
			if tc.isValid && err != nil {
				t.Errorf("Path %s should be valid but got error: %v", tc.path, err)
			}
			if !tc.isValid && err == nil {
				t.Errorf("Path %s should be invalid but got no error", tc.path)
			}
		}
	})

	// 测试4: 验证错误处理
	t.Run("ErrorHandling", func(t *testing.T) {
		// 测试无效配置的错误处理
		invalidSettings := NewSettings()
		invalidSettings.MaxSizeMB = -1 // 无效值

		err := SetLoggerSettingsWithError(invalidSettings)
		if err == nil {
			t.Error("Expected error for invalid MaxSizeMB but got none")
		}

		// 验证恢复机制
		validSettings := NewSettings()
		err = SetLoggerSettingsWithError(validSettings)
		if err != nil {
			t.Errorf("Valid settings should not return error: %v", err)
		}
	})
}

// TestGetLoggerErrorHandling 测试 GetLogger() 错误处理修复
func TestGetLoggerErrorHandling(t *testing.T) {
	// 重置状态
	resetState()

	// 测试正常初始化
	logger, err := GetLogger()
	if err != nil {
		t.Errorf("GetLogger() failed: %v", err)
	}
	if logger == nil {
		t.Error("GetLogger() returned nil logger")
	}

	// 测试多次调用应该返回相同的实例
	logger2, err2 := GetLogger()
	if err2 != nil {
		t.Errorf("Second GetLogger() failed: %v", err2)
	}
	if logger != logger2 {
		t.Error("GetLogger() should return the same instance")
	}
}

// TestGetLoggerUnsafeBackwardCompatibility 测试向后兼容性包装函数
func TestGetLoggerUnsafeBackwardCompatibility(t *testing.T) {
	// 重置状态
	resetState()

	// GetLoggerUnsafe 应该总是返回一个非 nil 的 logger
	logger := GetLoggerUnsafe()
	if logger == nil {
		t.Error("GetLoggerUnsafe() should never return nil")
	}
}

// TestConcurrentInitializationRaceCondition 测试并发初始化竞态条件修复
func TestConcurrentInitializationRaceCondition(t *testing.T) {
	// 重置状态
	resetState()

	const numGoroutines = 100
	var wg sync.WaitGroup
	loggers := make([]*logrus.Logger, 0, numGoroutines)
	var mu sync.Mutex

	// 并发调用 GetLogger 进行初始化
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger, err := GetLogger()
			if err != nil {
				t.Errorf("GetLogger failed in goroutine: %v", err)
				return
			}
			if logger == nil {
				t.Error("GetLogger returned nil in goroutine")
				return
			}

			mu.Lock()
			loggers = append(loggers, logger)
			mu.Unlock()
		}()
	}

	wg.Wait()

	// 验证所有 goroutine 获得的是同一个实例
	if len(loggers) != numGoroutines {
		t.Errorf("Expected %d loggers, got %d", numGoroutines, len(loggers))
	}

	firstLogger := loggers[0]
	for i, logger := range loggers {
		if logger != firstLogger {
			t.Errorf("Logger at index %d is different from first logger", i)
		}
	}
}

// TestWindowsGUIDetectionCaching 测试 Windows GUI 检测缓存
func TestWindowsGUIDetectionCaching(t *testing.T) {
	// 测试 Windows GUI 检测缓存
	result1 := isWindowsGUI()
	result2 := isWindowsGUI()

	// 多次调用应该返回相同的结果（缓存）
	if result1 != result2 {
		t.Error("isWindowsGUI() should return consistent results (caching)")
	}

	// 在非 Windows 系统上应该返回 false
	if runtime.GOOS != "windows" && result1 {
		t.Error("isWindowsGUI() should return false on non-Windows systems")
	}

	t.Logf("isWindowsGUI() result: %v (OS: %s)", result1, runtime.GOOS)
}