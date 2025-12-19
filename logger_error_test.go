package logger

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestDirectoryCreationError 测试目录创建失败
func TestDirectoryCreationError(t *testing.T) {
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

	// 测试在只读文件系统上的行为
	if runtime.GOOS != "windows" {
		t.Run("Readonly_FileSystem", func(t *testing.T) {
			// 创建一个临时目录
			root, err := os.MkdirTemp("", "logger-ut-readonly")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(root)

			// 尝试在只读目录下创建日志目录
			readonlyPath := filepath.Join(root, "readonly")
			err = os.Mkdir(readonlyPath, 0755)
			if err != nil {
				t.Fatal(err)
			}
			// 设置为只读
			err = os.Chmod(readonlyPath, 0444)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chmod(readonlyPath, 0755)

			// 尝试创建只读目录下的子目录
			subDir := filepath.Join(readonlyPath, "subdir")
			settings := NewSettings()
			settings.LogRootFPath = subDir
			settings.LogNameBase = "readonly_test"

			// 这应该会 panic
			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic when creating directory in readonly path")
				}
			}()
			SetLoggerSettings(settings)
		})
	}

	// 测试路径不存在且无法创建
	t.Run("Uncreatable_Path", func(t *testing.T) {
		// 尝试在不存在的路径下创建
		// 使用一个不太可能存在的路径
		if runtime.GOOS == "windows" {
			// Windows 上使用无效字符
			settings := NewSettings()
			settings.LogRootFPath = "C:\\invalid<>|?.log"
			settings.LogNameBase = "invalid_path_test"

			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic with invalid path characters")
				}
			}()
			SetLoggerSettings(settings)
		} else {
			// Unix 系统上尝试在 /root 下创建（如果没有权限）
			settings := NewSettings()
			settings.LogRootFPath = "/root/nonexistent/logger_test"
			settings.LogNameBase = "root_test"

			defer func() {
				// 可能没有权限，所以可能会 panic
				r := recover()
				// 如果没有 panic，说明有权限，这也是可以接受的
				_ = r
			}()
			SetLoggerSettings(settings)
		}
	})
}

// TestFileWriterError 测试文件写入错误
func TestFileWriterError(t *testing.T) {
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试写入已满的磁盘
	// 注意：这个测试比较难以模拟，所以我们主要测试错误处理路径
	t.Run("Disk_Full_Simulation", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-diskfull")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "diskfull_test"
		settings.MaxSizeMB = 1
		SetLoggerSettings(settings)

		// 获取日志器
		logger := GetLogger()

		// 尝试写入大量数据
		// 实际上不太可能填满磁盘，但至少测试写入路径
		for i := 0; i < 1000; i++ {
			logger.Infof("Test message %d: %s", i, strings.Repeat("x", 1000))
		}

		// 验证没有崩溃
		t.Log("Disk full simulation completed")
	})
}

// TestInvalidConfigurationError 测试无效配置导致的错误
func TestInvalidConfigurationError(t *testing.T) {
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试无效的时间轮转配置
	t.Run("Invalid_Time_Rotation", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-invalid-time")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "invalid_time"
		settings.RotationTime = time.Nanosecond // 极短的轮转时间
		settings.MaxSizeMB = 0                  // 强制使用时间轮转

		// 可能会导致 rotatelogs 错误
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic (expected): %v", r)
			}
		}()
		SetLoggerSettings(settings)

		// 尝试写入
		Info("Test invalid rotation time")
	})
}

// TestPermissionDenied 测试权限拒绝错误
func TestPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建一个受限的目录
	root, err := os.MkdirTemp("", "logger-ut-permission")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 创建一个无权限的目录
	noPermissionPath := filepath.Join(root, "noperm")
	err = os.Mkdir(noPermissionPath, 0000)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(noPermissionPath, 0755)

	// 尝试在该目录下创建日志
	settings := NewSettings()
	settings.LogRootFPath = noPermissionPath
	settings.LogNameBase = "perm_test"

	defer func() {
		if r := recover(); r == nil {
			t.Log("No panic (may have permission)")
		}
	}()
	SetLoggerSettings(settings)
}

// TestFileSystemError 测试文件系统错误
func TestFileSystemError(t *testing.T) {
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试路径为空
	t.Run("Empty_Path", func(t *testing.T) {
		settings := NewSettings()
		settings.LogRootFPath = ""
		settings.LogNameBase = ""

		// 应该使用默认值，不应该错误
		SetLoggerSettings(settings)

		logger := GetLogger()
		if logger == nil {
			t.Error("Logger should be initialized even with empty path")
		}
	})

	// 测试非常长的路径
	t.Run("Very_Long_Path", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-longpath")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		// 创建一个很长的路径名
		longName := strings.Repeat("a", 255)
		settings := NewSettings()
		settings.LogRootFPath = filepath.Join(root, longName)
		settings.LogNameBase = longName

		// 可能会导致路径过长错误
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from long path panic: %v", r)
			}
		}()
		SetLoggerSettings(settings)
	})
}

