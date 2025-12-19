package logger

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "github.com/sirupsen/logrus"
    easy "github.com/t-tomalak/logrus-easy-formatter"
)

func TestYAMLAndSizeRotationPath(t *testing.T) {
    root, err := os.MkdirTemp("", "logger-ut")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(root)
    yamlPath := filepath.Join(root, "logger.yaml")
    yaml := strings.Join([]string{
        "log_root: '" + filepath.ToSlash(root) + "'",
        "log_name_base: 'ut'",
        "level: 'debug'",
        "days_to_keep: 7",
        "max_size_mb: 1",
        "use_hierarchical_path: true",
        "",
    }, "\n")
    if err := ioutil.WriteFile(yamlPath, []byte(yaml), 0644); err != nil {
        t.Fatal(err)
    }
    if err := SetLoggerFromYAML(yamlPath); err != nil {
        t.Fatal(err)
    }
    for i := 0; i < 5000; i++ {
        Infof("%s", strings.Repeat("x", 300))
    }
    now := time.Now()
    dayDir := filepath.Join(root, now.Format("2006"), now.Format("01"), now.Format("02"))
    if _, err := os.Stat(dayDir); err != nil {
        t.Fatalf("day dir not exists: %s", dayDir)
    }
    expected := filepath.Join(dayDir, "ut.log")
    if LogLinkFileFPath() != expected {
        t.Fatalf("unexpected link path: %s", LogLinkFileFPath())
    }
}

func TestTimeRotationPath(t *testing.T) {
    root, err := os.MkdirTemp("", "logger-ut-time")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(root)
    s := NewSettings()
    s.LogRootFPath = root
    s.LogNameBase = "ut2"
    s.MaxSizeMB = 0
    s.UseHierarchicalPath = true
    SetLoggerSettings(s)
    Info("hello")
    now := time.Now()
    dayDir := filepath.Join(root, now.Format("2006"), now.Format("01"), now.Format("02"))
    cur := CurrentFileName()
    if !strings.HasPrefix(cur, dayDir+string(os.PathSeparator)) {
        t.Fatalf("current file not in day dir: %s", cur)
    }
}

func TestCleanupExpired(t *testing.T) {
    root, err := os.MkdirTemp("", "logger-ut-clean")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(root)
    old := time.Now().AddDate(0, 0, -30)
    y := old.Format("2006")
    m := old.Format("01")
    d := old.Format("02")
    dayDir := filepath.Join(root, y, m, d)
    if err := os.MkdirAll(dayDir, 0755); err != nil {
        t.Fatal(err)
    }
    f := filepath.Join(dayDir, "x.log")
    if err := ioutil.WriteFile(f, []byte("x"), 0644); err != nil {
        t.Fatal(err)
    }
    if err := CleanupExpiredLogs(root, 7); err != nil {
        t.Fatal(err)
    }
    if _, err := os.Stat(dayDir); err == nil {
        t.Fatalf("expired day dir still exists: %s", dayDir)
    }
    monthDir := filepath.Join(root, y, m)
    if _, err := os.Stat(monthDir); err == nil {
        t.Fatalf("empty month dir still exists: %s", monthDir)
    }
    yearDir := filepath.Join(root, y)
    if _, err := os.Stat(yearDir); err == nil {
        t.Fatalf("empty year dir still exists: %s", yearDir)
    }
}

// TestWithFieldFormatter 测试 WithFieldFormatter 格式器
func TestWithFieldFormatter(t *testing.T) {
    formatter := &WithFieldFormatter{
        TimestampFormat:  "2006-01-02 15:04:05.000",
        DisableTimestamp: false,
        DisableLevel:     false,
        DisableCaller:    true,
    }

    // 测试基本格式
    entry := &logrus.Entry{
        Time:    time.Date(2025, 12, 18, 18, 32, 7, 379000000, time.Local),
        Level:   logrus.InfoLevel,
        Message: "【实时通知】事件广播成功",
        Data:    logrus.Fields{
            "operation": "(a+b)-c",
            "result":    123.45,
        },
    }

    output, err := formatter.Format(entry)
    if err != nil {
        t.Fatal(err)
    }

    expected := "2025-12-18 18:32:07.379 - [INFO]: 【实时通知】事件广播成功 operation=(a+b)-c result=123.45\n"
    if string(output) != expected {
        t.Fatalf("unexpected output: %s", string(output))
    }
}

