package operation

import (
	"controllers"
	"library/mcp"
	"github.com/astaxie/beego"
	"initial"
	"models"
	"errors"
	"library/cfunc"
	"library/common"
	"strings"
	"encoding/json"
	"library/harbor"
	"high-conc"
	"time"
)

type McpOprController struct {
	controllers.BaseController
}

func (c *McpOprController) URLMapping() {
	c.Mapping("ServiceDetail", c.ServiceDetail)
	c.Mapping("ServiceUpgrade", c.ServiceUpgrade)
	c.Mapping("McpRecord", c.McpRecord)
	c.Mapping("McpRecordList", c.McpRecordList)
}

// @Title ServiceDetail
// @Description 获取容器平台的应用信息及状态，支持多容器平台
// @Param	unit_id	query	string	true	"发布单元英文名，会查找对应的service，返回service信息"
// @Success 200 true or false
// @Failure 403
// @router /mcp/service/detail [get]
func (c *McpOprController) ServiceDetail() {
	unit_id := c.GetString("unit_id")
	var conf models.UnitConfMcp
	err := initial.DB.Model(models.UnitConfMcp{}).Where("unit_id=? and is_delete=0", unit_id).First(&conf).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	base := cfunc.GetUnitInfoById(common.GetInt(unit_id))
	if conf.ContainerType == "istio" {
		err, data := IstioServiceDetail(conf)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		data.Summary.UnitEn = base.Unit
		data.Summary.UnitCn = base.Name
		c.SetJson(1, data, "istio的服务数据获取成功！")
		return
	}

	if conf.ContainerType == "caas" {
		err, data := CaasServiceDetail(conf)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		data.Summary.UnitEn = base.Unit
		data.Summary.UnitCn = base.Name
		c.SetJson(1, data, "caas的服务数据获取成功！")
		return
	}

	if conf.ContainerType == "rancher" {
		err, data := RancherServiceDetail(conf)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		data.Summary.UnitEn = base.Unit
		data.Summary.UnitCn = base.Name
		c.SetJson(1, data, "caas的服务数据获取成功！")
		return
	}
	if conf.ContainerType == "openshift" {
		c.SetJson(0, "", "功能开发中，敬请期待！")
		return
	}

	c.SetJson(0, "", "没有匹配到容器平台类型！")
}

type McpDetailData struct {
	Summary   SummaryData       `json:"summary"`
	Instance  []mcp.PodRet `json:"instance"`
}
type SummaryData struct {
	UnitEn         string   `json:"unit_en"`
	UnitCn         string   `json:"unit_cn"`
	ServiceName    string   `json:"service_name"`
	ServiceStatus  string   `json:"service_status"`
	ServiceNum     int      `json:"service_num"`
	DeployComp     string   `json:"deploy_comp"`
	McpName        string   `json:"mcp_name"`
}

func IstioServiceDetail(conf models.UnitConfMcp) (error, McpDetailData) {
	var istio_conf models.McpConfIstio
	err := initial.DB.Model(models.McpConfIstio{}).Where("mcp_id=? and is_delete=0", conf.Id).First(&istio_conf).Error
	if err != nil {
		return err, McpDetailData{}
	}

	deploy_comp := conf.DeployComp
	err, cass_config := GetCaasConfig(deploy_comp)
	if err != nil {
		beego.Info(err.Error())
		return errors.New("agent配置有误"), McpDetailData{}
	}
	istio := mcp.McpIstioOpr {
		AgentConf: cass_config,
		Namespace: istio_conf.Namespace,
		Deployment: istio_conf.Deployment,
		Version: istio_conf.Version,
		Container: istio_conf.Container,
	}
	err, pod_detail := istio.GetIstioPodDetail()
	if err != nil {
		beego.Info(err.Error())
		return err, McpDetailData{}
	}
	err, status_detail := istio.GetIstioStatus()
	if err != nil {
		beego.Info(err.Error())
		return err, McpDetailData{}
	}
	var ret McpDetailData
	ret.Instance = pod_detail
	ret.Summary.ServiceName = istio_conf.Deployment + "-" + istio_conf.Version
	ret.Summary.ServiceNum = status_detail.Spec.Replicas
	ret.Summary.DeployComp = deploy_comp
	ret.Summary.McpName = "服务治理平台"
	ret.Summary.ServiceStatus = "upgrading"
	if status_detail.Spec.Replicas == status_detail.Status.Replicas && status_detail.Spec.Replicas == status_detail.Status.UpdatedReplicas {
		ret.Summary.ServiceStatus = "active"
	}
	return nil, ret
}

