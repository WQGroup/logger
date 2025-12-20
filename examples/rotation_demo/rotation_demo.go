package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/WQGroup/logger"
)

func main() {
	fmt.Println("=== 日志轮转与清理功能演示 ===\n")

	// 演示1: 时间轮转功能
	demonstrateTimeRotation()

	// 演示2: 大小轮转功能
	demonstrateSizeRotation()

	// 演示3: 分层路径结构
	demonstrateHierarchicalPath()

	// 演示4: 自动清理功能
	demonstrateAutoCleanup()

	// 演示5: Easy 格式器
	demonstrateEasyFormatter()

	fmt.Println("所有演示完成！")
}

// demonstrateTimeRotation 演示时间轮转功能
func demonstrateTimeRotation() {
	fmt.Println("=== 1. 时间轮转功能演示 ===")

	settings := logger.NewSettings()
	settings.LogRootFPath = "./logs/time_rotation"
	settings.LogNameBase = "time_rotation_demo"
	settings.RotationTime = 1 * time.Minute // 设置较短的轮转时间用于演示（最小1分钟）
	settings.MaxAgeDays = 7
	settings.FormatterType = logger.FormatterTypeWithField

	logger.SetLoggerSettings(settings)

	fmt.Printf("日志轮转间隔设置为 %v\n", settings.RotationTime)
	fmt.Printf("日志文件位置: %s\n", logger.LogLinkFileFPath())

	// 记录初始文件
	initialFile := logger.CurrentFileName()
	fmt.Printf("初始日志文件: %s\n", initialFile)

	// 写入一些日志
	for i := 0; i < 5; i++ {
		logger.WithField("count", i+1).Info("时间轮转测试日志")
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("当前日志文件: %s\n", logger.CurrentFileName())
	fmt.Println("注意：每5秒会创建新的日志文件\n")
}

// demonstrateSizeRotation 演示大小轮转功能
func demonstrateSizeRotation() {
	fmt.Println("=== 2. 大小轮转功能演示 ===")

	settings := logger.NewSettings()
	settings.LogRootFPath = "./logs/size_rotation"
	settings.LogNameBase = "size_rotation_demo"
	settings.MaxSizeMB = 1 // 设置1MB的大小限制用于演示
	settings.FormatterType = logger.FormatterTypeWithField

	logger.SetLoggerSettings(settings)

	fmt.Printf("日志文件大小限制设置为 %d MB\n", settings.MaxSizeMB)
	fmt.Printf("日志文件位置: %s\n", logger.LogLinkFileFPath())

	// 生成大量日志以触发大小轮转
	for i := 0; i < 1000; i++ {
		logger.WithFields(logrus.Fields{
			"iteration": i,
			"data":      fmt.Sprintf("这是一段用于测试日志大小轮转的长文本数据，包含各种信息如时间戳、用户ID、操作类型等详细信息。当前是第%d次迭代，生成的数据应该足够触发文件大小轮转机制。", i),
			"user_id":   fmt.Sprintf("user_%d", i%10),
			"action":    []string{"login", "logout", "create", "update", "delete"}[i%5],
		}).Info("大小轮转测试日志")
	}

	fmt.Printf("当前日志文件: %s\n", logger.CurrentFileName())
	fmt.Println("注意：文件超过1MB时会自动轮转\n")
}

// demonstrateHierarchicalPath 演示分层路径结构
func demonstrateHierarchicalPath() {
	fmt.Println("=== 3. 分层路径结构演示 ===")

	settings := logger.NewSettings()
	settings.LogRootFPath = "./logs/hierarchical"
	settings.LogNameBase = "hierarchical_demo"
	settings.UseHierarchicalPath = true
	settings.FormatterType = logger.FormatterTypeWithField

	logger.SetLoggerSettings(settings)

	fmt.Printf("分层路径已启用\n")
	fmt.Printf("日志文件位置: %s\n", logger.LogLinkFileFPath())

	// 写入不同时间的日志来演示分层路径
	logger.Info("分层路径结构测试")
	logger.WithField("feature", "hierarchical").Info("启用分层路径 YYYY/MM/DD 结构")

	// 显示实际的分层路径结构
	logPath := logger.LogLinkFileFPath()
	dir := filepath.Dir(logPath)
	fmt.Printf("日志目录结构: %s\n", dir)

	// 列出分层路径
	filepath.Walk(settings.LogRootFPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(settings.LogRootFPath, path)
			fmt.Printf("日志文件: %s\n", relPath)
		}
		return nil
	})

	fmt.Println("注意：日志按 年/月/日 的层次结构存储\n")
}

// demonstrateAutoCleanup 演示自动清理功能
func demonstrateAutoCleanup() {
	fmt.Println("=== 4. 自动清理功能演示 ===")

	// 先创建一些旧的日志文件用于清理演示
	settings := logger.NewSettings()
	settings.LogRootFPath = "./logs/cleanup_demo"
	settings.LogNameBase = "cleanup_demo"
	settings.MaxAgeDays = 0 // 设置为0表示立即清理（仅用于演示）
	settings.RotationTime = 1 * time.Second
	settings.FormatterType = logger.FormatterTypeWithField

	logger.SetLoggerSettings(settings)

	fmt.Printf("日志保存时间设置为 %d 天（仅用于演示）\n", settings.MaxAgeDays)
	fmt.Printf("日志文件位置: %s\n", logger.LogLinkFileFPath())

	// 快速生成多个日志文件
	for i := 0; i < 5; i++ {
		logger.WithField("file", i+1).Info("清理测试日志")
		time.Sleep(2 * time.Second) // 等待轮转创建新文件
	}

	fmt.Println("生成的日志文件：")

	// 列出所有日志文件
	filepath.Walk(settings.LogRootFPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".log" {
			relPath, _ := filepath.Rel(settings.LogRootFPath, path)
			fmt.Printf("  %s (修改时间: %s)\n", relPath, info.ModTime().Format("15:04:05"))
		}
		return nil
	})

	fmt.Println("注意：超过保存时间的日志会被自动删除\n")
}

// demonstrateEasyFormatter 演示 Easy 格式器
func demonstrateEasyFormatter() {
	fmt.Println("=== 5. Easy 格式器演示 ===")

	settings := logger.NewSettings()
	settings.LogRootFPath = "./logs/easy_formatter"
	settings.LogNameBase = "easy_formatter_demo"
	settings.FormatterType = logger.FormatterTypeEasy
	settings.LogFormat = "%time% [%lvl%] => %msg% %fields%\n" // 自定义日志格式
	settings.TimestampFormat = "15:04:05.000"

	logger.SetLoggerSettings(settings)

	fmt.Printf("使用 Easy 格式器\n")
	fmt.Printf("自定义格式: %s\n", settings.LogFormat)
	fmt.Printf("日志文件位置: %s\n", logger.LogLinkFileFPath())

	// 测试 Easy 格式器
	logger.Info("Easy 格式器测试")
	logger.WithField("component", "auth").Info("认证模块启动")
	logger.WithFields(logrus.Fields{
		"user":   "admin",
		"status": "success",
		"ip":     "127.0.0.1",
	}).Info("用户登录成功")

	fmt.Println("注意：Easy 格式器支持自定义日志格式模板\n")
}