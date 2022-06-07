package mcp

import (
	"models"
	"fmt"
	"initial"
	"strings"
	"library/common"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"errors"
	"net/url"
)

type McpRancherOpr struct {
	AgentConf     models.CaasConf
	ProjectId        string
	StackId          string
	ServiceId        string
	Search           string
}

func (c *McpRancherOpr) GetRancherProjectList() (error, []RancherData) {
	project_url := "/v2-beta/projects"
	return c.GetSearchList(project_url)
}

func (c *McpRancherOpr) GetRancherStackList() (error, []RancherData) {
	stack_url := fmt.Sprintf("/v2-beta/projects/%s/stacks", c.ProjectId)
	return c.GetSearchList(stack_url)
}

func (c *McpRancherOpr) GetRancherServiceList() (error, []RancherData) {
	service_url := fmt.Sprintf("/v2-beta/projects/%s/services?name_prefix=%s&stackId=%s&limit=10&kind=service",
		c.ProjectId, c.Search, c.StackId)
	return c.GetSearchList(service_url)
}

func (c *McpRancherOpr) GetRancherInstanceList() (error, []RancherData) {
	service_url := fmt.Sprintf("/v2-beta/projects/%s/services/%s/instances", c.ProjectId, c.ServiceId)
	return c.GetSearchList(service_url)
}

// 返回查询信息，id和name的关键信息
func (c *McpRancherOpr) GetSearchList(search_url string) (error, []RancherData) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/rancher/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", search_url)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, []RancherData{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(string(ret_bytes))
		beego.Error(err.Error())
		return err, []RancherData{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), []RancherData{}
	}
	var data RancherRet
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(ret.Data)
		beego.Error(err.Error())
		return err, []RancherData{}
	}
	return nil, data.Data
}

// 操作部分
func (c *McpRancherOpr) GetRancherService() (error, RancherService) {
	service_url := fmt.Sprintf("/v2-beta/projects/%s/services/%s", c.ProjectId, c.ServiceId)
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/rancher/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", service_url)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, RancherService{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(string(ret_bytes))
		beego.Error(err.Error())
		return err, RancherService{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), RancherService{}
	}
	var data RancherService
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(ret.Data)
		beego.Error(err.Error())
		return err, RancherService{}
	}
	return nil, data
}

func (c *McpRancherOpr) ActionRancherService(action_url string, body_date_bytes []byte) (error, RancherService) {
	url_arr := strings.Split(action_url, "v2-beta")
	if len(url_arr) < 2 {
		return errors.New("url不正确，请确认接口是否正确"), RancherService{}
	}
	post_url := "/v2-beta" + url_arr[1]

	commont_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/rancher/post", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
		"post_url": []string{post_url},
	}
	req := httplib.Post(commont_url + "?" + args.Encode())
	req.Header("agent-auth", initial.AgentToken)
	req.Body(body_date_bytes)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, RancherService{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(string(ret_bytes))
		beego.Error(err.Error())
		return err, RancherService{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), RancherService{}
	}
	var data RancherService
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(ret.Data)
		beego.Error(err.Error())
		return err, RancherService{}
	}
	return nil, data
}