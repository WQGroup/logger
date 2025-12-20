package logger

import (
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

type YamlConfig struct {
	LogRoot             string `yaml:"log_root"`
	LogNameBase         string `yaml:"log_name_base"`
	Level               string `yaml:"level"`
	DaysToKeep          int    `yaml:"days_to_keep"`
	MaxSizeMB           int    `yaml:"max_size_mb"`
	UseHierarchicalPath bool   `yaml:"use_hierarchical_path"`

	// 新增的格式器配置字段
	FormatterType    string `yaml:"formatter_type"`
	TimestampFormat  string `yaml:"timestamp_format"`
	DisableTimestamp bool   `yaml:"disable_timestamp"`
	DisableLevel     bool   `yaml:"disable_level"`
	DisableCaller    bool   `yaml:"disable_caller"`
	FullTimestamp    bool   `yaml:"full_timestamp"`
	LogFormat        string `yaml:"log_format"`
}

func parseLevel(s string) logrus.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func LoadSettingsFromYAML(path string) (*Settings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg YamlConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	s := NewSettings()
	if cfg.LogRoot != "" {
		// 确保相对路径以 ./ 开头
		if len(cfg.LogRoot) > 0 && cfg.LogRoot[0] != '/' && cfg.LogRoot[0] != '.' && len(cfg.LogRoot) > 1 && cfg.LogRoot[1] != ':' {
			s.LogRootFPath = "./" + cfg.LogRoot
		} else {
			s.LogRootFPath = cfg.LogRoot
		}
	}
	if cfg.LogNameBase != "" {
		s.LogNameBase = cfg.LogNameBase
	}
	if cfg.Level != "" {
		s.Level = parseLevel(cfg.Level)
	}
	if cfg.DaysToKeep > 0 {
		s.MaxAgeDays = cfg.DaysToKeep
		s.MaxAge = time.Duration(s.MaxAgeDays*24) * time.Hour
	}
	if cfg.MaxSizeMB > 0 {
		s.MaxSizeMB = cfg.MaxSizeMB
	}
	s.UseHierarchicalPath = cfg.UseHierarchicalPath

	// 设置新的格式器配置字段
	if cfg.FormatterType != "" {
		s.FormatterType = cfg.FormatterType
	}
	if cfg.TimestampFormat != "" {
		s.TimestampFormat = cfg.TimestampFormat
	}
	s.DisableTimestamp = cfg.DisableTimestamp
	s.DisableLevel = cfg.DisableLevel
	s.DisableCaller = cfg.DisableCaller
	s.FullTimestamp = cfg.FullTimestamp
	if cfg.LogFormat != "" {
		s.LogFormat = cfg.LogFormat
	}

	return s, nil
}

func SetLoggerFromYAML(path string) error {
	s, err := LoadSettingsFromYAML(path)
	if err != nil {
		return err
	}
	SetLoggerSettings(s)
	return nil
}
