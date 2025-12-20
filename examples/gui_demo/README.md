# GUI Demo

Windows GUI模式日志示例，演示在无控制台环境下的日志处理。

## 功能特性

- Windows GUI模式检测
- 无控制台环境日志输出
- 文件日志写入
- GUI应用程序兼容性

## 运行方法

```bash
# 编译为GUI应用程序
go build -ldflags "-H=windowsgui" -o gui_demo.exe gui_demo.go

# 运行（双击或在命令行运行）
./gui_demo.exe
```

## GUI模式特性

- 自动检测Windows GUI环境
- 优雅处理无控制台情况
- 仅输出到日志文件
- 支持Windows服务模式

## 测试场景

1. **控制台模式** - 正常控制台输出+文件输出
2. **GUI模式** - 仅文件输出
3. **混合模式** - 根据环境自动选择输出方式