package caas

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"initial"
	"library/common"
	"models"
	"strings"
)

func (c *CaasOpr) RetryGetRouteConfig(reqLimit int) (error, *models.CaasRouteData) {
	for i := 0; ; i++ {
		if err, caasData := c.getRouteConfig(); err != nil {
			if i == reqLimit {
				return err, nil
			}
			err = nil
		} else {
			return nil, caasData
		}
	}
}

// 获取服务配置
func (c *CaasOpr) getRouteConfig() (error, *models.CaasRouteData) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/request/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	caasRoute := fmt.Sprintf("v1/team/%s/env/%s/stack/%s/service/%s/configInfo",
		c.TeamId, c.ClustUuid, c.StackName, c.ServiceName)
	req.Header("agent-auth", initial.AgentToken)
	req.Header("caas-route", caasRoute)
	ipList := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ipList)
	retBytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, nil
	}
	type StructRet struct {
		Code int                  `json:"code"`
		Msg  string               `json:"msg"`
		Data models.CaasRouteData `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(retBytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, nil
	}
	if ret.Code != 1 {
		return errors.New("caas接口返回错误，请重试！错误信息：" + ret.Msg), nil
	}
	return nil, &ret.Data
}

func (c *CaasOpr) RetryEditRoute(data models.CaasRouteData, reqLimit int) error {
	for i := 0; ; i++ {
		if err := c.editRoute(data); err != nil {
			beego.Error(err)
			if i == reqLimit {
				return err
			}
			err = nil
		} else {
			return nil
		}
	}
}

func (c *CaasOpr) editRoute(data models.CaasRouteData) error {
	ipStr := strings.Join(common.GetLocalIp(), ",")
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/request/post?ip_list=%s", c.AgentConf.AgentIp, c.AgentConf.AgentPort, ipStr)
	bodyDataBytes, _ := json.Marshal(data)
	caasRoute := fmt.Sprintf("/v1/team/%s/env/%s/stack/%s/route",
		c.TeamId, c.ClustUuid, c.StackName)
	header := map[string]string{"agent-auth": initial.AgentToken, "caas-route": caasRoute}
	retBytes, err := common.Post(url, header, bodyDataBytes)
	if err != nil {
		return err
	}
	var ret StructRet
	err = json.Unmarshal(retBytes, &ret)
	if err != nil {
		return err
	}
	if ret.Code == 1 {
		return nil
	} else {
		return errors.New(ret.Msg)
	}
}

func (c *CaasOpr) editRouteWithHttpLib(data models.CaasRouteData) (err error) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/request/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	caasRoute := fmt.Sprintf("v1/team/%s/env/%s/stack/%s/route",
		c.TeamId, c.ClustUuid, c.StackName)
	req.Header("agent-auth", initial.AgentToken)
	req.Header("caas-route", caasRoute)
	ipList := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ipList)
	bodyDataBytes, _ := json.Marshal(data)
	req, err = req.JSONBody(bodyDataBytes)
	if err != nil {
		return
	}
	retBytes, err := req.Bytes()
	if err != nil {
		return
	}
	var ret StructRet
	err = json.Unmarshal(retBytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	if ret.Code != 1 {
		return errors.New("caas接口返回错误，请重试！错误信息：" + ret.Msg)
	}
	return nil
}
