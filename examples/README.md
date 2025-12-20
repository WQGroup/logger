# Examples

这个目录包含了各种日志库使用示例。

## 示例列表

### 1. [Basic Demo](./demo/)
基本日志使用示例，展示核心功能：
- 基本日志输出
- 多种格式器
- YAML配置
- 日志轮转

### 2. [Formatter Demo](./formatter_demo/)
格式器使用示例：
- JSON格式
- 文本格式
- 结构化日志
- 自定义格式器

### 3. [Concurrent Demo](./concurrent_demo/)
并发日志示例：
- 多goroutine并发写入
- 线程安全
- 性能测试
- 高吞吐量

### 4. [GUI Demo](./gui_demo/)
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

# 并发示例
cd concurrent_demo
./concurrent_demo.exe

# GUI示例
cd gui_demo
./gui_demo.exe
```

## 注意事项

- 所有示例都已在Windows环境下测试通过
- 并发示例展示了修复后的线程安全性
- GUI示例演示了Windows特定功能
- 建议按顺序运行示例以了解不同功能