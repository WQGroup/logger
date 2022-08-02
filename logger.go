package logger

import (
	"errors"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
	"os"
	"path/filepath"
	"time"
)

func init() {
	logLevel = logrus.InfoLevel
	logInit()
}

func GetLogger() *logrus.Logger {
	return loggerBase
}

func SetLoggerLevel(level logrus.Level) {
	logLevel = level
}

func SetLoggerRootDir(loggerRootDir string) {
	logRootDirFPath = loggerRootDir
	// re init
	logInit()
}

func SetLoggerName(logName string) {
	if logName == "" {
		panic("Logger name is empty")
	}
	logNameBase = logName
	// re init
	logInit()
}

func Debugf(format string, args ...interface{}) {
	loggerBase.Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	loggerBase.Infof(format, args...)
}
func Printf(format string, args ...interface{}) {
	loggerBase.Printf(format, args...)
}
func Warnf(format string, args ...interface{}) {
	loggerBase.Warnf(format, args...)
}
func Warningf(format string, args ...interface{}) {
	loggerBase.Warningf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	loggerBase.Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	loggerBase.Fatalf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	loggerBase.Panicf(format, args...)
}

func Debug(args ...interface{}) {
	loggerBase.Debug(args...)
}
func Info(args ...interface{}) {
	loggerBase.Info(args...)
}
func Print(args ...interface{}) {
	loggerBase.Print(args...)
}
func Warn(args ...interface{}) {
	loggerBase.Warn(args...)
}
func Warning(args ...interface{}) {
	loggerBase.Warning(args...)
}
func Error(args ...interface{}) {
	loggerBase.Error(args...)
}
func Fatal(args ...interface{}) {
	loggerBase.Fatal(args...)
}
func Panic(args ...interface{}) {
	loggerBase.Panic(args...)
}

func Debugln(args ...interface{}) {
	loggerBase.Debugln(args...)
}
func Infoln(args ...interface{}) {
	loggerBase.Infoln(args...)
}
func Println(args ...interface{}) {
	loggerBase.Println(args...)
}
func Warnln(args ...interface{}) {
	loggerBase.Warnln(args...)
}
func Warningln(args ...interface{}) {
	loggerBase.Warningln(args...)
}
func Errorln(args ...interface{}) {
	loggerBase.Errorln(args...)
}
func Fatalln(args ...interface{}) {
	loggerBase.Fatalln(args...)
}
func Panicln(args ...interface{}) {
	loggerBase.Panicln(args...)
}

func logInit() {
	if logNameBase == "" {
		// 默认不设置的时候就是这个
		logNameBase = NameDef
	}
	loggerBase = NewLogHelper(logRootDirFPath, logNameBase, logLevel, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)
}

func NewLogHelper(logRootDirFPath, appName string, level logrus.Level, maxAge time.Duration, rotationTime time.Duration) *logrus.Logger {

	Logger := &logrus.Logger{
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}
	pathRoot := filepath.Join(logRootDirFPath, "Logs")
	// create dir if not exists
	if _, err := os.Stat(pathRoot); os.IsNotExist(err) {
		err = os.MkdirAll(pathRoot, os.ModePerm)
		if err != nil {
			panic(errors.New(fmt.Sprintf("Create log dir error: %s", err.Error())))
		}
	}

	fileAbsPath := filepath.Join(pathRoot, appName+".log")

	writer, _ := rotatelogs.New(
		filepath.Join(pathRoot, appName+"--%Y%m%d%H%M--.log"),
		rotatelogs.WithLinkName(fileAbsPath),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(rotationTime),
	)

	Logger.SetLevel(level)
	Logger.SetOutput(io.MultiWriter(os.Stderr, writer))

	return Logger
}

const (
	NameDef         = "logger"
	logRootFPathDef = "."
)

var (
	logLevel        logrus.Level
	logNameBase     = NameDef
	logRootDirFPath = logRootFPathDef
	loggerBase      *logrus.Logger
)
