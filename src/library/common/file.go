package common

import (
    "errors"
    "net/http"
    "os"
    "path"
    "strings"
)

// 根据文件路径，创建文件以及所有父级目录
// 如果路径不以 '/' 结尾，则认为最后是文件
// 如果路径以 '/' 结尾，则认为最后是一个目录，且路径下不带文件
func CreateAll(fullPath string) error {
    fullPath = strings.TrimSpace(fullPath)
    if fullPath == "" {
        return errors.New("路径为空")
    }
    if fullPath[:len(fullPath) - 1] == "/" {
        // 解析出最短路径
        // https://pkg.go.dev/path?tab=doc#Clean
        fullPath = path.Clean(fullPath)

        switch fullPath {
        case "/":
            return errors.New("输入路径为 '/'")
        case ".":
            return errors.New("输入路径为 '.'")
        }

        err := os.MkdirAll(fullPath, 0755)
        if err != nil {
            return err
        }
    } else {
        fullPath = path.Clean(fullPath)

        switch fullPath {
        case "/":
            return errors.New("输入路径为 '/'")
        case ".":
            return errors.New("输入路径为 '.'")
        }

        parents := strings.Split(fullPath, "/")
        parentPath := strings.Join(parents[:len(parents) - 1], "/")
        err := os.MkdirAll(parentPath, 0755)
        if err != nil {
            return err
        }
        _, err = os.Create(fullPath)
        if err != nil {
            return err
        }
    }
    return nil
}

// 检测文件的 MIME 类型，例如 .jar = zip，.war = zip
// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Basics_of_HTTP/MIME_types
// https://tools.ietf.org/html/rfc6838
func FileType(filePath string) (string, string, error) {
    buffer := make([]byte, 512)
    file, err := os.Open(filePath)
    if err != nil {
        return "", "", err
    }
    defer file.Close()
    _, err = file.Read(buffer)
    if err != nil {
        return "", "", err
    }
    filetype, subtype := splitContentType(http.DetectContentType(buffer))
    return filetype, subtype, nil
}

func splitContentType(s string) (string, string) {
    x := strings.Split(s, ";")
    y := strings.Split(x[0], "/")
    if len(y) > 1 {
        return y[0], y[1]
    }
    return y[0], ""
}