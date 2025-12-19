package logger

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

// TestNewSettingsDefaults 测试 NewSettings 的默认值
func TestNewSettingsDefaults(t *testing.T) {
	settings := NewSettings()

	// 验证所有默认值
	if settings.OnlyMsg != false {
		t.Errorf("Expected OnlyMsg to be false, got %v", settings.OnlyMsg)
	}
	if settings.Level != logrus.InfoLevel {
		t.Errorf("Expected Level to be InfoLevel, got %v", settings.Level)
	}
	if settings.LogRootFPath != logRootFPathDef {
		t.Errorf("Expected LogRootFPath to be %s, got %s", logRootFPathDef, settings.LogRootFPath)
	}
	if settings.LogNameBase != NameDef {
		t.Errorf("Expected LogNameBase to be %s, got %s", NameDef, settings.LogNameBase)
	}
	if settings.RotationTime != time.Duration(24)*time.Hour {
		t.Errorf("Expected RotationTime to be 24h, got %v", settings.RotationTime)
	}
	if settings.MaxAge != time.Duration(7*24)*time.Hour {
		t.Errorf("Expected MaxAge to be 7*24h, got %v", settings.MaxAge)
	}
	if settings.MaxAgeDays != 7 {
		t.Errorf("Expected MaxAgeDays to be 7, got %d", settings.MaxAgeDays)
	}
	if settings.MaxSizeMB != 0 {
		t.Errorf("Expected MaxSizeMB to be 0, got %d", settings.MaxSizeMB)
	}
	if settings.UseHierarchicalPath != false {
		t.Errorf("Expected UseHierarchicalPath to be false, got %v", settings.UseHierarchicalPath)
	}

	// 验证格式器相关默认值
	if settings.FormatterType != FormatterTypeWithField {
		t.Errorf("Expected FormatterType to be %s, got %s", FormatterTypeWithField, settings.FormatterType)
	}
	if settings.TimestampFormat != "2006-01-02 15:04:05.000" {
		t.Errorf("Expected TimestampFormat to be '2006-01-02 15:04:05.000', got %s", settings.TimestampFormat)
	}
	if settings.CustomFormatter != nil {
		t.Error("Expected CustomFormatter to be nil")
	}
	if settings.DisableTimestamp != false {
		t.Errorf("Expected DisableTimestamp to be false, got %v", settings.DisableTimestamp)
	}
	if settings.DisableLevel != false {
		t.Errorf("Expected DisableLevel to be false, got %v", settings.DisableLevel)
	}
	if settings.DisableCaller != true {
		t.Errorf("Expected DisableCaller to be true, got %v", settings.DisableCaller)
	}
}

// TestSetLoggerSettingsValidation 测试 SetLoggerSettings 的参数验证
func TestSetLoggerSettingsValidation(t *testing.T) {
	// 保存原始状态
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试空路径
	t.Run("Empty_LogRootFPath", func(t *testing.T) {
		settings := NewSettings()
		settings.LogRootFPath = "" // 空路径
		SetLoggerSettings(settings)

		// 应该被设置为默认值
		logger := GetLogger()
		if logger == nil {
			t.Error("Logger should not be nil")
		}
	})

	// 测试空日志名称
	t.Run("Empty_LogNameBase", func(t *testing.T) {
		settings := NewSettings()
		settings.LogNameBase = "" // 空名称
		SetLoggerSettings(settings)

		// 应该被设置为默认值
		logger := GetLogger()
		if logger == nil {
			t.Error("Logger should not be nil")
		}
	})

	// 测试负数轮转时间
	t.Run("Negative_RotationTime", func(t *testing.T) {
		settings := NewSettings()
		settings.RotationTime = -time.Hour
		settings.MaxSizeMB = 0 // 强制使用时间轮转
		SetLoggerSettings(settings)

		// 应该被设置为默认值
		if settings.RotationTime != time.Duration(24)*time.Hour {
			t.Errorf("Expected RotationTime to be set to default, got %v", settings.RotationTime)
		}
	})

	// 测试零轮转时间
	t.Run("Zero_RotationTime", func(t *testing.T) {
		settings := NewSettings()
		settings.RotationTime = 0
		settings.MaxSizeMB = 0 // 强制使用时间轮转
		SetLoggerSettings(settings)

		// 应该被设置为默认值
		if settings.RotationTime != time.Duration(24)*time.Hour {
			t.Errorf("Expected RotationTime to be set to default, got %v", settings.RotationTime)
		}
	})

	// 测试无效的日志级别
	t.Run("Invalid_Level_Should_Not_Panic", func(t *testing.T) {
		settings := NewSettings()
		// logrus.Level 是有效的枚举，所以无法设置无效值
		// 但我们可以测试极值
		settings.Level = logrus.Level(999) // 设置一个极高的值
		SetLoggerSettings(settings)

		// 不应该 panic
		logger := GetLogger()
		if logger == nil {
			t.Error("Logger should not be nil")
		}
	})
}

