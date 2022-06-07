package cae

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/httplib"
    "library/common"
    "time"
)

func TransFile(srcFile, destPath string, destHost string, args ...map[string]interface{}) (error, []string) {
    err, token := getToken()
    if err != nil {
        beego.Error(err)
        return err, []string{"获取cae token失败！"}
    }

    // 获取远程执行命令 API 以及可变参数
    cv := CaeInitVar()
    timeout := 90
    retry_times := 3
    if len(args) > 0 {
        if _, ok := args[0]["timeout"]; ok {
            timeout = common.GetInt(args[0]["timeout"])
        }
        if _, ok := args[0]["retry_times"]; ok {
            retry_times = common.GetInt(args[0]["retry_times"])
        }
    }
    // 构建body参数
    data := make(map[string]interface{})
    localHost := common.GetLocalIp()
    data["src_host"] = localHost[0]
    data["src_file"] = srcFile
    data["src_user"] = "rhlog"
    data["dest_dir"] = destPath
    data["dest_host"] = destHost
    data["dest_user"] = "rhlog"
    dataBytes, _ := json.Marshal(data)

    flag := false
    var msg []string
    for i:=0; i<retry_times; i++ {
        req := httplib.Post(cv.BaseUrl + cv.FileTransApi)
        req.Header("Content-Type", "application/json")
        req.Header("X-Auth-Token", token)
        req.SetTimeout(time.Duration(timeout)*time.Second, time.Duration(timeout)*time.Second)
        req.Body(dataBytes)
        resp, err := req.Bytes()
        if err != nil {
            beego.Error(err)
            msg = append(msg, fmt.Sprintf("第%d次接口调用失败。错误为：%s", i+1, err.Error()))
            continue
        }

        // 解析返回结果
        var res map[string]interface{}
        err = json.Unmarshal(resp, &res)
        if err != nil {
            msg = append(msg, fmt.Sprintf("第%d次接口调用失败。CAE返回数据解析失败！", i + 1))
            continue
        }
        beego.Info(string(resp))
        if _, ok := res["err_msg"]; ok {
            msg = append(msg, common.GetString(res["err_msg"]))
            return errors.New("接口请求有误"), msg
        }
        ret_data, ok := res["data"].(map[string]interface{})
        if !ok {
            msg = append(msg, "CAE接口请求失败")
            return errors.New("CAE接口请求失败"), msg
        }
        if common.GetString(ret_data["msg"]) == "" && common.GetInt(ret_data["status"]) == 0 {
            msg = append(msg, fmt.Sprintf("文件传输成功。"))
            flag = true
            break
        } else {
            beego.Error(string(resp))
            msg = append(msg, fmt.Sprintf("第%d次接口解析失败。错误为：%s", i+1, common.GetString(ret_data["msg"])))
        }
    }
    if flag {
        return nil, msg
    }
    return errors.New("文件传输失败"), msg
}
