# Examples Directory Structure

```
examples/
├── README.md                           # 主说明文档
├── STRUCTURE.md                        # 本文档，目录结构说明
├── demo/                               # 基本功能示例
│   ├── README.md                       # 基本示例说明
│   ├── demo.go                         # 基本示例源代码
│   ├── demo.exe                        # 编译后的可执行文件
│   └── Logs/                           # 日志输出目录
│       └── *.log                       # 生成的日志文件
├── formatter_demo/                     # 格式器示例
│   ├── README.md                       # 格式器示例说明
│   ├── formatter_examples.go           # 格式器示例源代码
│   ├── formatter_examples.exe          # 编译后的可执行文件
│   └── logs/                           # 日志输出目录
│       └── *.log                       # 各种格式的日志文件
├── rotation_demo/                      # 日志轮转示例
│   ├── README.md                       # 轮转示例说明
│   ├── rotation_demo.go                # 轮转示例源代码
│   ├── rotation_demo.exe               # 编译后的可执行文件
│   └── logs/                           # 日志输出目录
│       ├── time_rotation/              # 时间轮转日志
│       ├── size_rotation/              # 大小轮转日志
│       ├── hierarchical/               # 分层路径日志
│       ├── cleanup_demo/               # 清理演示日志
│       └── easy_formatter/             # Easy格式器日志
├── concurrent_demo/                    # 并发示例
│   ├── README.md                       # 并发示例说明
│   ├── concurrent_demo.go              # 并发示例源代码
│   ├── concurrent_demo.exe             # 编译后的可执行文件
│   └── logs/                           # 日志输出目录
│       └── *.log                       # 并发测试日志文件
└── gui_demo/                           # GUI模式示例
    ├── README.md                       # GUI示例说明
    ├── gui_demo.go                     # GUI示例源代码
    └── gui_demo.exe                    # 编译后的可执行文件
```

## 目录说明

### 根目录文件
- `README.md` - 总体说明文档，介绍所有示例
- `STRUCTURE.md` - 本文档，说明目录结构

### 示例目录
每个示例都有独立的目录，包含：
- `README.md` - 详细的使用说明
- `*.go` - 源代码文件
- `*.exe` - 编译后的可执行文件
- `Logs/` - 日志输出目录（部分示例有）

## 使用方法

1. 进入感兴趣的示例目录
2. 阅读该目录下的 README.md
3. 按照说明编译和运行示例

## 编译命令

每个示例都可以独立编译：

```bash
# 基本示例
cd examples/demo
go build -o demo.exe demo.go

# 格式器示例
cd examples/formatter_demo
go build -o formatter_examples.exe *.go

# 并发示例
cd examples/concurrent_demo
go build -o concurrent_demo.exe concurrent_demo.go

# GUI示例
cd examples/gui_demo
go build -ldflags "-H=windowsgui" -o gui_demo.exe gui_demo.go
```

## 日志文件

运行示例后，日志文件会生成在各自目录的 `Logs/` 子目录中，方便查看和比较不同示例的输出。