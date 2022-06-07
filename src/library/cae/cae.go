package cae

import (
    "encoding/json"
    "errors"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/httplib"
    "initial"
    "library/common"
    "time"
)

type loginResult struct {
    Status int
    Data   struct {
        Sessionid string
    }
}
// 保存和返回常理
type CaeVar struct {
    CaeUser        string `json:"cae_user"`
    CaePwd         string `json:"cae_pwd"`
    BaseUrl        string `json:"base_url"`
    LoginApi       string `json:"login_api"`
    LogoutApi      string `json:"logout_api"`
    CmdApi         string `json:"cmd_api"`
    FileTransApi   string `json:"file_trans_api"`
    FileCopyApi    string `json:"file_copy_api"`
}

func CaeInitVar() CaeVar {
    user_name := "deployplatform"
    user_pwd := "5706c7e6214a4c8dcb7200247a8e5e41743a6a4f6c5f35cf03086ee41d1375e9"
    return CaeVar{
        CaeUser: user_name,
        CaePwd: common.AesDecrypt(user_pwd),
        BaseUrl: beego.AppConfig.String("cae_baseurl"),
        LoginApi: "/login",
        LogoutApi: "/logout",
        CmdApi: "/v1/routes_web/files/command",
        FileTransApi: "/v2/apis/remote/file/transfer",
        FileCopyApi: "/v2/apis/remote/file/copy",
    }
}

func login() (error, string) {
    cv := CaeInitVar()
    loginData := map[string]string{
        "username": cv.CaeUser,
        "password": cv.CaePwd,
    }
    loginDataBytes, err := json.Marshal(loginData)
    if err != nil {
        beego.Error(err)
        return err, "获取 cae 用户失败！"
    }

    req := httplib.Post(cv.BaseUrl + cv.LoginApi)
    req.Header("Content-Type", "application/json")
    req.Body(loginDataBytes)
    resp, err := req.Bytes()
    if err != nil {
        beego.Error(err)
        return err, "cae 登陆请求失败！"
    }
    var loginResData loginResult
    if err = json.Unmarshal(resp, &loginResData); err != nil {
        beego.Error(err)
        return err, "cae 登陆失败！"
    }
    // token 写入缓存
    key := "cae_login_token"
    if initial.GetCache.IsExist(key) {
        if err := initial.GetCache.Delete(key); err != nil {
            beego.Error(err)
            return err, "cae 登陆缓存清理失败！"
        }
    }
    _ = initial.GetCache.Put(key, loginResData.Data.Sessionid, 30*time.Minute)
    return nil, "cae 登陆成功！"
}

func getToken() (error, string) {
    // get token from cache
    key := "cae_login_token"
    if !initial.GetCache.IsExist(key) || common.GetString(initial.GetCache.Get(key)) == "" {
        if err, msg := login(); err != nil {
            beego.Error(err)
            return err, msg
        }
    }
    token := common.GetString(initial.GetCache.Get(key))
    if token == "" {
        beego.Error("token is empty")
        return errors.New("token is empty"), "获取 cae token 失败！"
    }
    return nil, token
}

// 退出
func Logout(token string) error{
    cv := CaeInitVar()
    request := map[string]string{
        "username":cv.CaeUser,
    }
    bytes, _ := json.Marshal(request)
    req:= httplib.Post(cv.BaseUrl + cv.LogoutApi)
    req.Header("Content-Type", "application/json")
    req.Header("X-Auth-Token", token)
    req.Body(bytes)
    _, err := req.String()
    if err != nil {
        return err
    }
    beego.Info("cae退出登录成功！")
    return nil
}