package common

import (
    "bytes"
    "errors"
    "io"
    "net/http"
    "os"
    "os/exec"
    "regexp"
    "time"
)

// 使用 curl 带认证下载文件
func DownloadWithCurl(url, localFile, user, pwd string) (err error) {
    if err := URLCheck(url); err != nil {
        return err
    }
    authStr := user + ":" + pwd
    cmd := exec.Command("curl", "-u", authStr, "-o", localFile, url)
    var outErr bytes.Buffer
    cmd.Stderr = &outErr
    if err := cmd.Run(); err != nil {
        return errors.New(err.Error() + " " + outErr.String())
    }
    if err := os.Chmod(localFile, 0644); err != nil {
        return err
    }
    return nil
}

// 通过 http 下载文件
func DownloadWithHttp(url, localFile string, timeout time.Duration, headers map[string]string) error {
    var validType = []string{"application/x-gzip", "application/zip", "application/octet-stream"}
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return err
    }
    for k, v := range headers {
        req.Header.Set(k, v)
    }
    if timeout == 0 {
        timeout = 10 * time.Minute
    }
    client := http.Client{
        Timeout: timeout,
    }
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    fileType := resp.Header.Get("Content-Type")
    if !InList(fileType, validType) {
        return errors.New("不支持的文件类型：" + fileType)
    }
    err = CreateAll(localFile)
    if err != nil {
        return err
    }
    file, err := os.OpenFile(localFile, os.O_RDWR, 0755)
    if err != nil {
        return err
    }
    defer file.Close()
    _, err = io.Copy(file, resp.Body)
    if err != nil {
        return err
    }
    return nil
}

// 检查 http 链接格式
func URLCheck(url string) (err error) {
    ipv4Pattern := `((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)`
    ipv6Pattern := `(([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|` +
        `(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|` +
        `(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|` +
        `(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
        `(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
        `(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
        `(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
        `(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))`
    ipPattern := "(" + ipv4Pattern + ")|(" + ipv6Pattern + ")"
    domainPattern := `[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}(\.[a-zA-Z0-9][a-zA-Z0-9_-]{0,62})*(\.[a-zA-Z][a-zA-Z0-9]{0,10}){1}`
    urlPattern := `((https|http|ftp|rtsp|mms)://)` + // 协议
        `(([0-9a-zA-Z]+:)?[0-9a-zA-Z_-]+@)?` + // pwd:user@
        "(" + ipPattern + "|(" + domainPattern + "))" + // IP 或域名
        `(:\d{1,5})?` + // 端口
        `(/+[a-zA-Z0-9][a-zA-Z0-9_.-]*)*/*` + // path
        `(\?([a-zA-Z0-9_-]+(=.*&?)*)*)*` // query
    re, err := regexp.Compile(urlPattern)
    if err != nil {
        return err
    }
    if ok := re.Match([]byte(url)); !ok {
        return errors.New("invalid url")
    }
    return nil
}