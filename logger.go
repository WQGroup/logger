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
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
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

	if settings.MaxAgeDays > 0 {
		settings.MaxAge = time.Duration(settings.MaxAgeDays*24) * time.Hour
	}
	if settings.MaxAge <= 0 {
		settings.MaxAge = time.Duration(7*24) * time.Hour
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
			TimestampFormat: "2006-01-02 15:04:05.000",
			LogFormat:       outputFormatNow,
		},
	}

	pathRoot := filepath.Join(settings.LogRootFPath, "Logs")
	if settings.LogRootFPath != logRootFPathDef {
		pathRoot = settings.LogRootFPath
	}
	if _, err = os.Stat(pathRoot); os.IsNotExist(err) {
		err = os.MkdirAll(pathRoot, os.ModePerm)
		if err != nil {
			panic(errors.New(fmt.Sprintf("Create log dir error: %s", err.Error())))
		}
	}

	var fileWriter io.Writer
	if settings.MaxSizeMB > 0 {
		// 大小轮转模式
		var logDir string
		if settings.UseHierarchicalPath {
			// 新格式：按年/月/日分层
			now := time.Now()
			yearDir := filepath.Join(pathRoot, now.Format("2006"))
			monthDir := filepath.Join(yearDir, now.Format("01"))
			dayDir := filepath.Join(monthDir, now.Format("02"))
			logDir = dayDir
		} else {
			// 旧格式：扁平结构
			logDir = pathRoot
		}

		if _, err = os.Stat(logDir); os.IsNotExist(err) {
			err = os.MkdirAll(logDir, os.ModePerm)
			if err != nil {
				panic(errors.New(fmt.Sprintf("Create log dir error: %s", err.Error())))
			}
		}

		currentLogFileFPath = filepath.Join(logDir, settings.LogNameBase+".log")
		fileWriter = &lumberjack.Logger{
			Filename:  currentLogFileFPath,
			MaxSize:   settings.MaxSizeMB,
			MaxAge:    settings.MaxAgeDays,
			LocalTime: true,
			Compress:  false,
		}
		rotateLogsWriter = nil
	} else {
		// 时间轮转模式
		var logPattern string
		if settings.UseHierarchicalPath {
			// 新格式：按年/月/日分层
			logPattern = filepath.Join(pathRoot, "%Y", "%m", "%d", settings.LogNameBase+"--%H%M--.log")
		} else {
			// 旧格式：扁平结构
			logPattern = filepath.Join(pathRoot, settings.LogNameBase+"--%Y%m%d%H%M--.log")
		}

		rotateLogsWriter, err = rotatelogs.New(
			logPattern,
			rotatelogs.WithMaxAge(settings.MaxAge),
			rotatelogs.WithRotationTime(settings.RotationTime),
		)
		if err != nil {
			panic(errors.New(fmt.Sprintf("Create log file error: %s", err.Error())))
		}
		fileWriter = rotateLogsWriter
		// 使用 rotatelogs 提供的当前文件名
		currentLogFileFPath = rotateLogsWriter.CurrentFileName()
	}

	Logger.SetLevel(settings.Level)
	// 在Windows下，如果使用-H=windowsgui编译，os.Stderr将无效，所以需要特殊处理
	if isWindowsGUI() {
		Logger.SetOutput(fileWriter)
	} else {
		Logger.SetOutput(io.MultiWriter(os.Stderr, fileWriter))
	}

	_ = CleanupExpiredLogs(pathRoot, settings.MaxAgeDays)

	return Logger
}

// LogLinkFileFPath returns the path of the current log file.
func LogLinkFileFPath() string {
	return currentLogFileFPath
}

// CurrentFileName 当前日志文件名
func CurrentFileName() string {

	if rotateLogsWriter != nil {
		return rotateLogsWriter.CurrentFileName()
	}
	return currentLogFileFPath
}

const (
	NameDef             = "logger"
	logRootFPathDef     = "."
	outputFormat        = "%time% - [%lvl%]: %msg%\n"
	outputFormatOnlyMsg = "%msg%\n"
)

type Settings struct {
	OnlyMsg             bool          // 是否只输出消息内容
	Level               logrus.Level  // 日志级别
	LogRootFPath        string        // 日志根目录
	LogNameBase         string        // 日志名称
	RotationTime        time.Duration // 日志轮转时间
	MaxAge              time.Duration // 日志最大保存时间
	MaxAgeDays          int
	MaxSizeMB           int
	UseHierarchicalPath bool          // 是否使用分层路径（YYYY/MM/DD）
}

// NewSettings 创建一个新的日志设置
func NewSettings() *Settings {
	return &Settings{
		OnlyMsg:             false,
		Level:               logrus.InfoLevel,
		LogRootFPath:        logRootFPathDef,
		LogNameBase:         NameDef,
		RotationTime:        time.Duration(24) * time.Hour, // 默认每天轮转一次
		MaxAge:              time.Duration(7*24) * time.Hour,
		MaxAgeDays:          7,
		MaxSizeMB:           0,
		UseHierarchicalPath: false, // 默认使用旧格式，保持向后兼容
	}
}

var (
	loggerBase          *logrus.Logger         // 日志基础记录器
	rotateLogsWriter    *rotatelogs.RotateLogs // 日志轮转记录器
	currentLogFileFPath string                 // 当前日志文件路径
)

// isWindowsGUI 检测程序是否以Windows GUI模式运行
func isWindowsGUI() bool {
	// 尝试获取标准输出句柄，如果失败则可能是GUI模式
	_, err := os.Stderr.Stat()
	return err != nil
}
