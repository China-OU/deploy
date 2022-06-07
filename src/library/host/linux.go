package host

import (
    "errors"
    "fmt"
    "library/cae"
    "strconv"
    "strings"
    "time"
)

// 主机应用端口检查
// IP + 协议 + 端口可以标识一个唯一的进程
// 执行失败状态为0，执行成功状态为1，校验失败状态为2
func PortDetect(ip string, protocols, ports []string) (err error, stat int, msg []string) {
    var protStr string
    for _, p := range protocols {
        p = strings.ToLower(p)
        if p != "tcp" && p != "udp" {
            return errors.New("传输协议错误！"), 2, append(msg, "端口检查失败！")
        }
        p6 := p + "6"
        protStr += p + "|" + p6 + "|"
    }
    protStr = protStr[:len(protStr) - 1]

    var portStr string
    for _, port := range ports {
        p, err := strconv.Atoi(port)
        if err != nil || p < 1 || p > 65535 {
            return errors.New("端口值错误！"), 2, append(msg, "端口检查失败！")
        }
        portStr += port + "|"
    }
    portStr = portStr[:len(portStr) - 1]

    cmd := fmt.Sprintf(`netstat -an -t -u | grep -w -E '%s' | awk '{print $4}'| grep -w -E '%s'`, protStr, portStr)
    err, logs := cae.ExecCmd(cmd, "/tmp", "root", ip)
    if err != nil {
        return err, 0, logs
    }

    return nil, 1, logs
}

// 检测模式仅支持 "up" 和 "down" 两个参数
func CheckRunning(ip, mode string, count, interval int, protocols, ports []string) (stat int, log []string, err error) {
    var protocolStr string
    var portStr string
    for _, protocol := range protocols {
        protocol = strings.ToLower(protocol)
        if protocol != "tcp" && protocol != "udp" {
            log = append(log, "端口协议错误，当前仅支持tcp和udp")
            return 2, log, errors.New("端口协议错误！")
        }
        protocol6 := protocol + "6"
        protocolStr += protocol + "|" + protocol6 + "|"
    }
    protocolStr = protocolStr[:len(protocolStr) - 1]

    for _, port := range ports {
        p, err := strconv.Atoi(port)
        if err != nil || p < 1 || p > 65535 {
            log = append(log, "端口值错误，只允许1-65535以内的整数")
            return 2, log, errors.New("端口值错误！")
        }
        portStr += port + "|"
    }
    portStr = portStr[:len(portStr) - 1]
    cmd := fmt.Sprintf(`netstat -an -t -u | grep -w -E '%s'  | awk '{print $4}'| grep -w -E '%s'`, protocolStr, portStr)

    if interval < 2 || interval > 30 {
        log = append(log, "检测间隔只允许2-30之间的整数")
        return 2, log, errors.New("检测间隔参数错误！")
    }
    if count < 1 || count > 30 {
        log = append(log, "检测次数只允许1-30之间的整数")
        return 2, log, errors.New("检测次数参数错误！")
    }
    flag := false
    for i := 0; i < count; i++ {
        err, _ := cae.ExecCmd(cmd, "/tmp", "root", ip)

        // up 模式
        if mode == "up" {
            if err != nil {
                log = append(log, ip + ": 端口未在使用")
                time.Sleep(time.Duration(interval) * time.Second)
            } else {
                log = append(log, ip + ": 端口正在使用中")
                flag = true
                break
            }
        }

        // down 模式
        if mode == "down" {
            if err != nil {
                log = append(log, ip + ": 端口未在使用")
                flag = true
                break
            } else {
                log = append(log, ip + ": 端口正在使用中")
                time.Sleep(time.Duration(interval) * time.Second)
            }
        }
    }
    if flag == false {
        return 0, log, nil
    }
    return 1, log, nil
}

// 检测工作目录是否为合法目录，防止误操作对主机造成影响
// 部署操作，只允许在 /app/ 和 /tmp/ 目录下进行操作
func IsWSValid(path string) bool {
    if strings.HasPrefix(path, "/app/") {
        return true
    }
    if strings.HasPrefix(path, "/tmp/") {
        return true
    }
    return false
}

// 删除路径最后的 /
func PathWithoutSlash(path string) string {
    if strings.TrimSpace(path) == "" {
        return ""
    }
    if path == "/" {
        return path
    }
    if path[len(path) - 1:] == "/" {
        path = path[:len(path) - 1]
    } else {
        return path
    }
    return PathWithoutSlash(path)
}

// 给路径最后添加 /
func PathWithSlash(path string) string {
    if strings.TrimSpace(path) == "" {
        return ""
    }
    if path == "/" {
        return path
    }
    if path[len(path) - 1:] == "/" {
        return path
    }
    return path + "/"
}