# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🔧 Claude Code 交互规则

### 开发环境说明
- **使用语言**: 请使用中文回答问题
- **操作系统**: 当前开发系统为 Windows
- **编码规范**: 遵循项目已有的代码风格

### 脚本编辑规则
- **BAT 脚本**: 编辑 BAT 脚本时请避免使用中文字符，保持 ASCII 编码
- **修改原则**: 在原有脚本基础上进行修复，除非必需，否则不要创建新的脚本文件
- **注释语言**: 脚本中的注释可以使用中文

## 项目概述

这是一个基于 logrus 的 Go 日志库，提供了日志轮转、自动清理和灵活配置等功能。特别适合需要长期运行的服务端应用程序。

## 构建和测试命令

```powershell
# Build module
go build

# Run tests
go test -v

# Run tests with coverage
go test -v -cover

# Run specific test
go test -v -run TestCleanupExpired

# Format code
go fmt

# Check code specification
go vet

# Update dependencies
go mod tidy
```

### Windows 特定命令

```powershell
# Clean build cache
go clean -cache

# Run tests with race detection (if needed)
go test -race -v

# Build for Windows with specific architecture
go build -o logger.exe

# Install/update tools
go install github.com/air-verse/air@latest
```

### 开发工作流

```powershell
# 1. 安装依赖
go mod tidy

# 2. 运行所有测试
go test -v ./...

# 3. 检查代码格式和规范
go fmt ./...
go vet ./...

# 4. 构建项目
go build

# 5. 运行应用
.\logger.exe
```

## 核心架构

### 主要文件说明

1. **logger.go** - 核心实现
   - Settings 结构体：所有日志配置的定义
   - 日志轮转逻辑（基于时间和大小的双重策略）
   - Windows GUI 特殊处理（检测是否有控制台）

2. **logger_base.go** - logrus 封装
   - 提供所有 logrus 日志级别的包装方法
   - 方便用户直接使用 logger 而不是 logrus

3. **config_yaml.go** - YAML 配置支持
   - SetLoggerFromYAML() 函数从 YAML 文件加载配置
   - 支持 YAML 配置项与 Settings 结构体对应

4. **cleanup.go** - 日志清理功能
   - cleanupExpired() 删除超过 MaxAge 天数的日志
   - 支持分层路径结构（YYYY/MM/DD）下的清理
   - handleCleanupResult() 统计删除的文件和目录

### 核心功能

1. **双重轮转策略**
   - 时间轮转：默认 24 小时创建新文件
   - 大小轮转：当文件超过 MaxSizeMB 时轮转

2. **分层路径结构**
   - UseHierarchicalPath 为 true 时，日志按 YYYY/MM/DD 结构存储
   - 向后兼容传统扁平存储方式

3. **Windows 兼容性**
   - 通过 isRunningInConsole() 检测是否有控制台
   - GUI 模式下自动添加 writer 到防止日志丢失

4. **毫秒级时间戳**
   - 日志格式包含毫秒精度："2006-01-02 15:04:05.000"

## Settings 配置项详解

```go
type Settings struct {
    OnlyMsg               bool          // 仅输出消息，不包含时间戳等额外信息
    Level                 logrus.Level  // 日志级别（默认 InfoLevel）
    LogRootFPath          string        // 日志根目录（默认当前目录）
    LogNameBase           string        // 日志文件名前缀（默认 "logger"）
    RotationTime          time.Duration // 轮转时间间隔（默认 24 小时）
    MaxAge                int           // 日志最大保存天数（默认 7 天）
    MaxSizeMB             int           // 文件大小限制 MB，0 表示不启用
    UseHierarchicalPath   bool          // 是否使用分层路径 YYYY/MM/DD（默认 false）
}
```

## 使用示例

### 基本使用

```go
import "github.com/WQGroup/logger"

logger.SetLoggerName("MyApp")
logger.Info("应用程序启动")
```

### YAML 配置方式

```yaml
# config.yaml
level: info
logRootFPath: "/var/log/myapp"
logNameBase: "myapp"
rotationTime: 24h
maxAge: 30
maxSizeMB: 100
useHierarchicalPath: true
```

```go
err := logger.SetLoggerFromYAML("config.yaml")
if err != nil {
    panic(err)
}
```

### 代码配置方式

```go
settings := logger.NewSettings()
settings.LogNameBase = "myapp"
settings.Level = logrus.DebugLevel
settings.UseHierarchicalPath = true
settings.MaxSizeMB = 100
logger.SetLoggerSettings(settings)
```

## 注意事项

1. **Windows 特殊处理**：代码中包含对 Windows GUI 模式的特殊处理，确保在没有控制台的情况下日志也能正常输出

2. **清理机制**：日志会在轮转时自动触发清理，删除超过 MaxAge 天数的旧日志文件和空目录

3. **向后兼容**：UseHierarchicalPath 默认为 false，保持与旧版本的路径格式兼容

4. **依赖管理**：使用了 lumberjack 作为备选的日志轮转实现，但主要使用 rotatelogs

## 测试说明

主要测试用例：
- TestSetLoggerSettings：测试配置设置
- TestSetLoggerName：测试日志名称设置
- TestSetLoggerFromYAML：测试 YAML 配置加载
- TestHierarchicalPath：测试分层路径功能
- TestLogRotation：测试日志轮转
- TestCleanupExpired：测试日志清理功能

运行测试时注意清理生成的测试日志文件。