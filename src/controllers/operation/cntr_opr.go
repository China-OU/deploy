package operation

import (
	"controllers"
	"initial"
	"models"
	"github.com/astaxie/beego"
	"library/caas"
	"github.com/jinzhu/gorm"
	"errors"
	"encoding/json"
	"strings"
	"high-conc"
	"library/common"
	"library/harbor"
)

type CntrOprController struct {
	controllers.BaseController
	IsEdit bool
}

func (c *CntrOprController) URLMapping() {
	c.Mapping("SearchList", c.SearchList)
	c.Mapping("UpgradeService", c.UpgradeService)
	c.Mapping("GetRecord", c.GetRecord)
	c.Mapping("GetRecList", c.GetRecordList)
}

// SearchList 方法
// @Title SearchList
// @Description 获取容器平台的应用状态
// @Param	unit_id	query	string	true	"发布单元英文名，会查找对应的service_name，搜索k8s应用状态"
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/search [get]
func (c *CntrOprController) SearchList() {
	unit_id := c.GetString("unit_id")
	err, unit_cntr := GetCntrConfig(unit_id)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if unit_cntr.ServiceName == "" {
		c.SetJson(0, "", "容器服务没有关联，无法查询！")
		return
	}
	err, cass_config := GetCaasConfig(unit_cntr.DeployComp)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error() + "，请检查agent的配置是否正确！")
		return
	}
	// 初始化连接caas，获取相关数据
	opr := caas.CaasOpr{
		AgentConf: cass_config,
		TeamId: unit_cntr.CaasTeam,
		ClustUuid: unit_cntr.CaasCluster,
		StackName: unit_cntr.CaasStack,
		ServiceName: unit_cntr.ServiceName,
	}
	err, detail := opr.GetServiceDetail()
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	err, instance_list := opr.GetInstanceList()
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	ret := map[string]interface{}{
		"detail": detail,
		"instance": instance_list,
	}
	c.SetJson(1, ret, "服务数据获取成功！")
}

// 获取cass的配置
func GetCaasConfig(dcomp string) (error, models.CaasConf) {
	var data models.CaasConf
	err := initial.DB.Model(models.CaasConf{}).Where("deploy_comp=? and is_delete=0", dcomp).
		First(&data).Error
	if err != nil {
		return err, models.CaasConf{}
	}
	return nil, data
}

// 获取cntr发布单元的配置
func GetCntrConfig(unit_id string) (error, models.UnitConfCntr) {
	var cnt int
	var unit_cntr models.UnitConfCntr
	err := initial.DB.Model(models.UnitConfCntr{}).Where("unit_id = ? and is_delete = 0", unit_id).Count(&cnt).First(&unit_cntr).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		beego.Info(err.Error())
		return err, models.UnitConfCntr{}
	}
	if cnt == 0 {
		return errors.New("没有该发布单元，请重新选择！"), models.UnitConfCntr{}
	}
	return nil, unit_cntr
}

type UpgradeInput struct {
	UnitId string `json:"unit_id"`
	Image  string `json:"image"`
}

// @Title UpgradeService
// @Description 更新容器平台镜像
// @Param	body	body	operation.UpgradeInput	true	"body形式的数据，发布单元id名和镜像"
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/upgrade [post]
func (c *CntrOprController) UpgradeService() {
	// guest无操作权限；deploy-single 要单独判断权限
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
	initial.DB.Model(models.OprCntrUpgrade{}).Where("result = 2 and unit_id = ?", unit_id).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "镜像正在更新中，不允许再次点击！")
		return
	}

	// 获取配置信息
	err, unit_cntr := GetCntrConfig(unit_id)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if unit_cntr.ServiceName == "" {
		c.SetJson(0, "", "容器服务没有关联，不允许升级容器，请先关联容器服务！")
		return
	}
	if c.Role == "deploy-single" {
		auth_flag := controllers.CheckUnitSingleAuth(unit_cntr.UnitId, c.UserId)
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

	err, cass_config := GetCaasConfig(unit_cntr.DeployComp)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	// 初始化连接caas，获取相关数据
	opr := caas.CaasOpr{
		AgentConf: cass_config,
		TeamId: unit_cntr.CaasTeam,
		ClustUuid: unit_cntr.CaasCluster,
		StackName: unit_cntr.CaasStack,
		ServiceName: unit_cntr.ServiceName,
	}

	cntr_upgrade := CntrUpgradeWithImage{
		Opr: opr,
		UnitId: common.GetInt(unit_id),
		Image: image,
		Operator: c.UserId,
	}
	high_conc.JobQueue <- &cntr_upgrade

	c.SetJson(1, "", "镜像更新已成功进入队列，请耐心等待执行结果！")
}
