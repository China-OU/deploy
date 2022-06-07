package mcp

import (
	"models"
	"fmt"
	"initial"
	"strings"
	"library/common"
	"encoding/json"
	"github.com/astaxie/beego"
	"time"
	"github.com/astaxie/beego/httplib"
	"errors"
	"net/url"
)

type McpCaasOpr struct {
	AgentConf     models.CaasConf
	TeamId        string
	ClustUuid     string
	StackName     string
	ServiceName   string
}

func (c *McpCaasOpr) GetCaasTeamList() (error, []TeamDataDetail) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", "/v1/teamlist?requestPage=1&pageSize=10")
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, []TeamDataDetail{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, []TeamDataDetail{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), []TeamDataDetail{}
	}
	var data TeamList
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err, []TeamDataDetail{}
	}
	return nil, data.Data.Data
}

func (c *McpCaasOpr) GetCaasClustList() (error, []ClustData) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", "/v1/envidlist")
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, []ClustData{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, []ClustData{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), []ClustData{}
	}
	var data CaasClustList
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err, []ClustData{}
	}
	return nil, data.Data
}

func (c *McpCaasOpr) GetCaasStackList() (error, []StackDataDetail) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	page := 1
	var stack_list []StackDataDetail

	for {
		req := httplib.Get(url)
		req.Header("agent-auth", initial.AgentToken)
		ip_list := strings.Join(common.GetLocalIp(), ",")
		req.Param("ip_list", ip_list)
		req.Param("get_url", fmt.Sprintf("/v1/team/%s/env/%s/stacks?requestPage=%s&pageSize=10", c.TeamId, c.ClustUuid, common.GetString(page)))
		ret_bytes, err := req.Bytes()
		if err != nil {
			beego.Error(err.Error())
			return err, []StackDataDetail{}
		}

		var ret CommonRet
		err = json.Unmarshal(ret_bytes, &ret)
		if err != nil {
			beego.Error(err.Error())
			return err, []StackDataDetail{}
		}
		if ret.Code == 0 {
			return errors.New(ret.Msg), []StackDataDetail{}
		}
		var data CaasStackList
		err = json.Unmarshal([]byte(ret.Data), &data)
		if err != nil {
			beego.Error(err.Error())
			return err, []StackDataDetail{}
		}
		stack_list = append(stack_list, data.Data.Data...)
		if data.Data.TotalSize <= page * 10 {
			break
		}
		page = page + 1
		// 避免高速调用
		time.Sleep(1 * time.Millisecond)
	}
	return nil, stack_list
}

func (c *McpCaasOpr) GetCaasServiceList() (error, []ServiceDataDetail) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	page := 1
	var service_list []ServiceDataDetail

	for {
		req := httplib.Get(url)
		req.Header("agent-auth", initial.AgentToken)
		ip_list := strings.Join(common.GetLocalIp(), ",")
		req.Param("ip_list", ip_list)
		req.Param("get_url", fmt.Sprintf("/v1/team/%s/env/%s/stack/%s/servicelist?requestPage=%s&pageSize=20",
			c.TeamId, c.ClustUuid, c.StackName, common.GetString(page)))
		ret_bytes, err := req.Bytes()
		if err != nil {
			beego.Error(err.Error())
			return err, []ServiceDataDetail{}
		}

		var ret CommonRet
		err = json.Unmarshal(ret_bytes, &ret)
		if err != nil {
			beego.Error(err.Error())
			return err, []ServiceDataDetail{}
		}
		if ret.Code == 0 {
			return errors.New(ret.Msg), []ServiceDataDetail{}
		}
		var data CaasServiceList
		err = json.Unmarshal([]byte(ret.Data), &data)
		if err != nil {
			beego.Error(err.Error())
			return err, []ServiceDataDetail{}
		}
		service_list = append(service_list, data.Data.Data...)
		if data.Data.TotalSize <= page * 20 {
			break
		}
		page = page + 1
		// 避免高速调用
		time.Sleep(1 * time.Millisecond)
	}
	return nil, service_list
}

func (c *McpCaasOpr) GetCaasServiceDetail() (error, ServiceStatusDetail) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", fmt.Sprintf("/v1/team/%s/env/%s/stack/%s/service/%s", c.TeamId, c.ClustUuid, c.StackName, c.ServiceName))
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, ServiceStatusDetail{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, ServiceStatusDetail{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), ServiceStatusDetail{}
	}
	var data CaasServiceStatus
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err, ServiceStatusDetail{}
	}
	return nil, data.Data
}

func (c *McpCaasOpr) GetCaasInstanceList() (error, []InstanceDataDetail) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", fmt.Sprintf("/v1/team/%s/env/%s/stack/%s/service/%s/instance", c.TeamId,
		c.ClustUuid, c.StackName, c.ServiceName))
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, []InstanceDataDetail{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, []InstanceDataDetail{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), []InstanceDataDetail{}
	}
	var data CaasInstanceList
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err, []InstanceDataDetail{}
	}
	return nil, data.Data
}

// 升级镜像
func (c *McpCaasOpr) UpgradeCaasService(image string) error {
	commont_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/post", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
		"post_url": []string{fmt.Sprintf("/v1/team/%s/env/%s/stack/%s/service/%s", c.TeamId, c.ClustUuid, c.StackName, c.ServiceName)},
	}
	req := httplib.Post(commont_url + "?" + args.Encode())
	req.Header("agent-auth", initial.AgentToken)
	launch_config := LConfig{
		ImageUuid: image,
	}
	body_data := ServiceUpgradeInput{
		StartFirst: "false",
		LaunchConfig: launch_config,
		InitContainers: []string{},
	}
	body_date_bytes, _ := json.Marshal(body_data)
	req.Body(body_date_bytes)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg)
	}
	var data CaasApiOprRet
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	if data.Data["code"] == "201" && data.Data["message"] == "upgrade service success!" {
		return nil
	} else {
		return errors.New(data.Data["message"])
	}
}

// 完成升级
func (c *McpCaasOpr) FinishCaasService() error {
	commont_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/post", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
		"post_url": []string{fmt.Sprintf("/v1/team/%s/env/%s/stack/%s/service/%s/over", c.TeamId, c.ClustUuid, c.StackName, c.ServiceName)},
	}
	req := httplib.Post(commont_url + "?" + args.Encode())
	req.Header("agent-auth", initial.AgentToken)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg)
	}
	var data CaasApiOprRet
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	if data.Data["code"] == "201" && data.Data["message"] == "upgrade service over!" {
		return nil
	} else {
		return errors.New(data.Data["message"])
	}
}

// 删除pod
func (c *McpCaasOpr) DelCaasPod(pod_name string) error {
	commont_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/caas/delete", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
		"delete_url": []string{fmt.Sprintf("/v1/team/%s/env/%s/stack/%s/service/%s/instance/%s", c.TeamId,
			c.ClustUuid, c.StackName, c.ServiceName, pod_name)},
	}
	req := httplib.Post(commont_url + "?" + args.Encode())
	req.Header("agent-auth", initial.AgentToken)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg)
	}
	var data CaasApiOprRet
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	if data.Data["code"] == "201" && data.Data["message"] == "delete pod success" {
		return nil
	} else {
		return errors.New(data.Data["message"])
	}
}