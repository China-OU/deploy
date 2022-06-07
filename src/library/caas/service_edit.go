package caas

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"initial"
	"library/common"
	"strings"
)

func (c *CaasOpr)RetryGetServiceConfig(reqLimit int) (error, *ServiceConfigAll) {
	for i := 0 ; ; i ++ {
		if err, caasData := c.GetServiceConfig(); err != nil {
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
func (c *CaasOpr)GetServiceConfig() (error, *ServiceConfigAll) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/request/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	caasRoute := fmt.Sprintf("v1/team/%s/env/%s/stack/%s/service/%s/configInfo",
		c.TeamId, c.ClustUuid, c.StackName, c.ServiceName)
	req.Header("agent-auth", initial.AgentToken)
	req.Header("caas-route", caasRoute)
	ipList := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ipList)
	retBytes, err := req.Debug(true).Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, nil
	}
	type StructRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data ServiceConfigAll `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(retBytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, nil
	}
	if ret.Code != 1 {
		return errors.New("caas接口返回错误，请重试！错误信息："+ ret.Msg ), nil
	}
	return nil, &ret.Data
}

// 更新服务配置
func (c *CaasOpr)UpdateServiceConfig() {

}
