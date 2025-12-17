# Logger

一个基于 [logrus](https://github.com/sirupsen/logrus) 的 Go 日志库，提供了日志轮转、自动清理、YAML 配置支持和分层路径存储等功能。

## 主要特性

- **双重轮转策略**：支持基于时间（默认24小时）和基于文件大小的日志轮转
- **自动日志清理**：自动删除超过指定天数的旧日志文件和空目录
- **分层路径结构**：支持按年/月/日（YYYY/MM/DD）的分层存储结构
- **YAML 配置支持**：可通过配置文件设置日志参数
- **Windows GUI 兼容**：特殊处理 Windows GUI 模式下的日志输出
- **毫秒级时间戳**：日志时间戳包含毫秒精度，便于调试
- **所有 logrus 日志级别**：Trace、Debug、Info、Warn、Error、Fatal、Panic

## 快速开始

### 基本使用

```go
import "github.com/WQGroup/logger"

// 设置日志名称
logger.SetLoggerName("MyApp")
logger.Info("应用程序启动")
// 输出：2024-01-01 12:00:00.123 - [INFO]: 应用程序启动
```

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
    OnlyMsg             bool          // 是否只输出消息内容，不包含时间戳等
    Level               logrus.Level  // 日志级别（默认 InfoLevel）
    LogRootFPath        string        // 日志根目录（默认当前目录）
    LogNameBase         string        // 日志文件名前缀（默认 "logger"）
    RotationTime        time.Duration // 轮转时间间隔（默认24小时）
    MaxAge              time.Duration // 日志最大保存时间（已弃用，使用MaxAgeDays）
    MaxAgeDays          int           // 日志最大保存天数（默认7天）
    MaxSizeMB           int           // 文件大小限制(MB)，0表示不启用大小轮转
    UseHierarchicalPath bool          // 是否使用分层路径 YYYY/MM/DD（默认false）
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

### 大小轮转
当设置 `MaxSizeMB > 0` 时：
- 文件超过指定大小时立即轮转
- 使用 lumberjack 进行轮转
- 文件名格式：`logger.log`
- 文件名后会自动添加序号，如 `logger-2024-01-01.log.1`

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
```

## 依赖

* [sirupsen/logrus](https://github.com/sirupsen/logrus) v1.6.0 - 结构化日志库
* [lestrrat-go/file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) v2.4.0 - 日志文件轮转
* [t-tomalak/logrus-easy-formatter](https://github.com/t-tomalak/logrus-easy-formatter) - 简单的日志格式化器
* [natefinch/lumberjack](https://github.com/natefinch/lumberjack) v2.2.1 - 日志文件轮转（备选方案）
* [yaml.v3](https://github.com/go-yaml/yaml) v3.0.1 - YAML 配置解析

## 许可证

本项目采用 MIT 许可证。