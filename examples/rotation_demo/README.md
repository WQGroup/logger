# Rotation Demo

日志轮转与清理功能示例，演示日志库的核心特性。

## 功能特性

- **时间轮转** - 基于时间间隔的日志轮转
- **大小轮转** - 基于文件大小的日志轮转
- **分层路径** - YYYY/MM/DD 结构的日志存储
- **自动清理** - 过期日志文件的自动删除
- **Easy 格式器** - 支持自定义格式模板

## 运行方法

```bash
# 编译
go build -o rotation_demo.exe rotation_demo.go

# 运行
./rotation_demo.exe
```

## 演示内容

### 1. 时间轮转功能
- 轮转间隔：5秒（演示用）
- 自动创建新的日志文件
- 文件名格式：`logger--YYYYMMDDHHMM--.log`

### 2. 大小轮转功能
- 大小限制：1MB（演示用）
- 文件超过限制时自动轮转
- 使用 lumberjack 进行轮转

### 3. 分层路径结构
- 路径格式：`YYYY/MM/DD/`
- 便于日志文件的查找和管理
- 支持按日期快速定位日志

### 4. 自动清理功能
- 自动删除过期日志
- 同时清理空目录
- 保持日志目录的整洁

### 5. Easy 格式器
- 支持自定义格式模板
- 使用占位符：`%time%`, `%lvl%`, `%msg%`, `%fields%`
- 兼容旧版本的格式配置

## 配置参数

```go
// 时间轮转
RotationTime: 24 * time.Hour  // 默认24小时

// 大小轮转
MaxSizeMB: 100  // 100MB，0表示不启用

// 自动清理
MaxAgeDays: 7   // 保存7天

// 分层路径
UseHierarchicalPath: true  // 启用分层结构

// Easy格式器
FormatterType: "easy"
LogFormat: "%time% [%lvl%]: %msg%\n"
```

## 输出目录结构

```
./logs/
├── time_rotation/
│   └── *.log              # 时间轮转日志
├── size_rotation/
│   └── *.log              # 大小轮转日志
├── hierarchical/
│   └── YYYY/MM/DD/        # 分层路径结构
├── cleanup_demo/
│   └── *.log              # 清理演示日志
└── easy_formatter/
    └── *.log              # Easy格式器日志
```