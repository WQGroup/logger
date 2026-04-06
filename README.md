# Logger

一个基于 [logrus](https://github.com/sirupsen/logrus) 的 Go 日志库，提供日志轮转、自动清理和灵活配置。

## 快速开始

```go
import "github.com/WQGroup/logger"

func main() {
    // 零配置，直接使用
    logger.Info("应用程序启动")
    // 输出：2024-01-01 12:00:00.123 - [INFO]: 应用程序启动

    logger.WithField("user_id", 12345).Info("用户登录")
    // 输出：2024-01-01 12:00:00.123 - [INFO]: 用户登录 user_id=12345
}
```

默认日志级别为 `Info`，日志文件写入 `./Logs/` 目录，按天轮转，保留 7 天。

## 配置方式

### 通过代码

```go
import "github.com/sirupsen/logrus"

settings := logger.NewSettings()
settings.LogNameBase = "myapp"
settings.Level = logrus.DebugLevel
settings.MaxAgeDays = 30
settings.MaxSizeMB = 100
settings.UseHierarchicalPath = true

logger.SetLoggerSettings(settings)
```

### 通过 YAML

创建 `config.yaml`：

```yaml
log_root: "/var/log/myapp"
log_name_base: "myapp"
level: "info"
days_to_keep: 7
max_size_mb: 0
use_hierarchical_path: false
```

```go
err := logger.SetLoggerFromYAML("config.yaml")
if err != nil {
    panic(err)
}
```

## 日志存储

默认扁平结构（`UseHierarchicalPath = false`）：

```
./Logs/
├── myapp--202401010800--.log
├── myapp--202401020800--.log
└── myapp--202401030800--.log
```

分层结构（`UseHierarchicalPath = true`）按 `YYYY/MM/DD` 组织目录。

## 更多

- [高级用法](./docs/ADVANCED.md) — 格式器详解、轮转策略、API 参考、完整配置项
- [示例代码](./examples) — 各功能完整示例

## 依赖

- [sirupsen/logrus](https://github.com/sirupsen/logrus) - 结构化日志库
- [lestrrat-go/file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) - 时间轮转
- [natefinch/lumberjack](https://github.com/natefinch/lumberjack) - 大小轮转
- [go-yaml/yaml](https://github.com/go-yaml/yaml) - YAML 配置解析

## 许可证

MIT
