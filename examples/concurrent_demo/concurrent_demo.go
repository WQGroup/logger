package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/WQGroup/logger"
)

func main() {
	fmt.Println("=== 并发安全验证演示 ===\n")

	// 设置日志器
	settings := logger.NewSettings()
	settings.LogNameBase = "concurrent_demo"
	logger.SetLoggerSettings(settings)

	var wg sync.WaitGroup
	numGoroutines := 50
	numWrites := 100

	fmt.Printf("启动 %d 个 goroutine，每个写入 %d 条日志...\n", numGoroutines, numWrites)

	// 启动多个 goroutine 并发写入日志
	start := time.Now()
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numWrites; j++ {
				logger.WithField("goroutine", id).
					WithField("iteration", j).
					Info("并发日志测试")
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("完成！总共写入 %d 条日志，耗时 %v\n", numGoroutines*numWrites, duration)
	fmt.Printf("平均每秒写入 %.0f 条日志\n", float64(numGoroutines*numWrites)/duration.Seconds())

	// 测试并发设置 logger
	fmt.Println("\n测试并发设置 logger...")
	var configWg sync.WaitGroup
	configChanges := 10

	for i := 0; i < configChanges; i++ {
		configWg.Add(1)
		go func(id int) {
			defer configWg.Done()
			settings := logger.NewSettings()
			settings.LogNameBase = fmt.Sprintf("config_test_%d", id)
			logger.SetLoggerSettings(settings)
			logger.WithField("config_id", id).Info("配置更改测试")
		}(i)
	}

	configWg.Wait()
	fmt.Println("并发配置更改测试完成")

	fmt.Printf("\n日志文件位置: %s\n", logger.LogLinkFileFPath())
}