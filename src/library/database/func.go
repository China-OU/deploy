package database

import (
	"fmt"
	"initial"
	"strings"
	"library/common"
	"encoding/json"
	"errors"
	"models"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego"
	"time"
	"net/url"
	"library/cfunc"
)

type DBAgentOpr struct {
	AgentConf   models.CaasConf
	DeployComp  string
}

func (c *DBAgentOpr) GetAgentInfo() bool {
	if c.AgentConf.AgentIp == "" {
		conf, err := cfunc.GetAgentConfByComp(c.DeployComp)
		if err != nil {
			beego.Error(err.Error())
			return false
		}
		c.AgentConf = conf
	}
	return true
}

func (c *DBAgentOpr) PullGitDir(unit_dir, db_type, git_url, branch, commit_id string) (*map[string][]string, error) {
	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
	}
	query_str := args.Encode()
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/pull/file", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Post(url + "?" + query_str)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(300*time.Second, 300*time.Second)

	param := make(map[string]string)
	param["db_type"] = db_type
	param["unit_dir"] = unit_dir
	param["git_url"] = git_url
	param["git_branch"] = branch
	param["git_sha"] = commit_id
	data, _ := json.Marshal(param)
	req.Body(data)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	type StructRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data map[string][]string `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	if ret.Code != 1 {
		return nil, errors.New(ret.Msg)
	}
	return &ret.Data, nil
}

func (c *DBAgentOpr) FreshGitDirFunc(git_dir, db_type, commit_id string) (*map[string][]string, error) {
	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
	}
	query_str := args.Encode()
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/fresh/file", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Post(url + "?" + query_str)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(300*time.Second, 300*time.Second)

	param := make(map[string]string)
	param["db_type"] = db_type
	param["git_dir"] = git_dir
	param["git_sha"] = commit_id
	data, _ := json.Marshal(param)
	req.Body(data)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	type StructRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data map[string][]string `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	if ret.Code != 1 {
		return nil, errors.New(ret.Msg)
	}
	return &ret.Data, nil
}

func (c *DBAgentOpr) GetDDLSqlInfo(script_path string) (*string, error) {
	if strings.Contains(script_path, " ") {
		return nil, errors.New(script_path + "文件路径有空格，请修改！")
	}
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/script/detail", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(60*time.Second, 60*time.Second)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("path", script_path)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
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
		return nil, err
	}

	if ret.Code != 1 {
		return nil, errors.New(ret.Msg)
	}
	return &ret.Data, nil
}

func (c *DBAgentOpr) GetPkgSqlInfo(script_path, name string) (*map[string]string, error) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/script/pkgdetail", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(60*time.Second, 60*time.Second)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("path", script_path)
	req.Param("file_name", name)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	type StructRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data map[string]string `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	if ret.Code != 1 {
		return nil, errors.New(ret.Msg)
	}
	return &ret.Data, nil
}

func (c *DBAgentOpr) RmAgentDir(rm_dir string) error {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/rm/gitdir", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Post(url)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(60*time.Second, 60*time.Second)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("rm_dir", rm_dir)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	type StructRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
	}
	var ret StructRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	if ret.Code != 1 {
		return errors.New(ret.Msg)
	}
	return nil
}
