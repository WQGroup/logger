# Basic Demo

基本的日志使用示例，展示日志库的核心功能。

## 功能特性

- 基本的日志输出（Debug、Info、Warn、Error等）
- 多种日志格式器
- YAML配置文件支持
- 日志轮转和自动清理

## 运行方法

```bash
# 编译
go build -o demo.exe demo.go

# 运行
./demo.exe
```

## 文件说明

- `demo.go` - 源代码
- `demo.exe` - 编译后的可执行文件
- `config.yaml` - YAML配置文件
- `Logs/` - 日志文件输出目录