// TestYAMLConfigValidation 测试 YAML 配置验证
func TestYAMLConfigValidation(t *testing.T) {
	// 测试无效的日志级别
	t.Run("Invalid_Level", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-yaml-invalid")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		yamlPath := createTestYAML(t, root, map[string]interface{}{
			"level": "invalid_level",
		})

		settings, err := LoadSettingsFromYAML(yamlPath)
		if err != nil {
			t.Fatalf("Failed to load YAML: %v", err)
		}

		// 无效级别应该被设置为默认值 InfoLevel
		if settings.Level != logrus.InfoLevel {
			t.Errorf("Expected Level to be InfoLevel for invalid input, got %v", settings.Level)
		}
	})

	// 测试空的配置文件
	t.Run("Empty_Config", func(t *testing.T) {
		root, err := os.MkdirTemp("", "logger-ut-yaml-empty")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(root)

		yamlPath := createTestYAML(t, root, map[string]interface{}{})

		settings, err := LoadSettingsFromYAML(yamlPath)
		if err != nil {
			t.Fatalf("Failed to load empty YAML: %v", err)
		}

		// 应该使用所有默认值
		defaults := NewSettings()
		if settings.Level != defaults.Level {
			t.Error("Empty config should use default Level")
		}
		if settings.FormatterType != defaults.FormatterType {
			t.Error("Empty config should use default FormatterType")
		}
	})

	// 测试不存在的文件
	t.Run("Non_Existent_File", func(t *testing.T) {
		_, err := LoadSettingsFromYAML("/non/existent/file.yaml")
		if err == nil {
			t.Error("Should return error for non-existent file")
		}
	})

	// 测试大小写不敏感的日志级别
	t.Run("Case_Insensitive_Level", func(t *testing.T) {
		levels := []string{"DEBUG", "Info", "WARN", "error", "FATAL", "panic"}

		for _, levelStr := range levels {
			root, err := os.MkdirTemp("", "logger-ut-yaml-case")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(root)

			yamlPath := createTestYAML(t, root, map[string]interface{}{
				"level": levelStr,
			})

			settings, err := LoadSettingsFromYAML(yamlPath)
			if err != nil {
				t.Fatalf("Failed to load YAML with level %s: %v", levelStr, err)
			}

			// 验证级别被正确解析
			expectedLevel := parseLevel(levelStr)
			if settings.Level != expectedLevel {
				t.Errorf("Level mismatch for %s: expected %v, got %v", levelStr, expectedLevel, settings.Level)
			}
		}
	})
}

// TestFormatterTypeValidation 测试格式器类型验证
func TestFormatterTypeValidation(t *testing.T) {
	factory := &FormatterFactory{}
	settings := NewSettings()

	// 测试所有有效的格式器类型
	validTypes := []string{
		FormatterTypeWithField,
		FormatterTypeJSON,
		FormatterTypeText,
		FormatterTypeEasy,
	}

	for _, ftype := range validTypes {
		settings.FormatterType = ftype
		formatter := factory.CreateFormatter(settings)
		if formatter == nil {
			t.Errorf("Formatter should not be nil for type: %s", ftype)
		}
	}

	// 测试空格式器类型（应该使用默认）
	settings.FormatterType = ""
	formatter := factory.CreateFormatter(settings)
	if _, ok := formatter.(*WithFieldFormatter); !ok {
		t.Error("Empty formatter type should use WithFieldFormatter as default")
	}

	// 测试无效格式器类型（应该使用默认）
	settings.FormatterType = "invalid_type"
	formatter = factory.CreateFormatter(settings)
	if _, ok := formatter.(*WithFieldFormatter); !ok {
		t.Error("Invalid formatter type should use WithFieldFormatter as fallback")
	}
}

