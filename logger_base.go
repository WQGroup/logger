package logger

import "github.com/sirupsen/logrus"

// getLoggerInternal 获取当前日志器，内部使用，避免自动初始化
func getLoggerInternal() *logrus.Logger {
	loggerMutex.RLock()
	logger := loggerBase
	loggerMutex.RUnlock()

	// 如果 loggerBase 为 nil，返回 GetLogger() 的结果来触发自动初始化
	if logger == nil {
		return GetLogger()
	}
	return logger
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
