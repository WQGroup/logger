# Formatter Demo

格式器示例，展示各种日志格式的使用方法。

## 功能特性

- JSON格式日志
- 文本格式日志
- WithField结构化日志
- 自定义格式器
- 时间戳格式配置

## 运行方法

```bash
# 编译
go build -o formatter_examples.exe *.go

# 运行
./formatter_examples.exe
```

## 支持的格式器

1. **JSON格式** - 结构化日志输出
2. **Text格式** - 纯文本格式
3. **WithField格式** - 结构化字段输出
4. **Easy格式** - 简单格式化
5. **自定义格式** - 用户自定义格式