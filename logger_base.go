package logger

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}
func Printf(format string, args ...interface{}) {
	GetLogger().Printf(format, args...)
}
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}
func Warningf(format string, args ...interface{}) {
	GetLogger().Warningf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}

func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}
func Print(args ...interface{}) {
	GetLogger().Print(args...)
}
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}
func Warning(args ...interface{}) {
	GetLogger().Warning(args...)
}
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

func Debugln(args ...interface{}) {
	GetLogger().Debugln(args...)
}
func Infoln(args ...interface{}) {
	GetLogger().Infoln(args...)
}
func Println(args ...interface{}) {
	GetLogger().Println(args...)
}
func Warnln(args ...interface{}) {
	GetLogger().Warnln(args...)
}
func Warningln(args ...interface{}) {
	GetLogger().Warningln(args...)
}
func Errorln(args ...interface{}) {
	GetLogger().Errorln(args...)
}
func Fatalln(args ...interface{}) {
	GetLogger().Fatalln(args...)
}
func Panicln(args ...interface{}) {
	GetLogger().Panicln(args...)
}
