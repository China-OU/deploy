package host

import (
    "bufio"
    "crypto/md5"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/thedevsaddam/gojsonq"
    "hash"
    "io"
    "library/cae"
    "os"
    "strings"
)

type FileHash struct {
    Host    string  `json:"host"`
    File    string  `json:"file"`
    Method  string  `json:"method"`
    Hash    string  `json:"hash"`
}

// 支持多个IP
// filepath 必须传入绝对路径
// method：sha256,md5，默认sha256
func RemoteFileHash(hosts, filepath, method string) ([]FileHash, error) {
    var hashData []FileHash
    var hash FileHash
    hash.File = filepath
    if !strings.HasPrefix(filepath, "/") {
        return nil, errors.New("文件位置必须是绝对路径")
    }
    var hashCmd string
    switch method {
    case "sha256":
        hashCmd = fmt.Sprintf("sha256sum %s", filepath)
        hash.Method = "sha256"
    case "md5":
        hashCmd = fmt.Sprintf("md5sum %s", filepath)
        hash.Method = "md5"
    default:
        hashCmd = fmt.Sprintf("sha256sum %s", filepath)
        hash.Method = "sha256"
    }
    err, _, out := cae.ExecCmdOutput(hashCmd, "/tmp", "root", hosts)
    if err != nil {
        return nil, err
    }
    // 解析输出
    outBytes, err := json.Marshal(out)
    if err != nil {
        return nil, err
    }
    outJSON := gojsonq.New(gojsonq.SetSeparator("_")).FromString(string(outBytes))
    for _, h := range strings.Split(hosts, ";") {
        hash.Host = h
        status := fmt.Sprintf("%s", outJSON.Find(h + "_status"))
        outJSON.Reset()
        stdout := fmt.Sprintf("%s", outJSON.Find(h + "_return"))
        if status != "ok" {
            hash.Hash = ""
        } else {
            hash.Hash = strings.Split(stdout, " ")[0]
        }
        outJSON.Reset()
        hashData = append(hashData, hash)
    }
    return hashData, nil
}

// 本地文件hash计算
func LocalFileHash(filepath, method string) (FileHash, error) {
    var fileHash FileHash
    var h hash.Hash

    fileHash.Host = "localhost"
    fileHash.File = filepath
    fileHash.Method = method

    f, err := os.Open(filepath)
    if err != nil {
        return fileHash, err
    }
    defer f.Close()

    buf := bufio.NewReader(f)
    switch method {
    case "sha256":
        h = sha256.New()
    case "md5":
        h = md5.New()
    default:
        h = sha256.New()
        fileHash.Method = "sha256"
    }

    _, err = io.Copy(h, buf)
    if err != nil {
        return fileHash, err
    }
    fileHash.Hash = hex.EncodeToString(h.Sum(nil))
    return fileHash, nil
}