// TestFormatterFactory 测试格式器工厂
func TestFormatterFactory(t *testing.T) {
    factory := &FormatterFactory{}

    // 测试 withField 格式器（默认）
    settings := NewSettings()
    settings.FormatterType = FormatterTypeWithField
    formatter := factory.CreateFormatter(settings)
    if _, ok := formatter.(*WithFieldFormatter); !ok {
        t.Fatalf("expected WithFieldFormatter, got %T", formatter)
    }

    // 测试 JSON 格式器
    settings.FormatterType = FormatterTypeJSON
    formatter = factory.CreateFormatter(settings)
    if _, ok := formatter.(*logrus.JSONFormatter); !ok {
        t.Fatalf("expected JSONFormatter, got %T", formatter)
    }

    // 测试 Text 格式器
    settings.FormatterType = FormatterTypeText
    formatter = factory.CreateFormatter(settings)
    if _, ok := formatter.(*logrus.TextFormatter); !ok {
        t.Fatalf("expected TextFormatter, got %T", formatter)
    }

    // 测试 easy 格式器
    settings.FormatterType = FormatterTypeEasy
    formatter = factory.CreateFormatter(settings)
    if _, ok := formatter.(*easy.Formatter); !ok {
        t.Fatalf("expected easy.Formatter, got %T", formatter)
    }

    // 测试自定义格式器
    customFormatter := &logrus.JSONFormatter{}
    settings.CustomFormatter = customFormatter
    formatter = factory.CreateFormatter(settings)
    if formatter != customFormatter {
        t.Fatalf("expected custom formatter, got %T", formatter)
    }
}

// TestFormatterSettingsFromYAML 测试从 YAML 加载格式器配置
func TestFormatterSettingsFromYAML(t *testing.T) {
    root, err := os.MkdirTemp("", "logger-fmt-ut")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(root)

    yamlPath := filepath.Join(root, "logger.yaml")
    yaml := strings.Join([]string{
        "log_root: '" + filepath.ToSlash(root) + "'",
        "log_name_base: 'fmt_test'",
        "level: 'debug'",
        "days_to_keep: 7",
        "formatter_type: 'withField'",
        "timestamp_format: '2006-01-02T15:04:05.000Z'",
        "disable_level: false",
        "disable_caller: true",
        "",
    }, "\n")

    if err := ioutil.WriteFile(yamlPath, []byte(yaml), 0644); err != nil {
        t.Fatal(err)
    }

    settings, err := LoadSettingsFromYAML(yamlPath)
    if err != nil {
        t.Fatal(err)
    }

    if settings.FormatterType != FormatterTypeWithField {
        t.Fatalf("expected formatter_type 'withField', got %s", settings.FormatterType)
    }

    if settings.TimestampFormat != "2006-01-02T15:04:05.000Z" {
        t.Fatalf("unexpected timestamp_format: %s", settings.TimestampFormat)
    }

    if settings.DisableLevel != false {
        t.Fatalf("expected disable_level false, got %v", settings.DisableLevel)
    }

    if settings.DisableCaller != true {
        t.Fatalf("expected disable_caller true, got %v", settings.DisableCaller)
    }
}

// TestBackwardCompatibility 测试向后兼容性
func TestBackwardCompatibility(t *testing.T) {
    // 测试 OnlyMsg 向后兼容
    settings := NewSettings()
    settings.OnlyMsg = true
    // 不设置 FormatterType，应该默认使用 easy 格式器

    factory := &FormatterFactory{}
    formatter := factory.CreateFormatter(settings)

    // 应该返回 easy formatter
    if _, ok := formatter.(*easy.Formatter); !ok {
        t.Fatalf("expected easy.Formatter for OnlyMsg=true, got %T", formatter)
    }
}

// TestSetLoggerWithFormatter 测试设置不同格式器的日志器
func TestSetLoggerWithFormatter(t *testing.T) {
    // 保存原始设置
    originalLogger := loggerBase

    defer func() {
        loggerBase = originalLogger
    }()

    // 测试使用 withField 格式器
    settings := NewSettings()
    settings.FormatterType = FormatterTypeWithField
    SetLoggerSettings(settings)

    logger := GetLogger()
    if _, ok := logger.Formatter.(*WithFieldFormatter); !ok {
        t.Fatalf("expected WithFieldFormatter, got %T", logger.Formatter)
    }

    // 测试使用 JSON 格式器
    settings.FormatterType = FormatterTypeJSON
    SetLoggerSettings(settings)

    logger = GetLogger()
    if _, ok := logger.Formatter.(*logrus.JSONFormatter); !ok {
        t.Fatalf("expected JSONFormatter, got %T", logger.Formatter)
    }
}
