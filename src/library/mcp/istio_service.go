package mcp

import (
	"models"
	"fmt"
	"initial"
	"strings"
	"library/common"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego"
	"encoding/json"
	"time"
	"net/url"
	"errors"
)

type McpIstioOpr struct {
	AgentConf   models.CaasConf
	Namespace   string
	Deployment  string
	Version     string
	Container   string
}



type StructRet struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data interface{}`json:"data"`
}

// 命名空间查询
func (c *McpIstioOpr) GetIstioNamespace() (error, NamespaceData) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/istio/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", "/api/v1/namespaces")
	req.Param("run_env", beego.AppConfig.String("runmode"))
	req.Param("deploy_comp", c.AgentConf.DeployComp)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, NamespaceData{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, NamespaceData{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), NamespaceData{}
	}
	var data NamespaceData
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err, NamespaceData{}
	}
	return nil, data
}

// 部署查询，获取实例名、版本号和容器名
func (c *McpIstioOpr) GetIstioDeployment() (error, []DeploymentRet) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/istio/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments", c.Namespace))
	req.Param("run_env", beego.AppConfig.String("runmode"))
	req.Param("deploy_comp", c.AgentConf.DeployComp)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, []DeploymentRet{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, []DeploymentRet{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), []DeploymentRet{}
	}
	var data IstioData
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err, []DeploymentRet{}
	}

	var ret_data []DeploymentRet
	for _, v := range data.Items {
		var per DeploymentRet
		md_bytes, _ := json.Marshal(v.Metadata)
		var p DeploymentMetadata
		err := json.Unmarshal(md_bytes, &p)
		if err != nil {
			return err, []DeploymentRet{}
		}
		spec_bytes, _ := json.Marshal(v.Spec)
		var sd SpecData
		err = json.Unmarshal(spec_bytes, &sd)
		if err != nil {
			return err, []DeploymentRet{}
		}

		per.Name = p.Name
		per.App = p.Labels.App
		per.Version = p.Labels.Version
		if len(sd.Template.Spec.Containers) > 0 {
			per.Container = sd.Template.Spec.Containers[0].Name
		}
		ret_data = append(ret_data, per)
	}
	return nil, ret_data
}

// pod查询
func (c *McpIstioOpr) GetIstioPodDetail() (error, []PodRet) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/istio/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", fmt.Sprintf("/api/v1/namespaces/%s/pods?labelSelector=app=%s,version=%s",
		c.Namespace, c.Deployment, c.Version))
	req.Param("run_env", beego.AppConfig.String("runmode"))
	req.Param("deploy_comp", c.AgentConf.DeployComp)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, []PodRet{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, []PodRet{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), []PodRet{}
	}
	var data IstioData
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err, []PodRet{}
	}

	var ret_data []PodRet
	for _, v := range data.Items {
		var per PodRet
		md_bytes, _ := json.Marshal(v.Metadata)
		var p DeploymentMetadata
		err := json.Unmarshal(md_bytes, &p)
		if err != nil {
			return err, []PodRet{}
		}
		spec_bytes, _ := json.Marshal(v.Spec)
		var sd SpecContainer
		err = json.Unmarshal(spec_bytes, &sd)
		if err != nil {
			return err, []PodRet{}
		}
		status_bytes, _ := json.Marshal(v.Status)
		var ps PodStatus
		err = json.Unmarshal(status_bytes, &ps)
		if err != nil {
			return err, []PodRet{}
		}

		per.PodName = p.Name
		per.PodIP = ps.PodIP
		per.StartTime = ps.StartTime
		if strings.Contains(ps.StartTime, "T") && strings.Contains(ps.StartTime, "Z") {
			loc, _ := time.LoadLocation("Local")
			dt, _ := time.ParseInLocation("2006-01-02T15:04:05Z", ps.StartTime, loc)
			per.StartTime = dt.Add(8*time.Hour).Format(initial.DatetimeFormat)
		}
		per.Status = ps.Phase
		for _, v := range sd.Containers {
			if v.Name == c.Container || v.Name == c.Deployment {
				per.Image = v.Image
				break
			}
		}
		ret_data = append(ret_data, per)
	}
	return nil, ret_data
}

// 升级镜像
func (c *McpIstioOpr) UpgradeIstioService(image string) error {
	commont_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/istio/patch", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
		"post_url": []string{fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", c.Namespace, c.Deployment+"-"+c.Version)},
		"run_env": []string{beego.AppConfig.String("runmode")},
		"deploy_comp": []string{c.AgentConf.DeployComp},
	}
	req := httplib.Post(commont_url + "?" + args.Encode())
	req.Header("agent-auth", initial.AgentToken)
	var body UpgradeInput
	body.Spec.Template.Spec.Containers = append(body.Spec.Template.Spec.Containers, SpecContainerInfo{
		Name: c.Container,
		Image: image,
	})
	body_byte, _ := json.Marshal(body)
	req.Body(body_byte)
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
	var data StatusData
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	if data.Spec.Replicas > 0 {
		return nil
	}
	return errors.New("istio更新接口调用失败！")
}

// 部署状态查询，判断升级结果是否成功
func (c *McpIstioOpr) GetIstioStatus() (error, StatusData) {
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/mcp/istio/get", c.AgentConf.AgentIp, c.AgentConf.AgentPort)
	req := httplib.Get(url)
	req.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req.Param("ip_list", ip_list)
	req.Param("get_url", fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s/status", c.Namespace,
		c.Deployment+"-"+c.Version))
	req.Param("run_env", beego.AppConfig.String("runmode"))
	req.Param("deploy_comp", c.AgentConf.DeployComp)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return err, StatusData{}
	}

	var ret CommonRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		return err, StatusData{}
	}
	if ret.Code == 0 {
		return errors.New(ret.Msg), StatusData{}
	}
	var data StatusData
	err = json.Unmarshal([]byte(ret.Data), &data)
	if err != nil {
		beego.Error(err.Error())
		return err, StatusData{}
	}
	return nil, data
}