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

// copy函数可以指定属主，用于小文件传输；transfer只能用rhlog，用于传输大文件。
func CopyFile(srcFile, destPath string, destHosts []string) (error, string) {
    err, token := getToken()
    if err != nil {
        beego.Error(err)
        return err, "获取 cae token 失败！"
    }
    // 获取远程执行命令 API 以及可变参数
    cv := CaeInitVar()
    data := make(map[string]interface{})
    data["dest_dir"] = destPath
    data["dest_hosts"] = destHosts
    data["owner"] = "mwop"
    data["group"] = "mwop"
    data["is_asyn"] = false
    localHost := common.GetLocalIp()
    data["src_host"] = localHost[0]
    data["src_file"] = srcFile
    dataBytes, _ := json.Marshal(data)

    req := httplib.Post(cv.BaseUrl + cv.FileCopyApi)
    req.Header("Content-Type", "application/json")
    req.Header("X-Auth-Token", token)
    req.SetTimeout(300*time.Second, 300*time.Second)
    req.Body(dataBytes)
    resp, err := req.Bytes()
    if err != nil {
        beego.Error(err)
        return err, "cae 文件复制接口请求失败！"
    }

    // 解析返回结果
    var logs string
    var res map[string]interface{}
    if err := json.Unmarshal(resp, &res); err != nil {
        beego.Error(err)
        return err, "返回数据解析失败！"
    }
    
    errMsg, ok := res["err_msg"].(map[string]interface{})
    if ok {
        logs += fmt.Sprintf("CAE接口请求失败: \n")
        for k, v := range errMsg {
            logs += k + fmt.Sprintf(" %s\n", v)
        }
        return errors.New("CAE接口请求失败"), logs
    }
    resData, ok := res["data"].(map[string]interface{})
    if !ok {
        beego.Error("CAE 接口数据不正确：res[\"data\"] = ", res["data"])
        return errors.New("CAE接口请求失败"), "CAE 接口请求失败，请联系管理员"
    }
    resDataMsg, ok := resData["msg"].(map[string]interface{})
    if !ok {
        beego.Error("CAE 接口数据不正确：resData[\"msg\"] = ", resData["msg"])
        return errors.New("CAE接口请求失败"), "CAE 接口请求失败，请联系管理员"
    }
    resDataMsgF, ok := resDataMsg["failed"].(map[string]interface{})
    if !ok {
        beego.Error("CAE 接口数据不正确：resDataMsg[\"failed\"] = ", resDataMsg["failed"])
        return errors.New("CAE接口请求失败"), "CAE 接口请求失败，请联系管理员"
    }
    resDataMsgS, ok := resDataMsg["success"].(map[string]interface{})
    if !ok {
        beego.Error("CAE 接口数据不正确：resDataMsg[\"success\"] = ", resDataMsg["success"])
        return errors.New("CAE接口请求失败"), "CAE 接口请求失败，请联系管理员"
    }
    if resDataMsgF != nil {
        for k, v := range resDataMsgF {
            logs += k + fmt.Sprintf(" %s", v)
        }
    }
    if resDataMsgS != nil {
        for k, v := range resDataMsgS {
            logs += k + fmt.Sprintf(" %s\n", v)
        }
    }
    return nil, logs
}