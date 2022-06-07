package ext

import (
	"initial"
	"models"
	"github.com/astaxie/beego"
	"strings"
	"encoding/json"
	"controllers/operation"
	"high-conc"
	"library/caas"
	"library/harbor"
	"errors"
	"library/mcp"
	"library/common"
)

// @Title 持续部署平台专用接口，基于发布单元名来操作
// @Description 持续部署平台专用接口，基于发布单元名来操作
// @Param	body	body	ext.UnitInput 	true	"body形式的数据，发布单元id名和镜像"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	true	"调试模式"
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/cpds/upgrade [post]
func (c *ExtCntrDeployController) CpdsUpgrade() {
	if c.Role != "admin" {
		// 如果是受限用户，需要做权限认定
	}

	var input UnitInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if input.UnitEn == "" || strings.TrimSpace(input.Image) == "" || strings.TrimSpace(input.RecordId) == "" {
		c.SetJson(0, "", "输入参数不能为空！")
		return
	}

	var base_conf models.UnitConfList
	err = initial.DB.Model(models.UnitConfList{}).Where("unit=? and is_offline=0", input.UnitEn).First(&base_conf).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	var unit_cntr models.UnitConfCntr
	err = initial.DB.Model(models.UnitConfCntr{}).Where("unit_id=? and is_delete = 0", base_conf.Id).First(&unit_cntr).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	err, caas_conf := operation.GetCaasConfig(unit_cntr.DeployComp)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	// 正在更新中的应用不允许再次更新
	cnt := 0
	initial.DB.Model(models.OprCntrUpgrade{}).Where("result = 2 and unit_id = ?", base_conf.Id).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "镜像正在更新中，不允许再次点击！")
		return
	}
	// 校验镜像在harbor中是否存在
	err = harbor.HarborCheckImage(input.Image)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	// 初始化连接caas，获取相关数据
	opr := caas.CaasOpr{
		AgentConf: caas_conf,
		TeamId: unit_cntr.CaasTeam,
		ClustUuid: unit_cntr.CaasCluster,
		StackName: unit_cntr.CaasStack,
		ServiceName: unit_cntr.ServiceName,
	}

	cntr_upgrade := operation.CntrUpgradeWithImage{
		Opr: opr,
		UnitId: base_conf.Id,
		Image: input.Image,
		Operator: c.Username,
		SourceId: input.RecordId,
	}
	high_conc.JobQueue <- &cntr_upgrade

	c.SetJson(1, "", "镜像更新已成功进入队列，请耐心等待执行结果！")
}

type UnitInput struct {
	UnitEn      string  `json:"unit_en"`
	Image       string  `json:"image"`
	RecordId    string  `json:"record_id"'`          // devops唯一标致符是32位字符串类型，不是int型
}


// 更新后的cpds升级接口

// @Title 持续部署平台专用接口，基于发布单元名来操作
// @Description 持续部署平台专用接口，基于发布单元名来操作
// @Param	body	body	ext.UnitInput 	true	"body形式的数据，发布单元id名和镜像"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	true	"调试模式"
// @Success 200 {object} {}
// @Failure 403
// @router /cpds/cntr-upgrade [post]
func (c *ExtCpdsFuncController) CntrServiceUpgrade() {
	var input UnitInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if input.UnitEn == "" || strings.TrimSpace(input.Image) == "" || strings.TrimSpace(input.RecordId) == "" {
		c.SetJson(0, "", "输入参数不能为空！")
		return
	}

	var base_conf models.UnitConfList
	err = initial.DB.Model(models.UnitConfList{}).Where("unit=? and is_offline=0", input.UnitEn).First(&base_conf).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	var conf models.UnitConfMcp
	err = initial.DB.Model(models.UnitConfMcp{}).Where("unit_id=? and is_delete = 0", base_conf.Id).First(&conf).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	err, agent_conf := operation.GetCaasConfig(conf.DeployComp)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	// 正在更新中的应用不允许再次更新
	cnt := 0
	initial.DB.Model(models.McpUpgradeList{}).Where("result = 2 and unit_id = ?", base_conf.Id).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "镜像正在更新中，不允许再次升级！如果自助部署状态异常，请过三分钟再刷新页面！")
		return
	}
	// 校验镜像在harbor中是否存在
	err = harbor.HarborCheckImage(input.Image)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	// 分种类更新应用
	if conf.ContainerType == "istio" {
		err := CpdsIstioUpgrade(conf, agent_conf, input, c.Username)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	}
	if conf.ContainerType == "caas" {
		err := CpdsCaasUpgrade(conf, agent_conf, input, c.Username)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	}
	if conf.ContainerType == "rancher" {
		err := CpdsRancherUpgrade(conf, agent_conf, input, c.Username)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	}
	//if conf.ContainerType == "openshift" {
	//	IstioUpgrade
	//}

	c.SetJson(1, "", "镜像更新已成功进入队列，请耐心等待执行结果！")
}

