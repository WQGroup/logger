# Examples

这个目录包含了各种日志库使用示例。

## 示例列表

### 1. [Basic Demo](./demo/)
基本日志使用示例，展示核心功能：
- 基本日志输出
- 多种格式器
- 自定义格式器
- 时间戳格式

### 2. [Formatter Demo](./formatter_demo/)
格式器使用示例：
- JSON格式
- 文本格式
- 结构化日志
- 自定义格式器
- YAML配置

### 3. [Rotation Demo](./rotation_demo/)
日志轮转与清理功能示例：
- 时间轮转功能
- 大小轮转功能
- 分层路径结构（YYYY/MM/DD）
- 自动日志清理
- Easy格式器

### 4. [Concurrent Demo](./concurrent_demo/)
并发日志示例：
- 多goroutine并发写入
- 线程安全
- 性能测试
- 高吞吐量

### 5. [GUI Demo](./gui_demo/)
Windows GUI模式示例：
- GUI环境检测
- 无控制台输出
- Windows服务兼容
- 文件日志

## 运行示例

每个示例都有独立的目录和README文件。进入对应的目录查看详细说明：

```bash
# 基本示例
cd demo
./demo.exe

# 格式器示例
cd formatter_demo
./formatter_examples.exe

# 轮转功能示例
cd rotation_demo
./rotation_demo.exe

# 并发示例
cd concurrent_demo
./concurrent_demo.exe

# GUI示例
cd gui_demo
./gui_demo.exe
```

## 推荐学习路径

1. **Basic Demo** - 了解基本功能和多种格式器
2. **Formatter Demo** - 深入学习各种日志格式化选项
3. **Rotation Demo** - 掌握日志轮转、分层路径和自动清理
4. **Concurrent Demo** - 了解并发安全性
5. **GUI Demo** - 学习Windows特定功能

## 注意事项

- 所有示例都已在Windows环境下测试通过
- 并发示例展示了修复后的线程安全性
- GUI示例演示了Windows特定功能
- Rotation示例演示了日志管理的核心特性
- 建议按推荐学习路径运行示例