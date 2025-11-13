package logger

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

func TestYAMLAndSizeRotationPath(t *testing.T) {
    root, err := os.MkdirTemp("", "logger-ut")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(root)
    yamlPath := filepath.Join(root, "logger.yaml")
    yaml := strings.Join([]string{
        "log_root: '" + filepath.ToSlash(root) + "'",
        "log_name_base: 'ut'",
        "level: 'debug'",
        "days_to_keep: 7",
        "max_size_mb: 1",
        "",
    }, "\n")
    if err := ioutil.WriteFile(yamlPath, []byte(yaml), 0644); err != nil {
        t.Fatal(err)
    }
    if err := SetLoggerFromYAML(yamlPath); err != nil {
        t.Fatal(err)
    }
    for i := 0; i < 5000; i++ {
        Infof("%s", strings.Repeat("x", 300))
    }
    now := time.Now()
    dayDir := filepath.Join(root, now.Format("2006"), now.Format("01"), now.Format("02"))
    if _, err := os.Stat(dayDir); err != nil {
        t.Fatalf("day dir not exists: %s", dayDir)
    }
    expected := filepath.Join(dayDir, "ut.log")
    if LogLinkFileFPath() != expected {
        t.Fatalf("unexpected link path: %s", LogLinkFileFPath())
    }
}

func TestTimeRotationPath(t *testing.T) {
    root, err := os.MkdirTemp("", "logger-ut-time")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(root)
    s := NewSettings()
    s.LogRootFPath = root
    s.LogNameBase = "ut2"
    s.MaxSizeMB = 0
    SetLoggerSettings(s)
    Info("hello")
    now := time.Now()
    dayDir := filepath.Join(root, now.Format("2006"), now.Format("01"), now.Format("02"))
    cur := CurrentFileName()
    if !strings.HasPrefix(cur, dayDir+string(os.PathSeparator)) {
        t.Fatalf("current file not in day dir: %s", cur)
    }
}

func TestCleanupExpired(t *testing.T) {
    root, err := os.MkdirTemp("", "logger-ut-clean")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(root)
    old := time.Now().AddDate(0, 0, -30)
    y := old.Format("2006")
    m := old.Format("01")
    d := old.Format("02")
    dayDir := filepath.Join(root, y, m, d)
    if err := os.MkdirAll(dayDir, 0755); err != nil {
        t.Fatal(err)
    }
    f := filepath.Join(dayDir, "x.log")
    if err := ioutil.WriteFile(f, []byte("x"), 0644); err != nil {
        t.Fatal(err)
    }
    if err := CleanupExpiredLogs(root, 7); err != nil {
        t.Fatal(err)
    }
    if _, err := os.Stat(dayDir); err == nil {
        t.Fatalf("expired day dir still exists: %s", dayDir)
    }
    monthDir := filepath.Join(root, y, m)
    if _, err := os.Stat(monthDir); err == nil {
        t.Fatalf("empty month dir still exists: %s", monthDir)
    }
    yearDir := filepath.Join(root, y)
    if _, err := os.Stat(yearDir); err == nil {
        t.Fatalf("empty year dir still exists: %s", yearDir)
    }
}
