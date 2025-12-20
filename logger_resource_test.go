package logger

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestLumberjackResourceClosure 测试 lumberjack 资源正确关闭
func TestLumberjackResourceClosure(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "logger-resource-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 配置使用大小轮转（会使用 lumberjack）
	settings := NewSettings()
	settings.LogRootFPath = tmpDir
	settings.LogNameBase = "test"
	settings.MaxSizeMB = 1      // 1MB
	settings.MaxAgeDays = 1     // 1天
	settings.UseHierarchicalPath = false

	// 设置日志器
	SetLoggerSettings(settings)

	// 写入一些日志
	Infof("Test message for lumberjack resource test")

	// 关闭日志器
	err = Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// 验证可以重新设置日志器（资源已正确释放）
	settings2 := NewSettings()
	settings2.LogRootFPath = tmpDir
	settings2.LogNameBase = "test2"
	SetLoggerSettings(settings2)

	Infof("Test message after reopening")

	// 清理
	Close()
}

// TestRotatelvsLumberjack 测试轮转模式切换
func TestRotatelvsLumberjack(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "logger-rotation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试大小轮转（使用 lumberjack）
	settings1 := NewSettings()
	settings1.LogRootFPath = tmpDir
	settings1.LogNameBase = "size_rotation"
	settings1.MaxSizeMB = 1
	settings1.MaxAgeDays = 1
	settings1.UseHierarchicalPath = false

	SetLoggerSettings(settings1)
	Infof("Size rotation test")
	Close()

	// 测试时间轮转（使用 rotatelogs）
	settings2 := NewSettings()
	settings2.LogRootFPath = tmpDir
	settings2.LogNameBase = "time_rotation"
	settings2.RotationTime = time.Hour * 24
	settings2.MaxAgeDays = 7
	settings2.UseHierarchicalPath = false

	SetLoggerSettings(settings2)
	Infof("Time rotation test")
	Close()
}

// TestEnhancedPathValidation 测试增强的路径验证功能
func TestEnhancedPathValidation(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "absolute path should be valid",
			path:    "C:\\temp\\logs",
			wantErr: false,
		},
		{
			name:    "relative path should be invalid",
			path:    "./logs",
			wantErr: true,
			errMsg:  "log path must be absolute",
		},
		{
			name:    "path traversal should be invalid",
			path:    "C:\\temp\\..\\windows",
			wantErr: true,
			errMsg:  "path traversal detected",
		},
		{
			name:    "windows system directory should be invalid",
			path:    "C:\\Windows",
			wantErr: true,
			errMsg:  "cannot use system directory",
		},
		{
			name:    "empty path should be valid",
			path:    "",
			wantErr: false,
		},
		{
			name:    "windows system32 should be invalid",
			path:    "C:\\Windows\\System32",
			wantErr: true,
			errMsg:  "cannot use system directory",
		},
		{
			name:    "windows program files should be invalid",
			path:    "C:\\Program Files",
			wantErr: true,
			errMsg:  "cannot use system directory",
		},
		{
			name:    "windows users should be invalid",
			path:    "C:\\Users",
			wantErr: true,
			errMsg:  "cannot use system directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLogPath(tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}