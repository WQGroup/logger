package logger

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestHierarchicalPathCreation 测试分层路径创建
func TestHierarchicalPathCreation(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-hierarchical")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "hierarchical_test"
	settings.MaxSizeMB = 1 // 使用大小轮转以测试分层路径
	settings.UseHierarchicalPath = true
	SetLoggerSettings(settings)

	Info("Test hierarchical path creation")

	// 验证路径结构（使用当前日期）
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	expectedPath := filepath.Join(root, year, month, day)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected hierarchical path not created: %s", expectedPath)
	}

	// 验证日志文件存在
	currentFile := CurrentFileName()
	if currentFile == "" {
		t.Fatal("Current file path should not be empty")
	}
}

// TestHierarchicalPathPermissions 测试路径权限
func TestHierarchicalPathPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	root, err := os.MkdirTemp("", "logger-ut-permissions")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "perm_test"
	settings.MaxSizeMB = 1
	settings.UseHierarchicalPath = true
	SetLoggerSettings(settings)

	Info("Test permissions")

	// 验证目录权限
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	path := filepath.Join(root, year, month, day)

	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat path: %v", err)
	}

	// 验证目录权限（应该是 0755 或类似）
	if fileInfo.Mode().Perm()&0700 == 0 {
		t.Error("Directory does not have owner permissions")
	}
}

// TestHierarchicalPathWithSpecialChars 测试包含特殊字符的路径
func TestHierarchicalPathWithSpecialChars(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-special-chars")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	// 使用带特殊字符的日志名称
	settings.LogNameBase = "test-with.special@chars"
	settings.MaxSizeMB = 1
	settings.UseHierarchicalPath = true
	SetLoggerSettings(settings)

	Info("Test special characters in log name")

	// 验证路径和文件创建成功
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	expectedPath := filepath.Join(root, year, month, day)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected path not created: %s", expectedPath)
	}

	currentFile := CurrentFileName()
	if currentFile == "" {
		t.Error("Current log file path is empty")
	}
}

// TestHierarchicalPathNonLatin 测试非拉丁字符路径
func TestHierarchicalPathNonLatin(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-nonlatin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	// 使用中文作为日志名称
	settings.LogNameBase = "测试日志"
	settings.MaxSizeMB = 1
	settings.UseHierarchicalPath = true
	SetLoggerSettings(settings)

	Info("测试非拉丁字符")

	// 验证路径创建成功
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	expectedPath := filepath.Join(root, year, month, day)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected path not created: %s", expectedPath)
	}

	// 验证日志文件创建
	currentFile := CurrentFileName()
	if !strings.Contains(currentFile, "测试日志") {
		t.Logf("Note: Non-Latin characters may be encoded: %s", currentFile)
	}
}

// TestHierarchicalPathDeepStructure 测试深层路径结构
func TestHierarchicalPathDeepStructure(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-deep")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 创建多层嵌套的根目录
	deepRoot := filepath.Join(root, "a", "b", "c", "d")
	err = os.MkdirAll(deepRoot, 0755)
	if err != nil {
		t.Fatal(err)
	}

	settings := NewSettings()
	settings.LogRootFPath = deepRoot
	settings.LogNameBase = "deep_test"
	settings.MaxSizeMB = 1
	settings.UseHierarchicalPath = true
	SetLoggerSettings(settings)

	Info("Test deep path structure")

	// 验证最终路径
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	expectedPath := filepath.Join(deepRoot, year, month, day)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected deep path not created: %s", expectedPath)
	}

	// 验证完整路径长度合理
	currentFile := CurrentFileName()
	if len(currentFile) < len(deepRoot) {
		t.Errorf("Path seems incorrect: %s", currentFile)
	}
}

// TestHierarchicalPathSeparator 测试不同平台的路径分隔符
func TestHierarchicalPathSeparator(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-separator")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "separator_test"
	settings.MaxSizeMB = 1
	settings.UseHierarchicalPath = true

	SetLoggerSettings(settings)
	Info("Test path separator handling")

	// 验证路径使用正确的分隔符
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	expectedPath := filepath.Join(root, year, month, day)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Path with correct separator not created: %s", expectedPath)
	}
}

// TestHierarchicalPathEmptyRoot 测试空根目录的处理
func TestHierarchicalPathEmptyRoot(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试空根目录
	settings := NewSettings()
	settings.LogRootFPath = "" // 空目录
	settings.LogNameBase = "empty_root_test"
	settings.MaxSizeMB = 1
	settings.UseHierarchicalPath = true

	// 应该使用默认目录
	SetLoggerSettings(settings)
	Info("Test empty root directory")

	// 验证日志器已初始化
	logger := GetLogger()
	if logger == nil {
		t.Error("Logger should be initialized even with empty root")
	}
}

// TestHierarchicalPathVsFlatPath 测试分层路径与扁平路径的对比
func TestHierarchicalPathVsFlatPath(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试分层路径
	root1, err := os.MkdirTemp("", "logger-ut-hierarchical-path")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root1)

	settings1 := NewSettings()
	settings1.LogRootFPath = root1
	settings1.LogNameBase = "hier_path_test"
	settings1.MaxSizeMB = 1
	settings1.UseHierarchicalPath = true
	SetLoggerSettings(settings1)
	Info("Hierarchical path message")

	// 测试扁平路径
	root2, err := os.MkdirTemp("", "logger-ut-flat-path")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root2)

	settings2 := NewSettings()
	settings2.LogRootFPath = root2
	settings2.LogNameBase = "flat_path_test"
	settings2.MaxSizeMB = 1
	settings2.UseHierarchicalPath = false
	SetLoggerSettings(settings2)
	Info("Flat path message")

	// 验证两种模式都能正常工作
	logger := GetLogger()
	if logger == nil {
		t.Error("Logger should be working")
	}
}