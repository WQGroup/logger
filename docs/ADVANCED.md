# Logger 高级用法

本文档涵盖 logger 库的高级功能和详细配置。基本用法请参考 [README.md](../README.md)。

## 目录

- [格式器详解](#格式器详解)
- [轮转策略](#轮转策略)
- [API 参考](#api-参考)
- [完整配置项](#完整配置项)

## 格式器详解

logger 支持 4 种内置格式器，默认使用 `withField`。

### WithField 格式器（默认）

支持结构化字段，自动将 logrus 字段以 `key=value` 格式追加到日志后。

```go
settings := logger.NewSettings()
settings.FormatterType = logger.FormatterTypeWithField  // 默认值，可省略
logger.SetLoggerSettings(settings)

// 输出：2024-01-01 12:00:00.123 - [INFO]: 事件广播成功 operation=(a+b)-c result=123.45
logger.WithField("operation", "(a+b)-c").
       WithField("result", 123.45).
       Info("事件广播成功")
```

### JSON 格式器

输出 JSON 格式，便于日志分析工具处理。

```go
settings := logger.NewSettings()
settings.FormatterType = logger.FormatterTypeJSON
logger.SetLoggerSettings(settings)

// 输出：{"level":"info","msg":"用户登录","time":"2024-01-01T12:00:00.123+08:00","user_id":12345}
logger.WithField("user_id", 12345).Info("用户登录")
```

### Easy 格式器

兼容旧版本的格式器，支持自定义格式模板。

```go
settings := logger.NewSettings()
settings.FormatterType = logger.FormatterTypeEasy
settings.LogFormat = "%time% [%lvl%] %msg%\n"
logger.SetLoggerSettings(settings)

// 输出：2024-01-01 12:00:00.123 [INFO] 用户登录
logger.Info("用户登录")
```

支持的占位符：

| 占位符 | 说明 |
|--------|------|
| `%time%` | 时间戳 |
| `%lvl%` | 日志级别 |
| `%msg%` | 日志消息 |
| `%fields%` | 结构化字段 |

### Text 格式器

使用 logrus 的原生文本格式器。

```go
settings := logger.NewSettings()
settings.FormatterType = logger.FormatterTypeText
settings.FullTimestamp = true
logger.SetLoggerSettings(settings)

// 输出：INFO[2024-01-01 12:00:00.123] 用户登录
logger.Info("用户登录")
```

### 自定义格式器

实现 `logrus.Formatter` 接口即可。

```go
import "github.com/sirupsen/logrus"

type MyFormatter struct{}

func (f *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    return []byte(fmt.Sprintf("[CUSTOM] %s: %s\n", entry.Level, entry.Message)), nil
}

settings := logger.NewSettings()
settings.CustomFormatter = &MyFormatter{}
logger.SetLoggerSettings(settings)
```

## 轮转策略

### 时间轮转（默认）

- 默认每 24 小时创建新文件
- 文件名格式：`logger--YYYYMMDDHHMM--.log`
- 可通过 `RotationTime` 自定义间隔

### 大小轮转

设置 `MaxSizeMB > 0` 时启用：
- 文件超过指定大小立即轮转（使用 lumberjack）
- 文件名格式：`logger.log`，自动追加序号如 `logger-2024-01-01.log.1`

### 分层路径

设置 `UseHierarchicalPath = true` 后按 `YYYY/MM/DD` 分层存储：

```
./Logs/
├── 2024/
│   ├── 01/
│   │   ├── 01/
│   │   │   ├── logger--0800--.log
│   │   │   └── logger--0900--.log
│   │   └── 02/
│   │       └── logger--0800--.log
```

### 轮转配置示例

```go
settings := logger.NewSettings()
settings.RotationTime = 24 * time.Hour      // 24小时轮转一次
settings.MaxSizeMB = 100                    // 文件超过100MB时轮转
settings.MaxAgeDays = 30                    // 保存30天的日志
settings.UseHierarchicalPath = true         // 使用分层路径
logger.SetLoggerSettings(settings)
```

## API 参考

### 日志方法

```go
// 基本方法
logger.Debug("调试信息")
logger.Info("一般信息")
logger.Warn("警告信息")
logger.Error("错误信息")
logger.Fatal("致命错误")   // 会调用 os.Exit(1)
logger.Panic("恐慌错误")   // 会 panic

// 格式化方法
logger.Infof("用户 %s 登录成功", username)
logger.Errorf("连接失败: %v", err)

// 换行方法
logger.Infoln("自动换行")

// 结构化字段
logger.WithField("user_id", 12345).Info("用户登录")
logger.WithFields(map[string]interface{}{
    "module": "auth",
    "action": "login",
}).Info("认证成功")
```

### 辅助函数

```go
// 获取当前日志文件路径
path := logger.LogLinkFileFPath()

// 获取当前日志文件名
filename := logger.CurrentFileName()

// 设置自定义格式器
logger.SetCustomFormatter(&MyFormatter{})

// 从 YAML 加载设置
settings, err := logger.LoadSettingsFromYAML("config.yaml")

// 关闭日志器（程序退出前调用）
logger.Close()
```

### 格式器常量

```go
const (
    FormatterTypeWithField = "withField"  // 默认，支持字段输出
    FormatterTypeEasy      = "easy"       // 兼容旧版本
    FormatterTypeJSON      = "json"       // JSON 格式
    FormatterTypeText      = "text"       // logrus 原生文本
)
```

## 完整配置项

```go
type Settings struct {
    // 基本配置
    OnlyMsg             bool          // 废弃：仅输出消息，向后兼容
    Level               logrus.Level  // 日志级别（默认 InfoLevel）
    LogRootFPath        string        // 日志根目录（默认当前目录）
    LogNameBase         string        // 日志文件名前缀（默认 "logger"）
    RotationTime        time.Duration // 轮转时间间隔（默认24小时）
    MaxAge              time.Duration // 日志最大保存时间（已弃用，使用 MaxAgeDays）
    MaxAgeDays          int           // 日志最大保存天数（默认7天）
    MaxSizeMB           int           // 文件大小限制(MB)，0表示不启用大小轮转
    UseHierarchicalPath bool          // 是否使用分层路径 YYYY/MM/DD（默认false）

    // 格式器配置
    FormatterType    string           // 格式器类型："withField", "easy", "json", "text"
    TimestampFormat  string           // 时间戳格式（默认 "2006-01-02 15:04:05.000"）
    CustomFormatter  logrus.Formatter // 用户自定义格式器
    DisableTimestamp bool             // 是否禁用时间戳
    DisableLevel     bool             // 是否禁用日志级别
    DisableCaller    bool             // 是否禁用调用者信息（默认 true）
    FullTimestamp    bool             // 是否显示完整时间戳
    LogFormat        string           // 自定义日志格式（仅用于 easy 格式器）
}
```

### 日志级别

从低到高：`trace` < `debug` < `info`（默认）< `warn` < `error` < `fatal` < `panic`
