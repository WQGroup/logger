package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/WQGroup/logger"
)

func main() {
	fmt.Println("=== Windows GUI 模式检测验证 ===\n")

	// 显示当前系统信息
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("架构: %s\n", runtime.GOARCH)

	// 测试不同的日志器设置
	settings := logger.NewSettings()
	settings.LogNameBase = "gui_test"

	// 测试基本日志功能
	logger.SetLoggerSettings(settings)

	fmt.Println("\n测试基本日志功能:")
	logger.Info("GUI 模式测试开始")
	logger.WithField("platform", runtime.GOOS).Info("平台信息")

	// 测试不同日志级别
	logger.Debug("这是 Debug 级别的日志")
	logger.Info("这是 Info 级别的日志")
	logger.Warn("这是 Warning 级别的日志")
	logger.Error("这是 Error 级别的日志")

	// 测试格式器
	fmt.Println("\n测试格式器功能:")
	jsonSettings := logger.NewSettings()
	jsonSettings.FormatterType = logger.FormatterTypeJSON
	jsonSettings.LogNameBase = "gui_test_json"
	logger.SetLoggerSettings(jsonSettings)

	logger.WithField("test_type", "gui_mode").Info("JSON 格式测试")

	// 测试 YAML 配置加载
	fmt.Println("\n测试 YAML 配置加载:")
	yamlContent := `
log_root: "./logs"
log_name_base: "gui_test_yaml"
level: "info"
days_to_keep: 7
formatter_type: "withField"
`

	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "gui_test_config_*.yaml")
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

	logger.WithField("config_source", "yaml").Info("YAML 配置测试成功")

	fmt.Printf("\n测试完成！日志文件位置: %s\n", logger.LogLinkFileFPath())
	fmt.Printf("当前日志文件: %s\n", logger.CurrentFileName())
}