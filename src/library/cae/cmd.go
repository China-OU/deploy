package cae

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/httplib"
    "library/common"
    "os"
    "strings"
    "time"
)

// cmd 执行命令   path 远程执行路径   user  执行用户  host 远程ip组，可支持多个，如127.0.0.1;127.0.0.2
func ExecCmd(cmd, path, user, host string, args ...map[string]interface{}) (error, []string) {
    err, token := getToken()
    if err != nil {
        beego.Error(err)
        return err, []string{"获取cae token失败！"}
    }
    // 获取远程执行命令 API 以及可变参数
    cv := CaeInitVar()
    timeout := 150
    retry_times := 3
    if len(args) > 0 {
        if _, ok := args[0]["timeout"]; ok {
            timeout = common.GetInt(args[0]["timeout"])
        }
        if _, ok := args[0]["retry_times"]; ok {
            retry_times = common.GetInt(args[0]["retry_times"])
        }
    }
    hosts := strings.Split(host, ";")
    reqData := make(map[string]interface{})
    reqData["dst_ips"] = hosts
    reqData["remote_comand"] = fmt.Sprintf("/usr/bin/env bash -c \"%s\"", cmd)
    reqData["remote_path"] = path
    reqData["run_user"] = user
    reqData["timeout"] = timeout
    reqData["is_async"] = false
    reqDataBytes, _ := json.Marshal(reqData)

    var msg []string
    // 重试times次
    flag := false
    success_msg := ""
    info_msg := ""
    for i:=0; i<retry_times; i++ {
        beego.Info(fmt.Sprintf("远程执行命令：%s，第%d次尝试", cmd, i+1))
        req:= httplib.Post(cv.BaseUrl + cv.CmdApi)
        req.Header("Content-Type", "application/json")
        req.Header("X-Auth-Token", token)
        req.SetTimeout(time.Duration(timeout)*time.Second, time.Duration(timeout)*time.Second)
        req.Body(reqDataBytes)
        resp, err := req.Bytes()
        if err != nil {
            beego.Error(err)
            info_msg = string(resp)
            msg = append(msg, fmt.Sprintf("第%d次接口调用失败。错误为：%s", i+1, err.Error()))
            continue
        }

        // 解析请求结果
        var res map[string]interface{}
        err = json.Unmarshal(resp, &res)
        if err != nil {
            beego.Error(msg)
            info_msg = string(resp)
            msg = append(msg, fmt.Sprintf("第%d次接口调用失败。CAE返回数据解析失败！", i + 1))
            continue
        }
        if _, ok := res["err_msg"]; ok {
           msg = append(msg, common.GetString(res["err_msg"]))
           beego.Info(string(resp))
           return errors.New("接口请求有误"), msg
        }
        resData, ok := res["data"].([]interface{})
        if !ok {
            msg = append(msg, "CAE接口请求失败")
            beego.Error("CAE接口请求失败: ", res["data"])
            beego.Info(string(resp))
            return errors.New("CAE接口请求失败"), msg
        }
        result_flag := true
        out:
        for _, v := range resData {
            vv := v.(map[string]interface{})
            if _, ok := vv["prod"]; ok {
                ret_prod := vv["prod"].(map[string]interface{})
                for key, value := range ret_prod {
                    true_msg := value.(map[string]interface{})
                    if common.GetString(true_msg["status"]) != "ok" {
                        info_msg = string(resp)
                        msg = append(msg, fmt.Sprintf("%s执行命令失败%s", key, true_msg["return"]))
                        result_flag = false
                        break out
                    }
                }
            }
            // 测试环境，结果不定，采用层层反射
            if _, ok := vv["test"]; ok {
                ret_prod := vv["test"].(map[string]interface{})
                for key, value := range ret_prod {
                    true_msg := value.(map[string]interface{})
                    if common.GetString(true_msg["status"]) != "ok" {
                        info_msg = string(resp)
                        msg = append(msg, fmt.Sprintf("%s执行命令失败%s", key, true_msg["return"]))
                        result_flag = false
                        break out
                    }
                }
            }
        }
        if result_flag == true {
            flag = true
            success_msg = string(resp)
            break
        }
    }
    if flag {
        msg = append(msg, fmt.Sprintf("命令执行成功。 命令为：%s", cmd))
        msg = append(msg, fmt.Sprintf("执行结果为：%s", success_msg))
        return nil, msg
    }
    beego.Info(info_msg)
    return errors.New("命令执行失败"), msg
}

func GetLocalShellPath() string {
    pwd, _ := os.Getwd()
    return pwd + "/conf/file/"
}

// 截取 ExecCmd() 输出的日志信息
func TruncCaeOut(out []string, length int) string {
    if len(out) != 2 {
        return ""
    }
    stdout := out[1]
    if len(stdout) < length + 100 {
        return stdout
    } else {
        return stdout[len(stdout) - length:len(stdout) - 6]
    }
}