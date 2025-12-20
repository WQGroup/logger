# Logger

一个基于 [logrus](https://github.com/sirupsen/logrus) 的 Go 日志库，提供了日志轮转、自动清理、YAML 配置支持和分层路径存储等功能。

## 主要特性

- **双重轮转策略**：支持基于时间（默认24小时）和基于文件大小的日志轮转
- **自动日志清理**：自动删除超过指定天数的旧日志文件和空目录
- **分层路径结构**：支持按年/月/日（YYYY/MM/DD）的分层存储结构
- **YAML 配置支持**：可通过配置文件设置日志参数
- **Windows GUI 兼容**：特殊处理 Windows GUI 模式下的日志输出
- **毫秒级时间戳**：日志时间戳包含毫秒精度，便于调试
- **灵活的格式器支持**：支持多种日志格式器（withField、easy、json、text）
- **结构化日志字段**：自动将 logrus 字段以 key=value 格式追加到日志后
- **自定义格式器**：支持用户自定义日志格式器
- **向后兼容**：完全兼容旧版本配置和格式

## 快速开始

### 基本使用（默认使用 withField 格式器）

```go
import "github.com/WQGroup/logger"

// 设置日志名称
logger.SetLoggerName("MyApp")
logger.Info("应用程序启动")
// 输出：2024-01-01 12:00:00.123 - [INFO]: 应用程序启动

// 使用结构化字段
logger.WithField("user_id", 12345).Info("用户登录")
// 输出：2024-01-01 12:00:00.123 - [INFO]: 用户登录 user_id=12345

// 使用多个字段
logger.WithFields(map[string]interface{}{
    "module": "auth",
    "action": "login",
    "user": "john",
}).Info("认证成功")
// 输出：2024-01-01 12:00:00.123 - [INFO]: 认证成功 module=auth action=login user=john
```

### 示例代码

查看 [examples](./examples) 目录获取完整的使用示例：

- [Basic Demo](./examples/demo) - 基本功能和格式器演示
- [Formatter Demo](./examples/formatter_demo) - 各种格式器的详细使用
- [Rotation Demo](./examples/rotation_demo) - 日志轮转和自动清理功能
- [Concurrent Demo](./examples/concurrent_demo) - 并发安全演示
- [GUI Demo](./examples/gui_demo) - Windows GUI 模式使用

### 通过 Settings 配置

```go
import "github.com/sirupsen/logrus"

settings := logger.NewSettings()
settings.LogNameBase = "myapp"
settings.Level = logrus.DebugLevel
settings.MaxAgeDays = 30        // 保存30天
settings.MaxSizeMB = 100        // 文件超过100MB时轮转
settings.UseHierarchicalPath = true  // 使用分层路径

logger.SetLoggerSettings(settings)
logger.Debug("这是调试信息")
```

### 使用 YAML 配置文件

创建 `config.yaml`：

```yaml
log_root: "/var/log/myapp"           # 日志根目录
log_name_base: "myapp"               # 日志文件名前缀
level: "info"                        # 日志级别
days_to_keep: 7                      # 保存天数
max_size_mb: 0                       # 文件大小限制(MB)，0表示不启用
use_hierarchical_path: true          # 是否使用分层路径

# 格式器配置
formatter_type: "withField"          # 格式器类型: withField, easy, json, text
timestamp_format: "2006-01-02 15:04:05.000"  # 时间戳格式
disable_timestamp: false             # 是否禁用时间戳
disable_level: false                 # 是否禁用日志级别
disable_caller: true                 # 是否禁用调用者信息
full_timestamp: false                # 是否显示完整时间戳
log_format: "%time% - [%lvl%]: %msg%\n"  # 自定义日志格式（仅用于 easy 格式器）
```

在代码中使用：

```go
err := logger.SetLoggerFromYAML("config.yaml")
if err != nil {
    panic(err)
}
```

## 配置选项

### Settings 结构体

```go
type Settings struct {
    // 基本配置
    OnlyMsg             bool          // 废弃：仅输出消息，向后兼容
    Level               logrus.Level  // 日志级别（默认 InfoLevel）
    LogRootFPath        string        // 日志根目录（默认当前目录）
    LogNameBase         string        // 日志文件名前缀（默认 "logger"）
    RotationTime        time.Duration // 轮转时间间隔（默认24小时）
    MaxAge              time.Duration // 日志最大保存时间（已弃用，使用MaxAgeDays）
    MaxAgeDays          int           // 日志最大保存天数（默认7天）
    MaxSizeMB           int           // 文件大小限制(MB)，0表示不启用大小轮转
    UseHierarchicalPath bool          // 是否使用分层路径 YYYY/MM/DD（默认false）

    // 格式器配置
    FormatterType       string            // 格式器类型："withField", "easy", "json", "text"
    TimestampFormat     string            // 时间戳格式（默认 "2006-01-02 15:04:05.000"）
    CustomFormatter     logrus.Formatter  // 用户自定义格式器
    DisableTimestamp    bool              // 是否禁用时间戳
    DisableLevel        bool              // 是否禁用日志级别
    DisableCaller       bool              // 是否禁用调用者信息
    FullTimestamp       bool              // 是否显示完整时间戳
    LogFormat           string            // 自定义日志格式（用于 easy-formatter）
}
```

### 日志级别

支持的日志级别（从低到高）：
- `trace` - 最详细的日志信息
- `debug` - 调试信息
- `info` - 一般信息（默认）
- `warn` / `warning` - 警告信息
- `error` - 错误信息
- `fatal` - 致命错误，程序会退出
- `panic` - 恐慌错误，程序会 panic

## 格式器支持

### WithField 格式器（默认）

支持结构化字段的格式器，自动将 logrus 字段以 key=value 格式追加到日志后。

```go
// 使用 Settings 配置
settings := logger.NewSettings()
settings.FormatterType = logger.FormatterTypeWithField
logger.SetLoggerSettings(settings)

// 输出示例：2025-12-18 18:32:07.379 - [INFO]: 【实时通知】事件广播成功 operation=(a+b)-c result=123.45
logger.WithField("operation", "(a+b)-c").
       WithField("result", 123.45).
       Info("【实时通知】事件广播成功")
```

### JSON 格式器

输出 JSON 格式的日志，便于日志分析工具处理。

```go
settings.FormatterType = logger.FormatterTypeJSON
logger.SetLoggerSettings(settings)

// 输出示例：{"level":"info","msg":"用户登录","time":"2025-12-18T18:32:07.379+08:00","user_id":12345}
logger.WithField("user_id", 12345).Info("用户登录")
```

### Easy 格式器

兼容旧版本的格式器，支持自定义日志格式模板。

```go
settings.FormatterType = logger.FormatterTypeEasy
settings.LogFormat = "%time% [%lvl%] %msg%\n"  // 自定义格式模板
settings.TimestampFormat = "2006-01-02 15:04:05.000"  // 时间戳格式
logger.SetLoggerSettings(settings)

// 输出示例：2025-12-18 18:32:07.379 [INFO] 用户登录
logger.Info("用户登录")
```

支持的占位符：
- `%time%` - 时间戳
- `%lvl%` - 日志级别
- `%msg%` - 日志消息
- `%fields%` - 结构化字段

### Text 格式器

使用 logrus 的原生文本格式器。

```go
settings.FormatterType = logger.FormatterTypeText
logger.SetLoggerSettings(settings)

// 输出示例：INFO[0000] 用户登录
logger.Info("用户登录")
```

### 自定义格式器

用户可以实现自己的格式器。

```go
import "github.com/sirupsen/logrus"

// 实现自定义格式器
type MyFormatter struct{}

func (f *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    // 自定义格式逻辑
    return []byte(fmt.Sprintf("[CUSTOM] %s: %s\n", entry.Level, entry.Message)), nil
}

// 使用自定义格式器
settings := logger.NewSettings()
settings.CustomFormatter = &MyFormatter{}
logger.SetLoggerSettings(settings)
```

## 日志存储格式

### 扁平结构（默认）

```
./Logs/
├── logger--202401010800--.log
├── logger--202401010900--.log
└── logger--202401011000--.log
```

### 分层结构（UseHierarchicalPath=true）

```
./Logs/
├── 2024/
│   ├── 01/
│   │   ├── 01/
│   │   │   ├── logger--0800--.log
│   │   │   ├── logger--0900--.log
│   │   │   └── logger--1000--.log
│   │   └── 02/
│   │       └── logger--0800--.log
│   └── 02/
│       └── 01/
│           └── logger--0800--.log
```

## 轮转策略

### 时间轮转（默认）
- 默认每24小时创建一个新的日志文件
- 文件名格式：`logger--YYYYMMDDHHMM--.log`
- 自动删除超过 MaxAgeDays 天数的日志文件
- 可通过 `RotationTime` 设置轮转间隔

### 大小轮转
当设置 `MaxSizeMB > 0` 时：
- 文件超过指定大小时立即轮转
- 使用 lumberjack 进行轮转
- 文件名格式：`logger.log`
- 文件名后会自动添加序号，如 `logger-2024-01-01.log.1`
- 支持同时使用时间和大小轮转策略

### 示例配置
```go
settings := logger.NewSettings()
settings.RotationTime = 24 * time.Hour      // 24小时轮转一次
settings.MaxSizeMB = 100                    // 文件超过100MB时轮转
settings.MaxAgeDays = 30                    // 保存30天的日志
settings.UseHierarchicalPath = true         // 使用分层路径 YYYY/MM/DD
logger.SetLoggerSettings(settings)
```

## API 参考

### 基本日志方法
```go
logger.Debug("调试信息")
logger.Info("一般信息")
logger.Warn("警告信息")
logger.Error("错误信息")
logger.Fatal("致命错误")
logger.Panic("恐慌错误")

// 带格式化
logger.Infof("用户 %s 登录成功", username)
logger.Errorf("连接失败: %v", err)
```

### 辅助方法
```go
// 获取当前日志文件路径
path := logger.LogLinkFileFPath()

// 获取当前日志文件名
filename := logger.CurrentFileName()

// 创建默认设置
settings := logger.NewSettings()

// 从 YAML 加载设置
settings, err := logger.LoadSettingsFromYAML("config.yaml")

// 设置自定义格式器
logger.SetCustomFormatter(&MyFormatter{})
```

### 格式器常量

```go
const (
    FormatterTypeWithField = "withField"  // 默认格式器，支持字段输出
    FormatterTypeEasy      = "easy"       // 兼容旧版本的格式器
    FormatterTypeJSON      = "json"       // JSON 格式器
    FormatterTypeText      = "text"       // 文本格式器
)
```

## 依赖

* [sirupsen/logrus](https://github.com/sirupsen/logrus) v1.6.0 - 结构化日志库
* [lestrrat-go/file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) v2.4.0 - 日志文件轮转
* [t-tomalak/logrus-easy-formatter](https://github.com/t-tomalak/logrus-easy-formatter) - 简单的日志格式化器
* [natefinch/lumberjack](https://github.com/natefinch/lumberjack) v2.2.1 - 日志文件轮转（备选方案）
* [yaml.v3](https://github.com/go-yaml/yaml) v3.0.1 - YAML 配置解析

## 许可证

本项目采用 MIT 许可证。