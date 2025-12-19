package logger

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

	// 使用格式器工厂创建格式器
	factory := &FormatterFactory{}
	formatter := factory.CreateFormatter(settings)

	Logger := &logrus.Logger{
		Formatter: formatter,
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
	// 格式器类型常量
	FormatterTypeWithField = "withField"
	FormatterTypeEasy      = "easy"
	FormatterTypeJSON      = "json"
	FormatterTypeText      = "text"
)

type Settings struct {
	OnlyMsg             bool          // 废弃：仅输出消息，不包含时间戳等额外信息（向后兼容，内部映射到 FormatterType）
	Level               logrus.Level  // 日志级别
	LogRootFPath        string        // 日志根目录
	LogNameBase         string        // 日志名称
	RotationTime        time.Duration // 日志轮转时间
	MaxAge              time.Duration // 日志最大保存时间
	MaxAgeDays          int
	MaxSizeMB           int
	UseHierarchicalPath bool          // 是否使用分层路径（YYYY/MM/DD）

	// 新增的格式器配置字段
	FormatterType       string            // 格式器类型："withField", "easy", "json", "text"
	TimestampFormat     string            // 时间戳格式（默认 "2006-01-02 15:04:05.000"）
	CustomFormatter     logrus.Formatter  // 用户自定义格式器
	DisableTimestamp    bool              // 是否禁用时间戳
	DisableLevel        bool              // 是否禁用日志级别
	DisableCaller       bool              // 是否禁用调用者信息
	FullTimestamp       bool              // 是否显示完整时间戳
	LogFormat           string            // 自定义日志格式（用于 easy-formatter）
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

		// 新增字段的默认值
		FormatterType:       FormatterTypeWithField, // 默认使用 withField 格式器
		TimestampFormat:     "2006-01-02 15:04:05.000",
		CustomFormatter:     nil,
		DisableTimestamp:    false,
		DisableLevel:        false,
		DisableCaller:       true, // 默认不显示调用者信息，保持简洁
		FullTimestamp:       false,
		LogFormat:           "",
	}
}

var (
	loggerBase          *logrus.Logger         // 日志基础记录器
	rotateLogsWriter    *rotatelogs.RotateLogs // 日志轮转记录器
	currentLogFileFPath string                 // 当前日志文件路径
)

// WithFieldFormatter 自定义日志格式器，支持结构化字段输出
// 输出格式：2025-12-18 18:32:07.379 - [INFO]: 【实时通知】事件广播成功 operation=(a+b)-c result=123.45
type WithFieldFormatter struct {
	TimestampFormat  string // 时间戳格式
	DisableTimestamp bool   // 是否禁用时间戳
	DisableLevel     bool   // 是否禁用日志级别
	DisableCaller    bool   // 是否禁用调用者信息
}

// Format 实现 logrus.Formatter 接口
func (f *WithFieldFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// 添加时间戳
	if !f.DisableTimestamp {
		timestamp := entry.Time.Format(f.TimestampFormat)
		b.WriteString(timestamp)
		b.WriteString(" - ")
	}

	// 添加日志级别
	if !f.DisableLevel {
		b.WriteString("[")
		b.WriteString(strings.ToUpper(entry.Level.String()))
		b.WriteString("]: ")
	}

	// 添加调用者信息
	if !f.DisableCaller && entry.HasCaller() {
		b.WriteString(fmt.Sprintf("%s:%d - ", entry.Caller.File, entry.Caller.Line))
	}

	// 添加消息
	b.WriteString(entry.Message)

	// 如果有字段，将它们附加到消息后面
	if len(entry.Data) > 0 {
		var fields []string
		for k, v := range entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		if len(fields) > 0 {
			b.WriteString(" ")
			b.WriteString(strings.Join(fields, " "))
		}
	}

	b.WriteString("\n")
	return b.Bytes(), nil
}

// FormatterFactory 格式器工厂
type FormatterFactory struct{}

// CreateFormatter 根据设置创建相应的格式器
func (f *FormatterFactory) CreateFormatter(settings *Settings) logrus.Formatter {
	// 优先使用自定义格式器
	if settings.CustomFormatter != nil {
		return settings.CustomFormatter
	}

	// 处理向后兼容：如果设置了 OnlyMsg，则使用 easy-formatter
	formatterType := settings.FormatterType
	if settings.OnlyMsg {
		formatterType = FormatterTypeEasy
	}

	// 如果没有设置格式器类型，默认使用 withField
	if formatterType == "" {
		formatterType = FormatterTypeWithField
	}

	// 根据 FormatterType 创建格式器
	switch formatterType {
	case FormatterTypeJSON:
		return &logrus.JSONFormatter{
			TimestampFormat:  settings.TimestampFormat,
			DisableTimestamp: settings.DisableTimestamp,
		}
	case FormatterTypeText:
		return &logrus.TextFormatter{
			TimestampFormat:  settings.TimestampFormat,
			DisableTimestamp: settings.DisableTimestamp,
			DisableColors:    true,
			FullTimestamp:    settings.FullTimestamp,
		}
	case FormatterTypeEasy:
		// 向后兼容 OnlyMsg
		logFormat := settings.LogFormat
		if logFormat == "" {
			if settings.OnlyMsg {
				logFormat = outputFormatOnlyMsg
			} else {
				logFormat = outputFormat
			}
		}
		return &easy.Formatter{
			TimestampFormat: settings.TimestampFormat,
			LogFormat:       logFormat,
		}
	case FormatterTypeWithField:
		fallthrough
	default:
		// 使用 WithFieldFormatter
		return &WithFieldFormatter{
			TimestampFormat:  settings.TimestampFormat,
			DisableTimestamp: settings.DisableTimestamp,
			DisableLevel:     settings.DisableLevel,
			DisableCaller:    settings.DisableCaller,
		}
	}
}

// SetCustomFormatter 设置用户自定义格式器
func SetCustomFormatter(formatter logrus.Formatter) {
	if loggerBase == nil {
		SetLoggerSettings(NewSettings())
	}
	settings := NewSettings()
	settings.CustomFormatter = formatter
	SetLoggerSettings(settings)
}

// isWindowsGUI 检测程序是否以Windows GUI模式运行
func isWindowsGUI() bool {
	// 尝试获取标准输出句柄，如果失败则可能是GUI模式
	_, err := os.Stderr.Stat()
	return err != nil
}
