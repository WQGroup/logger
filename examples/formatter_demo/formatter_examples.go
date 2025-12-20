package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/WQGroup/logger"
)

// ExampleWithFieldFormatter 演示 WithField 格式器的使用
func ExampleWithFieldFormatter() {
	fmt.Println("=== WithField Formatter Example ===")

	settings := logger.NewSettings()
	settings.FormatterType = logger.FormatterTypeWithField
	settings.LogRootFPath = os.TempDir()
	settings.LogNameBase = "withfield_example"

	logger.SetLoggerSettings(settings)

	// 基本日志
	logger.Info("应用程序启动")

	// 使用单个字段
	logger.WithField("user_id", 12345).Info("用户登录")

	// 使用多个字段
	logger.WithFields(logrus.Fields{
		"module":  "auth",
		"action":  "login",
		"user":    "john",
		"ip":      "192.168.1.1",
		"status":  "success",
	}).Info("用户认证成功")

	// 支持中文
	logger.WithFields(logrus.Fields{
		"操作":    "事件广播",
		"类型":    "实时通知",
		"结果":    "成功",
	}).Info("【实时通知】事件广播成功 operation=(a+b)-c result=123.45")

	fmt.Println("日志文件路径:", logger.LogLinkFileFPath())
	fmt.Println()
}

// ExampleJSONFormatter 演示 JSON 格式器的使用
func ExampleJSONFormatter() {
	fmt.Println("=== JSON Formatter Example ===")

	settings := logger.NewSettings()
	settings.FormatterType = logger.FormatterTypeJSON
	settings.LogRootFPath = os.TempDir()
	settings.LogNameBase = "json_example"

	logger.SetLoggerSettings(settings)

	logger.Info("JSON 格式日志")
	logger.WithField("request_id", "req-123").Info("处理请求")
	logger.WithFields(logrus.Fields{
		"method": "GET",
		"path":   "/api/users",
		"status": 200,
	}).Info("HTTP 请求")

	fmt.Println("日志文件路径:", logger.LogLinkFileFPath())
	fmt.Println()
}

// ExampleTextFormatter 演示 Text 格式器的使用
func ExampleTextFormatter() {
	fmt.Println("=== Text Formatter Example ===")

	settings := logger.NewSettings()
	settings.FormatterType = logger.FormatterTypeText
	settings.LogRootFPath = os.TempDir()
	settings.LogNameBase = "text_example"

	logger.SetLoggerSettings(settings)

	logger.Info("Text 格式日志")
	logger.WithField("service", "payment").Info("支付服务启动")
	logger.WithFields(logrus.Fields{
		"order_id": "ORD-12345",
		"amount":   99.99,
	}).Info("订单创建")

	fmt.Println("日志文件路径:", logger.LogLinkFileFPath())
	fmt.Println()
}

// CustomFormatter 实现自定义格式器
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 简单的自定义格式：[时间]级别: 消息 (key1=value1 key2=value2)
	timeStr := entry.Time.Format("15:04:05")
	msg := fmt.Sprintf("[%s] %s: %s", timeStr, entry.Level, entry.Message)

	// 添加字段
	if len(entry.Data) > 0 {
		var fields []string
		for k, v := range entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		msg += fmt.Sprintf(" (%s)", fmt.Sprintf("%s", fields))
	}

	return []byte(msg + "\n"), nil
}

// ExampleCustomFormatter 演示自定义格式器的使用
func ExampleCustomFormatter() {
	fmt.Println("=== Custom Formatter Example ===")

	settings := logger.NewSettings()
	settings.CustomFormatter = &CustomFormatter{}
	settings.LogRootFPath = os.TempDir()
	settings.LogNameBase = "custom_example"

	logger.SetLoggerSettings(settings)

	logger.Info("自定义格式日志")
	logger.WithField("app", "myapp").Info("应用程序启动")
	logger.WithFields(logrus.Fields{
		"version": "1.0.0",
		"env":     "production",
	}).Info("环境信息")

	fmt.Println("日志文件路径:", logger.LogLinkFileFPath())
	fmt.Println()
}

// ExampleFormatterFromYAML 演示从 YAML 配置格式器
func ExampleFormatterFromYAML() {
	fmt.Println("=== YAML Configuration Example ===")

	yamlContent := `
log_root: "./logs"
log_name_base: "yaml_example"
level: "debug"
days_to_keep: 7

# 格式器配置
formatter_type: "withField"
timestamp_format: "2006-01-02T15:04:05.000Z"
disable_level: false
disable_caller: true
`

	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "logger_config_*.yaml")
	if err != nil {
		fmt.Printf("创建临时文件失败: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		fmt.Printf("写入配置文件失败: %v\n", err)
		return
	}
	tmpFile.Close()

	// 从 YAML 加载配置
	if err := logger.SetLoggerFromYAML(tmpFile.Name()); err != nil {
		fmt.Printf("加载 YAML 配置失败: %v\n", err)
		return
	}

	logger.Info("从 YAML 配置加载的日志器")
	logger.WithField("config", "yaml").Info("配置加载成功")

	fmt.Println("日志文件路径:", logger.LogLinkFileFPath())
	fmt.Println()
}

// main 运行所有格式器示例
func main() {
	ExampleWithFieldFormatter()
	ExampleJSONFormatter()
	ExampleTextFormatter()
	ExampleCustomFormatter()
	ExampleFormatterFromYAML()
}