// TestBackwardCompatibilityOnlyMsg 测试 OnlyMsg 向后兼容性
func TestBackwardCompatibilityOnlyMsg(t *testing.T) {
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试 OnlyMsg=true 时的行为
	settings := NewSettings()
	settings.OnlyMsg = true
	settings.FormatterType = "" // 不设置，验证向后兼容
	SetLoggerSettings(settings)

	logger := GetLogger()
	if _, ok := logger.Formatter.(*easy.Formatter); !ok {
		t.Error("OnlyMsg=true should use easy formatter")
	}

	// 测试 OnlyMsg=true 时同时设置 FormatterType
	settings.OnlyMsg = true
	settings.FormatterType = FormatterTypeJSON
	SetLoggerSettings(settings)

	logger = GetLogger()
	// OnlyMsg 应该优先，使用 easy formatter
	if _, ok := logger.Formatter.(*easy.Formatter); !ok {
		t.Error("OnlyMsg=true should override FormatterType")
	}
}

// TestCustomFormatterValidation 测试自定义格式器
func TestCustomFormatterValidation(t *testing.T) {
	originalLogger := loggerBase
	defer func() {
		loggerBase = originalLogger
	}()

	// 测试设置自定义格式器
	customFormatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02",
	}

	settings := NewSettings()
	settings.CustomFormatter = customFormatter
	SetLoggerSettings(settings)

	logger := GetLogger()
	if logger.Formatter != customFormatter {
		t.Error("Custom formatter should be used when set")
	}

	// 测试通过 SetCustomFormatter 函数
	anotherCustom := &logrus.TextFormatter{}
	SetCustomFormatter(anotherCustom)

	logger = GetLogger()
	if logger.Formatter != anotherCustom {
		t.Error("SetCustomFormatter should update the formatter")
	}
}

// TestConfigFieldInteraction 测试配置字段之间的相互影响
func TestConfigFieldInteraction(t *testing.T) {
	root, err := os.MkdirTemp("", "logger-ut-config-interaction")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// 测试 MaxAgeDays 和 MaxAge 的关系
	settings := NewSettings()
	settings.LogRootFPath = root
	settings.MaxAgeDays = 5
	SetLoggerSettings(settings)

	// MaxAge 应该被相应设置
	expectedMaxAge := time.Duration(5*24) * time.Hour
	if settings.MaxAge != expectedMaxAge {
		t.Errorf("MaxAge should be updated when MaxAgeDays is set: expected %v, got %v", expectedMaxAge, settings.MaxAge)
	}

	// 测试 MaxAgeDays 为 0 时的行为
	settings.MaxAgeDays = 0
	SetLoggerSettings(settings)
	// 应该使用默认值
	if settings.MaxAgeDays != 7 {
		t.Errorf("MaxAgeDays should use default when set to 0: expected 7, got %d", settings.MaxAgeDays)
	}
}

// createTestYAML 辅助函数：创建测试用的 YAML 文件
func createTestYAML(t *testing.T, dir string, config map[string]interface{}) string {
	t.Helper()

	yamlPath := dir + "/test_config.yaml"
	var yamlLines []string

	for key, value := range config {
		yamlLines = append(yamlLines, key+": "+formatYAMLValue(value))
	}

	yamlContent := strings.Join(yamlLines, "\n")
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML: %v", err)
	}

	return yamlPath
}

// formatYAMLValue 辅助函数：格式化 YAML 值
func formatYAMLValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return "'" + v + "'"
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}
