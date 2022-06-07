package database

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/httplib"
    "initial"
    "library/cfunc"
    "library/common"
    "models"
    "net/url"
    "strings"
    "time"
)

func PgsqlExecSQL(conf models.UnitConfDb, log models.OnlineDbLog) (string, error) {
    agent, err := cfunc.GetAgentConfByComp(conf.DeployComp)
    if err != nil {
        beego.Error(err.Error())
        return "", err
    }
    ipList := strings.Join(common.GetLocalIp(), ",")
    args := url.Values{
        "ip_list": []string{ipList},
    }
    query := args.Encode()
    url := fmt.Sprintf("http://%s:%s/agent/v1/db/pgsql/exec-sql?%s", agent.AgentIp, agent.AgentPort, query)
    req := httplib.Post(url)
    req.Header("agent-auth", initial.AgentToken)
    req.SetTimeout(30*time.Minute, 30*time.Minute)

    param := make(map[string]string)
    param["file_path"] = log.FilePath
    param["db_name"] = conf.Dbname
    param["db_host"] = conf.Host
    data, err := json.Marshal(param)
    if err != nil {
        beego.Error(err)
        return "", err
    }
    req.Body(data)
    retBytes, err := req.Bytes()
    if err != nil {
        beego.Error(err.Error())
        return "", err
    }
    type StructRet struct {
        Code int `json:"code"`
        Msg string `json:"msg"`
        Data string `json:"data"`
    }
    var ret StructRet
    err = json.Unmarshal(retBytes, &ret)
    if err != nil {
        beego.Error(err.Error())
        return "", err
    }

    if ret.Code != 1 {
        return ret.Msg, errors.New("执行失败")
    }
    return ret.Msg, nil
}
