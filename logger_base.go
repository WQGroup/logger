package logger

import (
	"fmt"
	"os"
	"github.com/sirupsen/logrus"
)

// getLoggerInternal 获取当前日志器，内部使用，避免自动初始化
// 为保持向后兼容性，此函数忽略错误，在初始化失败时会创建默认日志器
func getLoggerInternal() *logrus.Logger {
	logger, err := getLoggerInternalWithError()
	if err != nil {
		// 如果初始化失败，创建一个默认的日志器并打印错误到 stderr
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		logger = logrus.New()
	}
	return logger
}

// getLoggerInternalWithError 获取当前日志器，返回错误信息
func getLoggerInternalWithError() (*logrus.Logger, error) {
	// 快速路径：无锁读取
	if logger := loggerBase; logger != nil {
		return logger, nil
	}

	// 慢速路径：需要初始化
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// 双重检查
	if loggerBase == nil {
		settings := NewSettings()
		logger, err := NewLogHelperWithError(settings)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize logger: %w", err)
		}
		loggerBase = logger
	}

	return loggerBase, nil
}

func Debugf(format string, args ...interface{}) {
	getLoggerInternal().Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	getLoggerInternal().Infof(format, args...)
}
func Printf(format string, args ...interface{}) {
	getLoggerInternal().Printf(format, args...)
}
func Warnf(format string, args ...interface{}) {
	getLoggerInternal().Warnf(format, args...)
}
func Warningf(format string, args ...interface{}) {
	getLoggerInternal().Warningf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	getLoggerInternal().Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	getLoggerInternal().Fatalf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	getLoggerInternal().Panicf(format, args...)
}

func Debug(args ...interface{}) {
	getLoggerInternal().Debug(args...)
}
func Info(args ...interface{}) {
	getLoggerInternal().Info(args...)
}
func Print(args ...interface{}) {
	getLoggerInternal().Print(args...)
}
func Warn(args ...interface{}) {
	getLoggerInternal().Warn(args...)
}
func Warning(args ...interface{}) {
	getLoggerInternal().Warning(args...)
}
func Error(args ...interface{}) {
	getLoggerInternal().Error(args...)
}
func Fatal(args ...interface{}) {
	getLoggerInternal().Fatal(args...)
}
func Panic(args ...interface{}) {
	getLoggerInternal().Panic(args...)
}

func Debugln(args ...interface{}) {
	getLoggerInternal().Debugln(args...)
}
func Infoln(args ...interface{}) {
	getLoggerInternal().Infoln(args...)
}
func Println(args ...interface{}) {
	getLoggerInternal().Println(args...)
}
func Warnln(args ...interface{}) {
	getLoggerInternal().Warnln(args...)
}
func Warningln(args ...interface{}) {
	getLoggerInternal().Warningln(args...)
}
func Errorln(args ...interface{}) {
	getLoggerInternal().Errorln(args...)
}
func Fatalln(args ...interface{}) {
	getLoggerInternal().Fatalln(args...)
}
func Panicln(args ...interface{}) {
	getLoggerInternal().Panicln(args...)
}

// WithField 返回一个带有单个字段的 logrus.Entry
func WithField(key string, value interface{}) *logrus.Entry {
	return getLoggerInternal().WithField(key, value)
}

// WithFields 返回一个带有多个字段的 logrus.Entry
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return getLoggerInternal().WithFields(fields)
}

// SetLoggerName 设置日志名称（向后兼容）
func SetLoggerName(name string) {
	settings := NewSettings()
	settings.LogNameBase = name
	SetLoggerSettings(settings)
}
