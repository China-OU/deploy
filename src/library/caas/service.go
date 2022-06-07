package caas

import (
	"models"
	"fmt"
	"initial"
	"strings"
	"library/common"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego"
	"encoding/json"
	"errors"
	"time"
)

type CaasOpr struct {
	AgentConf   models.CaasConf
	TeamId      string
	ClustUuid   string
	StackName   string
	ServiceName string
}



type StructRet struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data interface{}`json:"data"`
}
// 带重试机制的初始化
func (c *InitServiceData) RetryInitService(retry int) error {
	for i := 0; ; i ++ {
		if err :=  c.InitService(); err != nil {
			beego.Error(err)
			if i > retry {
				return err
			}
			time.Sleep(2 * time.Second)
		} else {
			return nil
		}
	}
}

// 初始化、更新容器服务
func (c *InitServiceData) InitService() error {
	ipStr := strings.Join(common.GetLocalIp(), ",")
	beego.Info("ip_list:", ipStr)
	urlStr := "http://%s:%s/agent/v1/opr/service/init?" +
		"ip_list=%s" +
		"&team_id=%s" +
		"&cluster_uuid=%s" +
		"&stack_name=%s"
	url := fmt.Sprintf(urlStr, c.AgentConf.AgentIp, c.AgentConf.AgentPort, ipStr, c.TeamId, c.ClusterUuid, c.StackName)
	envMap := models.Environment{"RUN_MODE_ENV":strings.ToUpper(beego.AppConfig.String("runmode"))}
	for k, v := range *c.Environment {
		envMap[k] = v
	}
	bodyData := InitServiceAgentData {
		c.Image,
		c.ServiceName,
		"true",
		envMap,
		c.LogConfig,
		c.HealthCheck,
		&models.Scaling{DefaultInstances: c.InstanceNum},
		c.Volume,
		c.Scheduler,
	}
	// 把内存限制去掉
	//bodyDataWithMemLimit := InitServiceAgentDataWithMemLimit{
	//	InitServiceAgentData: bodyData,
	//	MemLimit:             c.MemLimit,
	//}
	bodyDataBytes, _ := json.Marshal(bodyData)
	//if c.AppType == "app" {
	//	bodyDataBytes, _ = json.Marshal(bodyDataWithMemLimit)
	//}
	header := map[string]string{"agent-auth": initial.AgentToken}
	retBytes, err := common.Post(url, header, bodyDataBytes)
	if err != nil {
		return err
	}
	var ret StructRet
	err = json.Unmarshal(retBytes, &ret)
	if err != nil {
		return err
	}
	if  ret.Code == 1  {
		return nil
	} else {
		return errors.New(ret.Msg)
	}
}

// 获取服务状态，包括镜像
func (c *CaasOpr) GetServiceDetail() (error, ServiceStatusDetail) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/info/service", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	req.Param("team_id", c.TeamId)
	req.Param("clust_uuid", c.ClustUuid)
	req.Param("stack_name", c.StackName)
	req.Param("service_name", c.ServiceName)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, ServiceStatusDetail{}
	}

	type StructRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data ServiceStatusDetail `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, ServiceStatusDetail{}
	}
	return nil, ret.Data
}

// 带重试机制的获取服务详情
func (c *CaasOpr)RetryGetService(reqLimit int) (error, *ServiceStatusDetail) {
	for i:=0; ; i ++ {
		if err, detail := c.GetServiceDetail(); err != nil {
			beego.Error(err)
			if i > reqLimit {
				return err, nil
			}
			time.Sleep(1 * time.Second)
		} else {
			return nil, &detail
		}
	}
}
// 获取实例列表
func (c *CaasOpr) GetInstanceList() (error, []InstanceDataDetail) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/info/instancelist", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	req.Param("team_id", c.TeamId)
	req.Param("clust_uuid", c.ClustUuid)
	req.Param("stack_name", c.StackName)
	req.Param("service_name", c.ServiceName)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, nil
	}

	type StructRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data []InstanceDataDetail `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, nil
	}
	return nil, ret.Data
}

// 删除实例
func (c *CaasOpr) DelCaasInstance(pod_name string) error {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/opr/instance/delete", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Post(url)
	req.Header("agent-auth", initial.AgentToken)
	req.Param("team_id", c.TeamId)
	req.Param("clust_uuid", c.ClustUuid)
	req.Param("stack_name", c.StackName)
	req.Param("service_name", c.ServiceName)
	req.Param("pod_name", pod_name)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err
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
		return err
	}

	data := ret.Data
	if data["code"] == "201" && data["message"] == "delete pod success" {
		return nil
	} else {
		return errors.New(data["message"])
	}
}

// 升级镜像
func (c *CaasOpr) UpgradeService(image string) error {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/opr/service/upgrade", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Post(url)
	req.Header("agent-auth", initial.AgentToken)
	req.Param("team_id", c.TeamId)
	req.Param("clust_uuid", c.ClustUuid)
	req.Param("stack_name", c.StackName)
	req.Param("service_name", c.ServiceName)
	req.Param("image", image)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err
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
		return err
	}

	data := ret.Data
	if data["code"] == "201" && data["message"] == "upgrade service success!" {
		return nil
	} else {
		return errors.New(data["message"])
	}
}

// 完成升级
func (c *CaasOpr) FinishUpgradeService() error {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/opr/service/finish", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Post(url)
	req.Header("agent-auth", initial.AgentToken)
	req.Param("team_id", c.TeamId)
	req.Param("clust_uuid", c.ClustUuid)
	req.Param("stack_name", c.StackName)
	req.Param("service_name", c.ServiceName)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err
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
		return err
	}

	data := ret.Data
	if data["code"] == "201" && data["message"] == "upgrade service over!" {
		return nil
	} else {
		return errors.New(data["message"])
	}
}

// 检查agent的存活状态
func (c *CaasOpr) CheckAgentStatus() (error, AgentCheck) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/info/status", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	// api到agent的超时时间 > agent到caas-api的超时时间，之前都是10秒
	req.SetTimeout(15 * time.Second, 15 * time.Second)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, AgentCheck{"unreachable", "unreachable"}
	}

	type StructRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data AgentCheck `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, AgentCheck{"unreachable", "unreachable"}
	}
	return nil, ret.Data
}