func CaasServiceDetail(conf models.UnitConfMcp) (error, McpDetailData) {
	var caas_conf models.McpConfCaas
	err := initial.DB.Model(models.McpConfCaas{}).Where("mcp_id=? and is_delete=0", conf.Id).First(&caas_conf).Error
	if err != nil {
		return err, McpDetailData{}
	}

	deploy_comp := conf.DeployComp
	err, cass_config := GetCaasConfig(deploy_comp)
	if err != nil {
		beego.Info(err.Error())
		return errors.New("agent配置有误"), McpDetailData{}
	}
	caas := mcp.McpCaasOpr {
		AgentConf: cass_config,
		TeamId: caas_conf.TeamId,
		ClustUuid: caas_conf.ClusterUuid,
		StackName: caas_conf.StackName,
		ServiceName: caas_conf.ServiceName,
	}
	err, detail := caas.GetCaasServiceDetail()
	if err != nil {
		beego.Info(err.Error())
		return err, McpDetailData{}
	}
	err, instance_list := caas.GetCaasInstanceList()
	if err != nil {
		beego.Info(err.Error())
		return err, McpDetailData{}
	}
	var ret McpDetailData
	ret.Summary.ServiceName = detail.Name
	ret.Summary.ServiceNum = len(instance_list)
	ret.Summary.DeployComp = deploy_comp
	ret.Summary.McpName = "caas容器平台"
	ret.Summary.ServiceStatus = detail.State
	for _, v := range instance_list {
		var ins mcp.PodRet
		ins.PodName = v.Name
		ins.Image = detail.Image
		ins.PodIP = v.Ip
		ins.StartTime = "无"
		ins.Status = v.State
		ret.Instance = append(ret.Instance, ins)
	}
	return nil, ret
}

func RancherServiceDetail(conf models.UnitConfMcp) (error, McpDetailData) {
	var rancher_conf models.McpConfRancher
	err := initial.DB.Model(models.McpConfRancher{}).Where("mcp_id=? and is_delete=0", conf.Id).First(&rancher_conf).Error
	if err != nil {
		return err, McpDetailData{}
	}

	deploy_comp := conf.DeployComp
	err, cass_config := GetCaasConfig(deploy_comp)
	if err != nil {
		beego.Info(err.Error())
		return errors.New("agent配置有误"), McpDetailData{}
	}
	rancher := mcp.McpRancherOpr {
		AgentConf: cass_config,
		ProjectId: rancher_conf.ProjectId,
		StackId: rancher_conf.StackId,
		ServiceId: rancher_conf.ServiceId,
	}
	err, detail := rancher.GetRancherService()
	if err != nil {
		beego.Info(err.Error())
		return err, McpDetailData{}
	}
	err, instance_list := rancher.GetRancherInstanceList()
	if err != nil {
		beego.Info(err.Error())
		return err, McpDetailData{}
	}
	var ret McpDetailData
	ret.Summary.ServiceName = detail.Name
	ret.Summary.ServiceNum = detail.CurrentScale
	ret.Summary.DeployComp = deploy_comp
	ret.Summary.McpName = "rancher容器平台"
	ret.Summary.ServiceStatus = detail.State
	for _, v := range instance_list {
		var ins mcp.PodRet
		ins.PodName = v.Name
		ins.Image = strings.TrimLeft(v.ImageUuid, "docker:")
		ins.PodIP = v.PrimaryIpAddress
		ins.Status = v.State
		tt, _ := time.Parse("2006-01-02T15:04:05Z", v.Created)
		ins.StartTime = tt.Add(8*time.Hour).Format(initial.DatetimeFormat)
		ret.Instance = append(ret.Instance, ins)
	}
	return nil, ret
}



// @Title UpgradeService
// @Description 更新多容器平台镜像
// @Param	body	body	operation.UpgradeInput	true	"body形式的数据，发布单元id名和镜像"
// @Success 200 {object} {}
// @Failure 403
// @router /mcp/service/upgrade [post]
func (c *McpOprController) ServiceUpgrade() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	var input UpgradeInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	unit_id := input.UnitId
	image := input.Image

	// 正在更新中的应用不允许再次更新
	var cnt int
	initial.DB.Model(models.McpUpgradeList{}).Where("result = 2 and unit_id = ?", unit_id).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "镜像正在更新中，不允许再次点击！")
		return
	}
	// 获取配置信息
	var conf models.UnitConfMcp
	err = initial.DB.Model(models.UnitConfMcp{}).Where("unit_id=? and is_delete=0", unit_id).First(&conf).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" {
		auth_flag := controllers.CheckUnitSingleAuth(conf.UnitId, c.UserId)
		if !auth_flag {
			c.SetJson(0, "", "您没有权限更新此发布单元，只有此发布单元的负责人、开发人员和测试人员才可以更新！")
			return
		}
	}
	// 校验镜像在harbor中是否存在
	err = harbor.HarborCheckImage(image)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	err, agent_conf := GetCaasConfig(conf.DeployComp)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	if conf.ContainerType == "istio" {
		err := IstioUpgrade(conf, agent_conf, image, c.UserId, 0)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	}
	if conf.ContainerType == "caas" {
		err := CaasUpgrade(conf, agent_conf, image, c.UserId, 0)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	}
	if conf.ContainerType == "rancher" {
		err := RancherUpgrade(conf, agent_conf, image, c.UserId, 0)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	}
	//if conf.ContainerType == "openshift" {
	//	IstioUpgrade
	//}



	c.SetJson(1, "", "多容器平台镜像更新已成功进入队列，请耐心等待执行结果！")
}

