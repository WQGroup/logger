package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func init() {
	logLevel = logrus.InfoLevel
	logOnlyMsg = false
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

func SetLoggerOnlyMsg(onlyMsg bool) {
	logOnlyMsg = onlyMsg
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
	loggerBase = NewLogHelper(logRootDirFPath, logNameBase, logLevel,
		time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)
}

func NewLogHelper(logRootDirFPath, appName string, level logrus.Level, maxAge time.Duration, rotationTime time.Duration) *logrus.Logger {

	outputFormatNow := outputFormat
	if logOnlyMsg == true {
		// only msg
		outputFormatNow = outputFormatOnlyMsg
	}

	Logger := &logrus.Logger{
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       outputFormatNow,
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

	loggerLinkFileFPath = filepath.Join(pathRoot, appName+".log")
	rotateLogsWriter, _ = rotatelogs.New(
		filepath.Join(pathRoot, appName+"--%Y%m%d%H%M--.log"),
		rotatelogs.WithLinkName(loggerLinkFileFPath),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(rotationTime),
	)

	Logger.SetLevel(level)
	Logger.SetOutput(io.MultiWriter(os.Stderr, rotateLogsWriter))

	return Logger
}

func NewLogger(logRootDirFPath, logFileName string) *logrus.Logger {

	var err error
	nowLogger := logrus.New()
	outputFormatNow := outputFormat
	if logOnlyMsg == true {
		// only msg
		outputFormatNow = outputFormatOnlyMsg
	}
	nowLogger.Formatter = &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       outputFormatNow,
	}
	pathRoot := logRootDirFPath
	// create dir if not exists
	if _, err := os.Stat(pathRoot); os.IsNotExist(err) {
		err = os.MkdirAll(pathRoot, os.ModePerm)
		if err != nil {
			panic(errors.New(fmt.Sprintf("Create log dir error: %s", err.Error())))
		}
	}
	fileName := fmt.Sprintf("%v.log", logFileName)
	fileAbsPath := filepath.Join(pathRoot, fileName)

	onceLoggerFile, err := os.OpenFile(fileAbsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Create log file error: %s", err.Error())))
	}
	nowLogger.SetOutput(onceLoggerFile)

	return nowLogger
}

func LogLinkFileFPath() string {
	return loggerLinkFileFPath
}

func CurrentFileName() string {

	if rotateLogsWriter == nil {
		return ""
	}
	return rotateLogsWriter.CurrentFileName()
}

const (
	NameDef             = "logger"
	logRootFPathDef     = "."
	outputFormat        = "%time% - [%lvl%]: %msg%\n"
	outputFormatOnlyMsg = "%msg%\n"
)

var (
	logOnlyMsg          bool
	logLevel            logrus.Level
	logNameBase         = NameDef
	logRootDirFPath     = logRootFPathDef
	loggerBase          *logrus.Logger
	loggerLinkFileFPath = ""
	rotateLogsWriter    *rotatelogs.RotateLogs
)
