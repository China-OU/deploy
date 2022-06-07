package database

import (
	"models"
	"library/cfunc"
	"strings"
	"library/common"
	"net/url"
	"fmt"
	"initial"
	"time"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

// 请求agent执行相关sql语句命令

func OracleExecSql(conf models.UnitConfDb, log models.OnlineDbLog) (string, error) {
	agent, err := cfunc.GetAgentConfByComp(conf.DeployComp)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
	}
	query_str := args.Encode()
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/oracle/exec-sql", agent.AgentIp, agent.AgentPort)
	req := httplib.Post(url + "?" + query_str)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(30*time.Minute, 30*time.Minute)

	param := make(map[string]string)
	param["file_path"] = log.FilePath
	param["db_name"] = conf.Dbname
	param["db_host"] = conf.Host
	param["proxy_user"] = log.ProxyUser
	data, _ := json.Marshal(param)
	req.Body(data)
	ret_bytes, err := req.Bytes()
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
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	if ret.Code != 1 {
		return ret.Msg, errors.New("执行失败")
	}
	return ret.Msg, nil
}

// sql_name 为 agent对应的 src/file/oracle/下面 aa.sql文件，此时 sql_name = 'aa.sql'
func OracleCommonQuery(conf models.UnitConfDb, sql_name string) (string, error) {
	agent, err := cfunc.GetAgentConfByComp(conf.DeployComp)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
	}
	query_str := args.Encode()
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/oracle/common/query", agent.AgentIp, agent.AgentPort)
	req := httplib.Post(url + "?" + query_str)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(1*time.Minute, 1*time.Minute)

	param := make(map[string]string)
	param["file_path"] = sql_name
	param["db_name"] = conf.Dbname
	param["db_host"] = conf.Host
	data, _ := json.Marshal(param)
	req.Body(data)
	ret_bytes, err := req.Bytes()
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
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	if ret.Code != 1 {
		return ret.Msg, errors.New("执行失败")
	}
	return ret.Msg, nil
}