func IstioUpgrade(conf models.UnitConfMcp, agent_conf models.CaasConf, image, operator string, parent_id int) error {
	var istio_conf models.McpConfIstio
	err := initial.DB.Model(models.McpConfIstio{}).Where("mcp_id=? and is_delete=0", conf.Id).First(&istio_conf).Error
	if err != nil {
		return err
	}
	if istio_conf.Namespace == "" || istio_conf.Deployment == "" || istio_conf.Version == "" || istio_conf.Container == "" {
		return errors.New("容器服务没有关联，不允许升级")
	}

	istio := mcp.McpIstioOpr {
		AgentConf: agent_conf,
		Namespace: istio_conf.Namespace,
		Deployment: istio_conf.Deployment,
		Version: istio_conf.Version,
		Container: istio_conf.Container,
	}
	err, status := istio.GetIstioStatus()
	if err != nil {
		beego.Info(err.Error())
		return err
	}
	upgrade_flag := false
	old_image := ""
	for _, v := range status.Spec.Template.Spec.Containers {
		if (v.Name == istio_conf.Container || v.Name == istio_conf.Deployment) && v.Image != image && v.Image != "" {
			upgrade_flag = true
			old_image = v.Image
		}
	}
	if upgrade_flag == false {
		return errors.New("镜像相同，不能更新，请构建不同名镜像！")
	}

	istio_upgrade := IstioUpgradeWithImage {
		Opr: istio,
		UnitId: common.GetInt(conf.UnitId),
		Image: image,
		OldImage: old_image,
		Operator: operator,
		ParentId: parent_id,
	}
	high_conc.JobQueue <- &istio_upgrade
	return nil
}

func CaasUpgrade(conf models.UnitConfMcp, agent_conf models.CaasConf, image, operator string, parent_id int) error {
	var caas_conf models.McpConfCaas
	err := initial.DB.Model(models.McpConfCaas{}).Where("mcp_id=? and is_delete=0", conf.Id).First(&caas_conf).Error
	if err != nil {
		return err
	}
	if caas_conf.TeamId == "" || caas_conf.ClusterUuid == "" || caas_conf.StackName == "" || caas_conf.ServiceName == "" {
		return errors.New("容器服务没有关联，不允许升级")
	}

	caas := mcp.McpCaasOpr {
		AgentConf: agent_conf,
		TeamId: caas_conf.TeamId,
		ClustUuid: caas_conf.ClusterUuid,
		StackName: caas_conf.StackName,
		ServiceName: caas_conf.ServiceName,
	}
	caas_upgrade := McpCaasUpgrade {
		Opr: caas,
		UnitId: conf.UnitId,
		Image: image,
		Operator: operator,
		ParentId: parent_id,
	}
	high_conc.JobQueue <- &caas_upgrade
	return nil
}

func RancherUpgrade(conf models.UnitConfMcp, agent_conf models.CaasConf, image, operator string, parent_id int) error {
	var rancher_conf models.McpConfRancher
	err := initial.DB.Model(models.McpConfRancher{}).Where("mcp_id=? and is_delete=0", conf.Id).First(&rancher_conf).Error
	if err != nil {
		return err
	}
	if rancher_conf.ProjectId == "" || rancher_conf.StackId == "" || rancher_conf.ServiceId == "" {
		return errors.New("容器服务没有关联，不允许升级")
	}

	rancher := mcp.McpRancherOpr {
		AgentConf: agent_conf,
		ProjectId: rancher_conf.ProjectId,
		StackId: rancher_conf.StackId,
		ServiceId: rancher_conf.ServiceId,
	}
	rancher_upgrade := McpRancherUpgrade {
		Opr: rancher,
		UnitId: conf.UnitId,
		Image: image,
		Operator: operator,
		ParentId: parent_id,
	}
	high_conc.JobQueue <- &rancher_upgrade
	return nil
}