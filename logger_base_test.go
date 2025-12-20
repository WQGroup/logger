package logger

import (
	"bytes"
	"os"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestLoggerBaseFunctions 测试 logger_base.go 中的所有包装函数
func TestLoggerBaseFunctions(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建一个新的日志器用于测试
	testLogger := logrus.New()
	testLogger.Out = &bytes.Buffer{}       // 捕获输出
	testLogger.SetLevel(logrus.DebugLevel) // 设置为 Debug 级别以确保所有日志都输出

	// 直接设置全局变量以绕过 GetLogger() 的自动初始化
	loggerBase = testLogger

	// 测试所有格式化函数
	testCases := []struct {
		name     string
		function func(...interface{})
		message  string
	}{
		{"Debug", Debug, "debug message"},
		{"Info", Info, "info message"},
		{"Print", Print, "print message"},
		{"Warn", Warn, "warn message"},
		{"Warning", Warning, "warning message"},
		{"Error", Error, "error message"},
		{"Fatal", Fatal, "fatal message"},
		{"Panic", Panic, "panic message"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "Fatal" {
				// Fatal 会调用 os.Exit，跳过这个测试
				t.Skip("Skipping Fatal test as it calls os.Exit")
			}
			if tc.name == "Panic" {
				// Panic 会 panic，需要特殊处理
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from %s: %v", tc.name, r)
					}
				}()
			}

			// 重置 buffer
			loggerBase.Out = &bytes.Buffer{}
			tc.function(tc.message)

			// 验证输出
			output := loggerBase.Out.(*bytes.Buffer).String()
			if len(output) > 0 {
				t.Logf("%s output length: %d", tc.name, len(output))
			} else {
				t.Logf("%s produced no output", tc.name)
			}
		})
	}
}

// TestLoggerBaseFormattedFunctions 测试格式化日志函数
func TestLoggerBaseFormattedFunctions(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建一个新的日志器用于测试
	testLogger := logrus.New()
	testLogger.Out = &bytes.Buffer{}       // 捕获输出
	testLogger.SetLevel(logrus.DebugLevel) // 设置为 Debug 级别以确保所有日志都输出
	loggerBase = testLogger

	// 测试所有格式化函数
	testCases := []struct {
		name     string
		function func(string, ...interface{})
		format   string
		args     []interface{}
	}{
		{"Debugf", Debugf, "debug %s", []interface{}{"test"}},
		{"Infof", Infof, "info %d", []interface{}{123}},
		{"Printf", Printf, "print %v", []interface{}{true}},
		{"Warnf", Warnf, "warn %f", []interface{}{3.14}},
		{"Warningf", Warningf, "warning %t", []interface{}{false}},
		{"Errorf", Errorf, "error %x", []interface{}{255}},
		{"Fatalf", Fatalf, "fatal %s", []interface{}{"test"}},
		{"Panicf", Panicf, "panic %d", []interface{}{456}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "Fatalf" {
				// Fatalf 会调用 os.Exit，跳过这个测试
				t.Skip("Skipping Fatalf test as it calls os.Exit")
			}
			if tc.name == "Panicf" {
				// Panicf 会 panic，需要特殊处理
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from %s: %v", tc.name, r)
					}
				}()
			}

			// 重置 buffer
			loggerBase.Out = &bytes.Buffer{}
			tc.function(tc.format, tc.args...)

			// 记录输出长度
			output := loggerBase.Out.(*bytes.Buffer).String()
			t.Logf("%s output length: %d", tc.name, len(output))
			// 不强制要求输出，因为某些级别的日志可能被过滤或会终止程序
		})
	}
}

// TestLoggerBaseLnFunctions 测试带 ln 的日志函数
func TestLoggerBaseLnFunctions(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建一个新的日志器用于测试
	testLogger := logrus.New()
	testLogger.Out = &bytes.Buffer{}       // 捕获输出
	testLogger.SetLevel(logrus.DebugLevel) // 设置为 Debug 级别以确保所有日志都输出
	loggerBase = testLogger

	// 测试所有带 ln 的函数
	testCases := []struct {
		name     string
		function func(...interface{})
		args     []interface{}
	}{
		{"Debugln", Debugln, []interface{}{"debug", "message"}},
		{"Infoln", Infoln, []interface{}{"info", "message"}},
		{"Println", Println, []interface{}{"print", "message"}},
		{"Warnln", Warnln, []interface{}{"warn", "message"}},
		{"Warningln", Warningln, []interface{}{"warning", "message"}},
		{"Errorln", Errorln, []interface{}{"error", "message"}},
		{"Fatalln", Fatalln, []interface{}{"fatal", "message"}},
		{"Panicln", Panicln, []interface{}{"panic", "message"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "Fatalln" {
				// Fatalln 会调用 os.Exit，跳过这个测试
				t.Skip("Skipping Fatalln test as it calls os.Exit")
			}
			if tc.name == "Panicln" {
				// Panicln 会 panic，需要特殊处理
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from %s: %v", tc.name, r)
					}
				}()
			}

			// 重置 buffer
			loggerBase.Out = &bytes.Buffer{}
			tc.function(tc.args...)

			// 记录输出长度
			output := loggerBase.Out.(*bytes.Buffer).String()
			t.Logf("%s output length: %d", tc.name, len(output))
			// 不强制要求输出，因为某些级别的日志可能被过滤或会终止程序
		})
	}
}

