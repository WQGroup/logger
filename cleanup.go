package logger

import (
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "strconv"
    "time"
)

func CleanupExpiredLogs(root string, days int) error {
    if days <= 0 {
        return nil
    }
    return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil
        }
        if !d.IsDir() {
            return nil
        }
        rel, err := filepath.Rel(root, p)
        if err != nil {
            return nil
        }
        parts := filepath.SplitList(rel)
        if len(parts) == 0 {
            return nil
        }
        segs := splitPath(rel)
        if len(segs) == 3 {
            y, _ := strconv.Atoi(segs[0])
            m, _ := strconv.Atoi(segs[1])
            dday, _ := strconv.Atoi(segs[2])
            t := time.Date(y, time.Month(m), dday, 0, 0, 0, 0, time.Local)
            if time.Since(t) > time.Duration(days*24)*time.Hour {
                _ = os.RemoveAll(p)
                pm := filepath.Dir(p)
                if isEmpty(pm) {
                    _ = os.Remove(pm)
                    py := filepath.Dir(pm)
                    if isEmpty(py) {
                        _ = os.Remove(py)
                    }
                }
            }
        }
        return nil
    })
}

func splitPath(p string) []string {
    var out []string
    for {
        dir, file := filepath.Split(p)
        if file != "" {
            out = append([]string{file}, out...)
        }
        if dir == "" || dir == "." || dir == "/" {
            break
        }
        p = filepath.Clean(dir[:len(dir)-1])
    }
    return out
}

func isEmpty(p string) bool {
    f, err := os.Open(p)
    if err != nil {
        return false
    }
    defer f.Close()
    _, err = f.Readdirnames(1)
    return err == io.EOF
}
