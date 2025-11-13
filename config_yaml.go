package logger

import (
    "io/ioutil"
    "strings"
    "time"

    "github.com/sirupsen/logrus"
    yaml "gopkg.in/yaml.v3"
)

type YamlConfig struct {
    LogRoot     string `yaml:"log_root"`
    LogNameBase string `yaml:"log_name_base"`
    Level       string `yaml:"level"`
    DaysToKeep  int    `yaml:"days_to_keep"`
    MaxSizeMB   int    `yaml:"max_size_mb"`
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
    b, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg YamlConfig
    if err := yaml.Unmarshal(b, &cfg); err != nil {
        return nil, err
    }
    s := NewSettings()
    if cfg.LogRoot != "" {
        s.LogRootFPath = cfg.LogRoot
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