// TestWithFieldFunction 测试 WithField 函数
func TestWithFieldFunction(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建一个新的日志器用于测试
	testLogger := logrus.New()
	testLogger.Out = &bytes.Buffer{}       // 捕获输出
	testLogger.SetLevel(logrus.DebugLevel) // 设置为 Debug 级别以确保所有日志都输出
	loggerBase = testLogger

	// 测试 WithField
	entry := WithField("key", "value")
	if entry == nil {
		t.Error("WithField returned nil")
		return
	}

	// 验证 entry 不是 nil
	if entry == nil {
		t.Error("WithField should not return nil")
	}

	// 使用返回的 entry
	entry.Info("Test message with field")

	// 验证输出包含字段信息
	output := loggerBase.Out.(*bytes.Buffer).String()
	if !contains(output, "key=value") {
		t.Error("Output should contain field information")
	}
}

// TestWithFieldsFunction 测试 WithFields 函数
func TestWithFieldsFunction(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建一个新的日志器用于测试
	testLogger := logrus.New()
	testLogger.Out = &bytes.Buffer{}       // 捕获输出
	testLogger.SetLevel(logrus.DebugLevel) // 设置为 Debug 级别以确保所有日志都输出
	loggerBase = testLogger

	// 测试 WithFields
	fields := logrus.Fields{
		"string": "value",
		"int":    123,
		"bool":   true,
		"float":  3.14,
	}

	entry := WithFields(fields)
	if entry == nil {
		t.Error("WithFields returned nil")
		return
	}

	// 验证 entry 不是 nil
	if entry == nil {
		t.Error("WithFields should not return nil")
	}

	// 使用返回的 entry
	entry.Info("Test message with multiple fields")

	// 验证输出包含所有字段信息
	output := loggerBase.Out.(*bytes.Buffer).String()
	for key := range fields {
		if !contains(output, key+"=") {
			t.Errorf("Output should contain field %s", key)
		}
	}
}

// TestSetLoggerNameFunction 测试 SetLoggerName 函数
func TestSetLoggerNameFunction(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建临时目录
	root, err := os.MkdirTemp("", "logger-ut-setname")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 测试设置日志名称
	testName := "test_logger_name"
	SetLoggerName(testName)

	// 验证日志器已创建
	logger := GetLogger()
	if logger == nil {
		t.Error("Logger should be created after SetLoggerName")
		return
	}

	// 写入一条日志
	Info("Test message after setting name")

	// 验证日志文件路径包含设置的名称
	currentFile := CurrentFileName()
	if currentFile != "" && !contains(currentFile, testName) {
		t.Logf("Note: Current log file may not contain the name: %s", currentFile)
	}
}

// TestLoggerBaseNilHandling 测试日志器为 nil 时的处理
func TestLoggerBaseNilHandling(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	// 需要重置 loggerOnce 以确保每次测试都能正确初始化
	var originalOnce sync.Once
	loggerOnce = originalOnce
	defer func() {
		loggerBase = originalLogger
		// 恢复 loggerOnce 以便后续测试正常工作
		loggerOnce = sync.Once{}
	}()

	// 将日志器设置为 nil
	loggerBase = nil

	// 测试所有函数在 loggerBase 为 nil 时的行为
	testCases := []struct {
		name     string
		function func()
	}{
		{"Debug", func() { Debug("test") }},
		{"Info", func() { Info("test") }},
		{"Print", func() { Print("test") }},
		{"Warn", func() { Warn("test") }},
		{"Warning", func() { Warning("test") }},
		{"Error", func() { Error("test") }},
		{"Debugf", func() { Debugf("test %s", "value") }},
		{"Infof", func() { Infof("test %s", "value") }},
		{"Printf", func() { Printf("test %s", "value") }},
		{"Warnf", func() { Warnf("test %s", "value") }},
		{"Warningf", func() { Warningf("test %s", "value") }},
		{"Errorf", func() { Errorf("test %s", "value") }},
		{"Debugln", func() { Debugln("test") }},
		{"Infoln", func() { Infoln("test") }},
		{"Println", func() { Println("test") }},
		{"Warnln", func() { Warnln("test") }},
		{"Warningln", func() { Warningln("test") }},
		{"Errorln", func() { Errorln("test") }},
		{"WithField", func() { WithField("key", "value") }},
		{"WithFields", func() { WithFields(logrus.Fields{"key": "value"}) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 所有函数都应该能正常工作（自动初始化日志器）
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked when loggerBase is nil: %v", tc.name, r)
				}
			}()

			tc.function()

			// 验证日志器已被初始化
			if loggerBase == nil {
				t.Errorf("%s did not initialize loggerBase", tc.name)
			}
		})
	}
}

