# Concurrent Demo

并发日志示例，展示高并发场景下的日志处理能力。

## 功能特性

- 多goroutine并发写入
- 并发安全的日志操作
- 性能测试和统计
- 大量数据写入测试

## 运行方法

```bash
# 编译
go build -o concurrent_demo.exe concurrent_demo.go

# 运行
./concurrent_demo.exe
```

## 性能指标

- 支持数千个并发goroutine
- 平均写入速度：16,000+ 条日志/秒
- 线程安全的日志操作
- 无数据丢失或损坏

## 测试场景

1. **并发写入测试** - 50个goroutine，每个写入100条日志
2. **大数据量测试** - 单个goroutine写入大量日志
3. **性能基准测试** - 测量写入吞吐量