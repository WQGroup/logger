package logger

import (
	"testing"
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