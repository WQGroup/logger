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

func GetLogger() *logrus.Logger {

	if loggerBase == nil {
		// 如果没有设置日志记录器，则使用默认设置
		SetLoggerSettings(NewSettings())
	}
	return loggerBase
}

func SetLoggerSettings(inSettings ...*Settings) {

	var settings *Settings
	if len(inSettings) > 0 {
		settings = inSettings[0]
	} else {
		settings = NewSettings()
	}

	if settings.LogRootFPath == "" {
		settings.LogRootFPath = logRootFPathDef
	}

	if settings.LogNameBase == "" {
		settings.LogNameBase = NameDef
	}

	if settings.RotationTime <= 0 {
		settings.RotationTime = time.Duration(24) * time.Hour // 默认每天轮转一次
	}

	if settings.MaxAge <= 0 {
		settings.MaxAge = time.Duration(7*24) * time.Hour // 默认保存7天
	}

	loggerBase = NewLogHelper(settings)
}

func NewLogHelper(settings *Settings) *logrus.Logger {

	var err error
	outputFormatNow := outputFormat
	if settings.OnlyMsg == true {
		// only msg
		outputFormatNow = outputFormatOnlyMsg
	}

	Logger := &logrus.Logger{
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       outputFormatNow,
		},
	}

	// 默认使用当前目录下的 Logs 目录
	pathRoot := filepath.Join(settings.LogRootFPath, "Logs")
	if settings.LogRootFPath != logRootFPathDef {
		// 如果设置了日志根目录，则使用该目录
		pathRoot = settings.LogRootFPath
	}
	// create dir if not exists
	if _, err = os.Stat(pathRoot); os.IsNotExist(err) {
		err = os.MkdirAll(pathRoot, os.ModePerm)
		if err != nil {
			panic(errors.New(fmt.Sprintf("Create log dir error: %s", err.Error())))
		}
	}
	loggerLinkFileFPath = filepath.Join(pathRoot, settings.LogNameBase+".log")
	rotateLogsWriter, err = rotatelogs.New(
		filepath.Join(pathRoot, settings.LogNameBase+"--%Y%m%d%H%M--.log"),
		rotatelogs.WithLinkName(loggerLinkFileFPath),
		rotatelogs.WithMaxAge(settings.MaxAge),
		rotatelogs.WithRotationTime(settings.RotationTime),
	)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Create log file error: %s", err.Error())))
	}

	Logger.SetLevel(settings.Level)
	// 在Windows下，如果使用-H=windowsgui编译，os.Stderr将无效，所以需要特殊处理
	if isWindowsGUI() {
		Logger.SetOutput(rotateLogsWriter)
	} else {
		Logger.SetOutput(io.MultiWriter(os.Stderr, rotateLogsWriter))
	}

	return Logger
}

// LogLinkFileFPath returns the path of the log file that is linked to the current log writer.
func LogLinkFileFPath() string {
	return loggerLinkFileFPath
}

// CurrentFileName 当前日志文件名
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

type Settings struct {
	OnlyMsg      bool          // 是否只输出消息内容
	Level        logrus.Level  // 日志级别
	LogRootFPath string        // 日志根目录
	LogNameBase  string        // 日志名称
	RotationTime time.Duration // 日志轮转时间
	MaxAge       time.Duration // 日志最大保存时间
}

// NewSettings 创建一个新的日志设置
func NewSettings() *Settings {
	return &Settings{
		OnlyMsg:      false,
		Level:        logrus.InfoLevel,
		LogRootFPath: logRootFPathDef,
		LogNameBase:  NameDef,
		RotationTime: time.Duration(24) * time.Hour,   // 默认每天轮转一次
		MaxAge:       time.Duration(7*24) * time.Hour, // 默认保存7天
	}
}

var (
	loggerLinkFileFPath = ""                   // 日志链接文件路径
	loggerBase          *logrus.Logger         // 日志基础记录器
	rotateLogsWriter    *rotatelogs.RotateLogs // 日志轮转记录器
)

// isWindowsGUI 检测程序是否以Windows GUI模式运行
func isWindowsGUI() bool {
	// 尝试获取标准输出句柄，如果失败则可能是GUI模式
	_, err := os.Stderr.Stat()
	return err != nil
}
