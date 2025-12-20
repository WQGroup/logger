package main

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/WQGroup/logger"
)

// MyFormatter 自定义格式器
type MyFormatter struct{}

func (f *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	level := entry.Level.String()
	msg := entry.Message

	// 构建字段字符串
	var fields string
	if len(entry.Data) > 0 {
		fieldParts := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		fields = " [" + strings.Join(fieldParts, ", ") + "]"
	}

	return []byte(fmt.Sprintf("%s | %-5s | %s%s\n", timestamp, level, msg, fields)), nil
}

func main() {
	fmt.Println("=== Logger 格式器演示 ===\n")

	// 1. 使用默认的 withField 格式器
	fmt.Println("1. 默认 withField 格式器:")
	logger.SetLoggerName("DemoApp")
	logger.Info("应用程序启动")
	logger.WithField("user_id", 12345).Info("用户登录")
	logger.WithFields(logrus.Fields{
		"module": "payment",
		"action": "transfer",
		"amount": 100.50,
	}).Info("转账操作完成")
	fmt.Println()

	// 2. 测试 JSON 格式器
	fmt.Println("2. JSON 格式器:")
	settings := logger.NewSettings()
	settings.FormatterType = logger.FormatterTypeJSON
	settings.LogNameBase = "demo_json"
	logger.SetLoggerSettings(settings)

	logger.Info("JSON 格式的日志")
	logger.WithField("request_id", "req-123").Info("处理请求")
	fmt.Println()

	// 3. 测试自定义格式器
	fmt.Println("3. 自定义格式器:")
	settings = logger.NewSettings()
	settings.CustomFormatter = &MyFormatter{}
	settings.LogNameBase = "demo_custom"
	logger.SetLoggerSettings(settings)

	logger.Info("自定义格式的日志")
	logger.WithField("component", "database").Info("连接成功")
	logger.WithFields(logrus.Fields{
		"query": "SELECT * FROM users",
		"rows":  100,
	}).Info("查询执行")
	fmt.Println()

	// 4. 测试时间戳格式自定义
	fmt.Println("4. 自定义时间戳格式:")
	settings = logger.NewSettings()
	settings.FormatterType = logger.FormatterTypeWithField
	settings.TimestampFormat = "2006-01-02T15:04:05.000Z"
	settings.LogNameBase = "demo_timestamp"
	logger.SetLoggerSettings(settings)

	logger.Info("自定义时间戳格式的日志")
	logger.WithField("event", "startup").Info("服务启动完成")
	fmt.Println()

	fmt.Printf("日志文件位置: %s\n", logger.LogLinkFileFPath())
	fmt.Printf("当前日志文件: %s\n", logger.CurrentFileName())
}