package logger

import (
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "time"
)

func CleanupExpiredLogs(root string, days int) error {
    if days <= 0 {
        return nil
    }

    // 处理分层路径结构 (YYYY/MM/DD)
    err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
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

        segs := splitPath(rel)
        // 检查是否是分层路径结构 (YYYY/MM/DD)
        if len(segs) == 3 {
            y, err1 := strconv.Atoi(segs[0])
            m, err2 := strconv.Atoi(segs[1])
            dday, err3 := strconv.Atoi(segs[2])

            // 只有当所有部分都是有效数字时才处理
            if err1 == nil && err2 == nil && err3 == nil {
                t := time.Date(y, time.Month(m), dday, 0, 0, 0, 0, time.Local)
                if time.Since(t) > time.Duration(days*24)*time.Hour {
                    _ = os.RemoveAll(p)
                    // 清理空的父目录
                    cleanupEmptyParents(p, root)
                }
            }
        }
        return nil
    })

    if err != nil {
        return err
    }

    // 处理扁平路径结构的文件
    files, err := os.ReadDir(root)
    if err != nil {
        return err
    }

    // 匹配旧格式的日志文件: basename--YYYYMMDDHHMM--.log
    re := regexp.MustCompile(`^(.+)--(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})--\.log$`)

    for _, file := range files {
        if file.IsDir() {
            continue
        }

        matches := re.FindStringSubmatch(file.Name())
        if len(matches) == 7 {
            year, _ := strconv.Atoi(matches[2])
            month, _ := strconv.Atoi(matches[3])
            day, _ := strconv.Atoi(matches[4])
            hour, _ := strconv.Atoi(matches[5])
            minute, _ := strconv.Atoi(matches[6])

            t := time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local)
            if time.Since(t) > time.Duration(days*24)*time.Hour {
                _ = os.Remove(filepath.Join(root, file.Name()))
            }
        }
    }

    return nil
}

// cleanupEmptyParents 递归清理空的父目录
func cleanupEmptyParents(dirPath, root string) {
    pm := filepath.Dir(dirPath)
    if pm == root || pm == dirPath {
        return
    }

    if isEmpty(pm) {
        _ = os.Remove(pm)
        cleanupEmptyParents(pm, root)
    }
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
