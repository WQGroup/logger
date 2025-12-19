package logger

import (
	"os"
	"runtime"
	"testing"
)

// TestIsWindowsGUI 测试 isWindowsGUI 函数
func TestIsWindowsGUI(t *testing.T) {
	// 这个测试主要验证函数不会 panic
	// 在不同平台上的行为会有所不同
	isGUI := isWindowsGUI()

	// 只在 Windows 上进行更详细的验证
	if runtime.GOOS == "windows" {
		// 在正常情况下，通常不是 GUI 模式
		// 但我们无法保证测试环境的状态，所以只验证函数能正常执行
		t.Logf("Running on Windows, isWindowsGUI() returned: %v", isGUI)
	} else {
		// 非 Windows 系统应该返回 false
		if isGUI {
			t.Errorf("Expected isWindowsGUI() to return false on non-Windows system")
		}
	}
}

// TestWindowsGUIOutput 测试 Windows GUI 模式下的输出
func TestWindowsGUIOutput(t *testing.T) {
	// 保存原始的 loggerBase
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	root, err := os.MkdirTemp("", "logger-ut-windows-gui")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "gui_test"
	SetLoggerSettings(settings)

	logger := GetLogger()

	// 验证 logger 的输出设置
	// 在 GUI 模式下，输出应该只写入文件
	// 在非 GUI 模式下，输出会同时写入文件和 stderr
	// 我们无法直接测试 GUI 模式，但可以验证输出设置
	if logger.Out == nil {
		t.Error("Expected logger.Out to be set")
	}
}

// TestWindowsPathHandling 测试 Windows 路径处理
func TestWindowsPathHandling(t *testing.T) {
	// 测试带盘符的路径
	if runtime.GOOS == "windows" {
		root, err := os.MkdirTemp("", "logger-ut-windows-path")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		// 测试使用带盘符的绝对路径
		settings := NewSettings()
		settings.LogRootFPath = root
		settings.LogNameBase = "path_test"
		settings.UseHierarchicalPath = true
		SetLoggerSettings(settings)

		Info("Test Windows path handling")

		// 验证日志文件被创建
		currentFile := CurrentFileName()
		if currentFile == "" {
			t.Error("Expected current log file path to be set")
		}

		// 验证路径格式正确
		if len(currentFile) < len(root) {
			t.Errorf("Current file path seems incorrect: %s", currentFile)
		}
	}
}

// TestConcurrentAccessWindows 测试 Windows 下的并发访问
func TestConcurrentAccessWindows(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	root, err := os.MkdirTemp("", "logger-ut-concurrent-win")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "concurrent_test"
	settings.MaxSizeMB = 1
	SetLoggerSettings(settings)

	// 并发写入日志
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				Infof("Goroutine %d, message %d", id, j)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证没有数据竞争或崩溃
	t.Log("Concurrent access test completed successfully")
}

// TestStderrReplacement 测试 stderr 替换行为
func TestStderrReplacement(t *testing.T) {
	// 保存原始 stderr
	originalStderr := os.Stderr
	defer func() {
		os.Stderr = originalStderr
	}()

	// 创建一个管道来捕获输出
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	// 使用会写入 stderr 的设置
	settings := NewSettings()
	settings.LogRootFPath = os.TempDir()
	settings.LogNameBase = "stderr_test"
	SetLoggerSettings(settings)

	// 写入日志
	Info("Test stderr output")

	// 关闭写入端以获取所有输出
	w.Close()

	// 在 Windows GUI 模拟环境下，可能不会写入 stderr
	// 所以我们只验证没有 panic
	t.Log("Stderr test completed")

	// 恢复原始 stderr
	os.Stderr = originalStderr
	r.Close()
}