func CpdsIstioUpgrade(conf models.UnitConfMcp, agent_conf models.CaasConf, input UnitInput, operator string) error {
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
		if (v.Name == istio_conf.Container || v.Name == istio_conf.Deployment) && v.Image != input.Image && v.Image != "" {
			upgrade_flag = true
			old_image = v.Image
		}
	}
	if upgrade_flag == false {
		return errors.New("镜像相同，不能更新，请构建不同名镜像！")
	}

	istio_upgrade := operation.IstioUpgradeWithImage {
		Opr: istio,
		UnitId: common.GetInt(conf.UnitId),
		Image: input.Image,
		OldImage: old_image,
		Operator: operator,
		SourceId: input.RecordId,
	}
	high_conc.JobQueue <- &istio_upgrade
	return nil
}

func CpdsCaasUpgrade(conf models.UnitConfMcp, agent_conf models.CaasConf, input UnitInput, operator string) error {
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
	caas_upgrade := operation.McpCaasUpgrade {
		Opr: caas,
		UnitId: conf.UnitId,
		Image: input.Image,
		Operator: operator,
		SourceId: input.RecordId,
	}
	high_conc.JobQueue <- &caas_upgrade
	return nil
}

func CpdsRancherUpgrade(conf models.UnitConfMcp, agent_conf models.CaasConf, input UnitInput, operator string) error {
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
	rancher_upgrade := operation.McpRancherUpgrade {
		Opr: rancher,
		UnitId: conf.UnitId,
		Image: input.Image,
		Operator: operator,
		SourceId: input.RecordId,
	}
	high_conc.JobQueue <- &rancher_upgrade
	return nil
}

// @Title 轮询接口，查询执行结果
// @Description 轮询接口，查询执行结果
// @Param	record_list	query	string	true	"记录列表，比如 aaabb,ccc,dd,ee"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	true	"调试模式"
// @Success 200 true or false
// @Failure 403
// @router /cpds/cntr-poll [get]
func (c *ExtCpdsFuncController) CntrServicePoll() {
	record_list := c.GetString("record_list")
	r_list := strings.Split(record_list, ",")

	type PollRet struct {
		RecordId    string    `json:"record_id"`
		Result      int       `json:"result"`
		Msg         string    `json:"msg"`
		Cost        int       `json:"cost"`
	}

	var ret []PollRet
	for _, v := range r_list {
		record_id := strings.Trim(v, " ")
		var per PollRet
		per.RecordId = record_id
		var cntr_upgrade models.McpUpgradeList
		err := initial.DB.Model(models.McpUpgradeList{}).Where("source_id=?", record_id).Order("field (result, 1, 2, 10, 0)").
			First(&cntr_upgrade).Error
		if err != nil {
			beego.Error(err.Error())
			per.Msg = err.Error()
			per.Result = 100
			ret = append(ret, per)
			continue
		}
		per.Result = cntr_upgrade.Result
		per.Msg = cntr_upgrade.Message
		per.Cost = cntr_upgrade.CostTime
		ret = append(ret, per)
	}

	c.SetJson(1, ret, "结果查询成功！")
}