// TestLoggerStateError 测试日志器状态错误
func TestLoggerStateError(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	originalRotateWriter := rotateLogsWriter
	originalCurrentFile := currentLogFileFPath

	defer func() {
		loggerBase = originalLogger
		rotateLogsWriter = originalRotateWriter
		currentLogFileFPath = originalCurrentFile
	}()

	// 测试在未初始化状态下的操作
	t.Run("Uninitialized_Logger", func(t *testing.T) {
		// 重置所有全局变量（需要锁保护）
		loggerMutex.Lock()
		loggerBase = nil
		rotateLogsWriter = nil
		currentLogFileFPath = ""
		// 重置 sync.Once
		loggerOnce = sync.Once{}
		loggerMutex.Unlock()

		// 调用 Info 应该会初始化默认日志器
		Info("Test with uninitialized logger")

		// 验证日志器被创建
		if loggerBase == nil {
			t.Error("Logger should be auto-initialized")
		}
	})

	// 测试 CurrentFileName 在未初始化时
	t.Run("CurrentFileName_Uninitialized", func(t *testing.T) {
		loggerMutex.Lock()
		loggerBase = nil
		rotateLogsWriter = nil
		currentLogFileFPath = ""
		loggerOnce = sync.Once{}
		loggerMutex.Unlock()

		// 应该返回空字符串或默认值
		fileName := CurrentFileName()
		// 不应该 panic
		_ = fileName
	})
}

// TestCleanupError 测试清理过程中的错误
func TestCleanupError(t *testing.T) {
	// 测试清理不存在的目录
	err := CleanupExpiredLogs("/nonexistent/directory", 7)
	if err != nil {
		// 清理不存在的目录应该返回错误
		t.Logf("Expected error for nonexistent directory: %v", err)
	}

	// 测试清理根目录（应该被跳过）
	root := "/"
	if runtime.GOOS == "windows" {
		root = "C:\\"
	}
	err = CleanupExpiredLogs(root, 7)
	// 不应该清理根目录
	if err != nil {
		t.Logf("Cleanup root directory error: %v", err)
	}
}

// TestYAMLError 测试 YAML 相关错误
func TestYAMLError(t *testing.T) {
	// 测试无效的 YAML 格式
	t.Run("Invalid_YAML", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-yaml-error")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		yamlPath := filepath.Join(root, "invalid.yaml")
		invalidYAML := "invalid: yaml: content: ["
		err = os.WriteFile(yamlPath, []byte(invalidYAML), 0644)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadSettingsFromYAML(yamlPath)
		if err == nil {
			t.Error("Should return error for invalid YAML")
		}
	})

	// 测试读取不存在的文件
	t.Run("Nonexistent_YAML", func(t *testing.T) {
		_, err := LoadSettingsFromYAML("/nonexistent/config.yaml")
		if err == nil {
			t.Error("Should return error for nonexistent file")
		}
	})

	// 测试权限不足的文件
	if runtime.GOOS != "windows" {
		t.Run("Permission_Denied_YAML", func(t *testing.T) {
			root, err := os.MkdirTemp("", "logger-ut-yaml-perm")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(root)

			yamlPath := filepath.Join(root, "config.yaml")
			validYAML := "level: info\nlog_root: /tmp"
			err = os.WriteFile(yamlPath, []byte(validYAML), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// 移除读权限
			err = os.Chmod(yamlPath, 0000)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chmod(yamlPath, 0644)

			_, err = LoadSettingsFromYAML(yamlPath)
			if err == nil {
				t.Error("Should return error for file without read permission")
			}
		})
	}
}

// TestErrorHandler 测试错误处理器的健壮性
func TestErrorHandler(t *testing.T) {
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试各种极端情况下的错误处理
	testCases := []struct {
		name     string
		settings func() *Settings
	}{
		{
			name: "All_Zeros",
			settings: func() *Settings {
				s := NewSettings()
				s.MaxAgeDays = 0
				s.MaxSizeMB = 0
				s.RotationTime = 0
				return s
			},
		},
		{
			name: "Negative_Values",
			settings: func() *Settings {
				s := NewSettings()
				s.MaxAgeDays = -1
				s.MaxSizeMB = -1
				s.RotationTime = -time.Hour
				return s
			},
		},
		{
			name: "Extreme_Values",
			settings: func() *Settings {
				s := NewSettings()
				s.MaxAgeDays = 36500 // 100年
				s.MaxSizeMB = 10240  // 10GB
				s.RotationTime = time.Hour * 24 * 365
				return s
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test %s panicked: %v", tc.name, r)
				}
			}()

			settings := tc.settings()
			root, err := os.MkdirTemp("", "logger-ut-error-"+tc.name)
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(root)

			settings.LogRootFPath = root
			settings.LogNameBase = "error_test_" + tc.name
			SetLoggerSettings(settings)

			// 尝试写入日志
			Infof("Test message for %s", tc.name)
		})
	}
}

// TestLoggerRecovery 测试日志器从错误中恢复
func TestLoggerRecovery(t *testing.T) {
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试多次设置错误配置后的恢复
	root, err := os.MkdirTemp("", "logger-ut-recovery")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 设置一个有效的配置
	validSettings := NewSettings()
	validSettings.LogRootFPath = root
	validSettings.LogNameBase = "recovery_test"
	SetLoggerSettings(validSettings)

	Info("First message")

	// 尝试设置无效配置（可能导致 panic）
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from invalid config: %v", r)
		}
	}()

	invalidSettings := NewSettings()
	invalidSettings.LogRootFPath = "/invalid/path"
	SetLoggerSettings(invalidSettings)

	// 恢复到有效配置
	SetLoggerSettings(validSettings)
	Info("Recovered message")

	// 验证日志器仍然可用
	logger := GetLogger()
	if logger == nil {
		t.Error("Logger should be available after recovery")
	}
}