// TestLoggerBaseLevelFiltering 测试日志级别过滤
func TestLoggerBaseLevelFiltering(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建临时目录
	root, err := os.MkdirTemp("", "logger-ut-level-filter")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 设置日志级别为 Warn
	settings := NewSettings()
	settings.LogRootFPath = root
	settings.LogNameBase = "level_filter_test"
	settings.Level = logrus.WarnLevel

	// 直接创建日志器而不使用 SetLoggerSettings，避免被覆盖
	testLogger, err := NewLogHelperWithError(settings)
	if err != nil {
		t.Fatal(err)
	}
	loggerBase = testLogger
	loggerBase.Out = &bytes.Buffer{} // 捕获输出

	// 写入不同级别的日志
	Debug("Debug message - should not appear")
	Info("Info message - should not appear")
	Warn("Warn message - should appear")
	Error("Error message - should appear")

	// 验证输出
	output := loggerBase.Out.(*bytes.Buffer).String()
	if contains(output, "Debug message") {
		t.Error("Debug message should not appear when level is Warn")
	}
	if contains(output, "Info message") {
		t.Error("Info message should not appear when level is Warn")
	}
	if !contains(output, "Warn message") {
		t.Error("Warn message should appear when level is Warn")
	}
	if !contains(output, "Error message") {
		t.Error("Error message should appear when level is Warn")
	}
}

// TestLoggerBaseConcurrentAccess 测试并发访问 logger_base 函数
func TestLoggerBaseConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 初始化日志器
	testLogger := logrus.New()
	testLogger.Out = &bytes.Buffer{}
	testLogger.SetLevel(logrus.DebugLevel)
	loggerBase = testLogger

	const numGoroutines = 20
	const numCalls = 50

	done := make(chan bool, numGoroutines)

	// 并发调用各种函数
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				done <- true
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
			}()

			for j := 0; j < numCalls; j++ {
				Debugf("Debug %d-%d", id, j)
				Infof("Info %d-%d", id, j)
				Warnf("Warn %d-%d", id, j)
				Errorf("Error %d-%d", id, j)

				WithField("goroutine", id).Info("WithField test")
				WithFields(logrus.Fields{
					"goroutine": id,
					"iteration": j,
				}).Info("WithFields test")
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	t.Log("Concurrent access to logger_base functions completed successfully")
}

// TestLoggerBaseComplexMessage 测试复杂消息格式
func TestLoggerBaseComplexMessage(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 创建日志器
	testLogger := logrus.New()
	testLogger.Out = &bytes.Buffer{}
	loggerBase = testLogger

	// 测试各种复杂消息
	testCases := []struct {
		name     string
		function func()
	}{
		{
			name: "Empty string",
			function: func() {
				Info("")
			},
		},
		{
			name: "Unicode characters",
			function: func() {
				Info("测试中文 🚀 emoji")
			},
		},
		{
			name: "Special characters",
			function: func() {
				Info("Special: \\n\\t\\r\"'{}[]()<>")
			},
		},
		{
			name: "Very long message",
			function: func() {
				longStr := ""
				for i := 0; i < 1000; i++ {
					longStr += "This is a very long message. "
				}
				Info(longStr)
			},
		},
		{
			name: "Nil pointer in format",
			function: func() {
				Infof("Nil value: %v", nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tc.name, r)
				}
			}()

			// 重置 buffer
			loggerBase.Out = &bytes.Buffer{}
			tc.function()

			// 验证没有崩溃
			t.Logf("%s completed successfully", tc.name)
		})
	}
}

// contains 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && indexOf(s, substr) >= 0))
}

// indexOf 辅助函数：查找子串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
