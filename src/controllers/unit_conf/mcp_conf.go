package unit_conf

import (
	"controllers"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"library/datalist"
	"models"
	"strings"
	"time"
)

type MultiContainerConfController struct {
	controllers.BaseController
}

func (c *MultiContainerConfController) URLMapping() {
	c.Mapping("McpConfList", c.McpConfList)
	c.Mapping("McpConfEdit", c.McpConfEdit)
	c.Mapping("McpConfDel", c.McpConfDel)
	c.Mapping("McpConfDetail", c.McpConfDetail)
}

// @Title ConfList
// @Description 获取多租户平台配置
// @Param	unit_name	query	string	false	"发布单元英文名，支持模糊搜索"
// @Param	app_type	query	string	false	"标准容器类型，app/web"
// @Param	deploy_comp	query	string	false	"部署租户，CMFT/CMRH"
// @Param	container_type	query	string	false	"容器类型, istio/.."
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} models.UnitConfMcp
// @Failure 403
// @router /mcp/list [get]
func (c *MultiContainerConfController) McpConfList() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	unit_name := c.GetString("unit_name")
	app_type := c.GetString("app_type")
	deploy_comp := c.GetString("deploy_comp")
	container_type := c.GetString("container_type")
	page, rows := c.GetPageRows()
	cond := " a.is_delete=0 "
	if strings.TrimSpace(unit_name) != "" {
		cond += fmt.Sprintf(" and b.unit like '%%%s%%' ", unit_name)
	}
	if app_type != "" {
		cond += fmt.Sprintf(" and a.app_type = '%s' ", app_type)
	}
	if deploy_comp != "" {
		cond += fmt.Sprintf(" and a.deploy_comp = '%s' ", deploy_comp)
	}
	if container_type != "" {
		cond += fmt.Sprintf(" and a.container_type = '%s' ", container_type)
	}

	type McpInfo struct {
		models.UnitConfMcp
		Unit string `json:"unit"`
		Name string `json:"name"`
		CompName     string    `json:"comp_name"`
		TypeName     string    `json:"type_name"`
		// 下面两个字段跟据不同的容器服务，自主定义，要求重要可读即可
		Namespace    string    `json:"namespace"`
		ServiceName  string    `json:"service_name"`
	}

	var cnt int
	var mcp []McpInfo
	err := initial.DB.Table("unit_conf_mcp a").Select("a.*, b.unit, b.name").
		Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond).Count(&cnt).Order("a.id desc").Offset((page - 1)*rows).Limit(rows).Find(&mcp).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	for i:=0; i<len(mcp); i++ {
		mcp[i].CompName = cfunc.GetCompCnName(mcp[i].DeployComp)
		mcp[i].TypeName = cfunc.GetTypeCnName(mcp[i].AppType)
		mcp[i].Namespace, mcp[i].ServiceName = GetContainerInfo(mcp[i].Id, mcp[i].ContainerType)
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": mcp,
	}
	c.SetJson(1, ret, "数据获取成功！")
}

func GetContainerInfo(mcp_id int, ctype string) (string, string) {
	if ctype == "istio" {
		err, istio := GetIstioConfDetail(mcp_id)
		if err != nil {
			return "error", "error"
		}
		return istio.Namespace, istio.Deployment
	}

	if ctype == "caas" {
		err, caas := GetCaasConfDetail(mcp_id)
		if err != nil {
			return "error", "error"
		}
		return fmt.Sprintf("%s / %s", caas.TeamName, caas.StackName), caas.ServiceName
	}

	if ctype == "openshift" {
		return "", ""
	}

	if ctype == "rancher" {
		err, rancher := GetRancherConfDetail(mcp_id)
		if err != nil {
			return "error", "error"
		}
		return rancher.StackName, rancher.ServiceName
	}

	return "", ""
}

// @Title 多容器平台配置录入和修改
// @Description 多容器平台配置录入和修改，支持rancher、istio和openshift
// @Param	body	body	models.UnitConfMcp	true	"body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /mcp/edit [post]
func (c *MultiContainerConfController) McpConfEdit() {
	if c.Role == "guest" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	if beego.AppConfig.String("runmode") == "prd" && strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "生产环境权限收缩，您没有权限操作！")
		return
	}
	var mcp datalist.McpCommonInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &mcp)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if mcp.BaseInfo.ContainerType == "" || mcp.BaseInfo.UnitId == 0 {
		c.SetJson(0, "", "数据格式传入错误，请修改！")
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(mcp.BaseInfo.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的编辑权限，请联系发布单元负责人进行操作！")
		return
	}
	var cnt int
	initial.DB.Model(models.UnitConfMcp{}).Where("id != ? and unit_id = ? and is_delete = 0", mcp.BaseInfo.Id,
		mcp.BaseInfo.UnitId).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "此发布单元已经创建，不能重复创建。如果需要创建备份发布单元，请在主列表页录入新发布单元！")
		return
	}

	ret_msg := "容器配置新增成功"
	if mcp.BaseInfo.ContainerType == "istio" {
		// 走云原生的istio逻辑
		var istio datalist.IstioConfInput
		err := json.Unmarshal(c.Ctx.Input.RequestBody, &istio)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		istio.IstioInfo.Operator = c.UserId
		err, msg := istioConfEdit(istio)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		} else {
			ret_msg = msg
		}
	}

	if mcp.BaseInfo.ContainerType == "caas" {
		// 走caas逻辑
		var caas datalist.CaasConfInput
		err := json.Unmarshal(c.Ctx.Input.RequestBody, &caas)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		caas.CaasInfo.Operator = c.UserId
		err, msg := caasConfEdit(caas)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		} else {
			ret_msg = msg
		}
	}

	if mcp.BaseInfo.ContainerType == "rancher" {
		// 走rancher部署逻辑
		var rancher datalist.RancherConfInput
		err := json.Unmarshal(c.Ctx.Input.RequestBody, &rancher)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		rancher.RancherInfo.Operator = c.UserId
		err, msg := rancherConfEdit(rancher)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		} else {
			ret_msg = msg
		}
	}

	//if mcp.BaseInfo.ContainerType == "openshift" {
	//	// 走openshift录入逻辑，后续补充
	//}


	c.SetJson(1, "", ret_msg)
}

func istioConfEdit(conf datalist.IstioConfInput) (error, string) {
	// 校验
	flag, msg := McpBaseCheck(conf.BaseInfo)
	if flag == false {
		return errors.New(msg), ""
	}
	if conf.IstioInfo.Namespace == "" || conf.IstioInfo.Deployment == "" || conf.IstioInfo.Version == "" || conf.IstioInfo.Container == "" {
		return errors.New("istio服务路径不全，请配置完全！"), ""
	}
	// 录入或者更新
	tx := initial.DB.Begin()
	input_id := conf.BaseInfo.Id
	if conf.BaseInfo.Id > 0 {
		update_base_map := map[string]interface{}{
			"app_type": conf.BaseInfo.AppType,
			"app_sub_type" : conf.BaseInfo.AppSubType,
			"deploy_comp": conf.BaseInfo.DeployComp,
			"deploy_network": conf.BaseInfo.DeployNetwork,
			"container_type": "istio",
		}
		err := tx.Model(models.UnitConfMcp{}).Where("id=?", conf.BaseInfo.Id).Updates(update_base_map).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}

		update_istio_map := map[string]interface{}{
			"namespace": conf.IstioInfo.Namespace,
			"deployment" : conf.IstioInfo.Deployment,
			"version": conf.IstioInfo.Version,
			"container": conf.IstioInfo.Container,
			"operator": conf.IstioInfo.Operator,
		}
		err = tx.Model(models.McpConfIstio{}).Where("mcp_id=? and is_delete=0", conf.BaseInfo.Id).Updates(update_istio_map).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}
	} else {
		conf.BaseInfo.InsertTime = time.Now().Format(initial.DatetimeFormat)
		conf.BaseInfo.IsDelete = 0
		err := tx.Create(&conf.BaseInfo).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}

		conf.IstioInfo.InsertTime = conf.BaseInfo.InsertTime
		conf.IstioInfo.IsDelete = 0
		conf.IstioInfo.McpId = conf.BaseInfo.Id
		err = tx.Create(&conf.IstioInfo).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}
	}
	tx.Commit()
	ret_msg := "k8s-istion发布单元配置新增成功！"
	if input_id > 0 {
		ret_msg = "k8s-istion发布单元配置维护成功！"
	}
	return nil, ret_msg
}

func caasConfEdit(conf datalist.CaasConfInput) (error, string) {
	// 校验
	flag, msg := McpBaseCheck(conf.BaseInfo)
	if flag == false {
		return errors.New(msg), ""
	}
	if conf.CaasInfo.TeamId == "" || conf.CaasInfo.ClusterUuid == "" || conf.CaasInfo.StackName == "" || conf.CaasInfo.ServiceName == "" {
		return errors.New("caas服务路径不全，请配置完全！"), ""
	}
	// 录入或者更新
	tx := initial.DB.Begin()
	input_id := conf.BaseInfo.Id
	if conf.BaseInfo.Id > 0 {
		update_base_map := map[string]interface{}{
			"app_type": conf.BaseInfo.AppType,
			"app_sub_type" : conf.BaseInfo.AppSubType,
			"deploy_comp": conf.BaseInfo.DeployComp,
			"deploy_network": conf.BaseInfo.DeployNetwork,
			"container_type": "caas",
		}
		err := tx.Model(models.UnitConfMcp{}).Where("id=?", conf.BaseInfo.Id).Updates(update_base_map).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}

		update_caas_map := map[string]interface{}{
			"team_id": conf.CaasInfo.TeamId,
			"team_name": conf.CaasInfo.TeamName,
			"cluster_uuid": conf.CaasInfo.ClusterUuid,
			"cluster_name": conf.CaasInfo.ClusterName,
			"stack_name": conf.CaasInfo.StackName,
			"service_name": conf.CaasInfo.ServiceName,
			"operator": conf.CaasInfo.Operator,
		}
		err = tx.Model(models.McpConfCaas{}).Where("mcp_id=? and is_delete=0", conf.BaseInfo.Id).Updates(update_caas_map).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}
	} else {
		conf.BaseInfo.InsertTime = time.Now().Format(initial.DatetimeFormat)
		conf.BaseInfo.IsDelete = 0
		err := tx.Create(&conf.BaseInfo).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}

		conf.CaasInfo.InsertTime = conf.BaseInfo.InsertTime
		conf.CaasInfo.IsDelete = 0
		conf.CaasInfo.McpId = conf.BaseInfo.Id
		err = tx.Create(&conf.CaasInfo).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}
	}
	tx.Commit()
	ret_msg := "caas发布单元配置新增成功！"
	if input_id > 0 {
		ret_msg = "caas发布单元配置维护成功！"
	}
	return nil, ret_msg
}

func rancherConfEdit(conf datalist.RancherConfInput) (error, string) {
	// 校验
	flag, msg := McpBaseCheck(conf.BaseInfo)
	if flag == false {
		return errors.New(msg), ""
	}
	if conf.RancherInfo.ProjectId == "" || conf.RancherInfo.StackId == "" || conf.RancherInfo.ServiceId == "" {
		return errors.New("rancher服务路径不全，请配置完全！"), ""
	}
	// 录入或者更新
	tx := initial.DB.Begin()
	input_id := conf.BaseInfo.Id
	if conf.BaseInfo.Id > 0 {
		update_base_map := map[string]interface{}{
			"app_type": conf.BaseInfo.AppType,
			"app_sub_type" : conf.BaseInfo.AppSubType,
			"deploy_comp": conf.BaseInfo.DeployComp,
			"deploy_network": conf.BaseInfo.DeployNetwork,
			"container_type": "rancher",
		}
		err := tx.Model(models.UnitConfMcp{}).Where("id=?", conf.BaseInfo.Id).Updates(update_base_map).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}

		update_rancher_map := map[string]interface{}{
			"project_id": conf.RancherInfo.ProjectId,
			"project_name": conf.RancherInfo.ProjectName,
			"stack_id": conf.RancherInfo.StackId,
			"stack_name": conf.RancherInfo.StackName,
			"service_id": conf.RancherInfo.ServiceId,
			"service_name": conf.RancherInfo.ServiceName,
			"operator": conf.RancherInfo.Operator,
		}
		err = tx.Model(models.McpConfRancher{}).Where("mcp_id=? and is_delete=0", conf.BaseInfo.Id).Updates(update_rancher_map).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}
	} else {
		conf.BaseInfo.InsertTime = time.Now().Format(initial.DatetimeFormat)
		conf.BaseInfo.IsDelete = 0
		err := tx.Create(&conf.BaseInfo).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}

		conf.RancherInfo.InsertTime = conf.BaseInfo.InsertTime
		conf.RancherInfo.IsDelete = 0
		conf.RancherInfo.McpId = conf.BaseInfo.Id
		err = tx.Create(&conf.RancherInfo).Error
		if err != nil {
			tx.Rollback()
			return err, ""
		}
	}
	tx.Commit()
	ret_msg := "rancher发布单元配置新增成功！"
	if input_id > 0 {
		ret_msg = "rancher发布单元配置维护成功！"
	}
	return nil, ret_msg
}






func McpBaseCheck(conf models.UnitConfMcp) (bool, string) {
	if conf.AppType == "db" {
		return false, "标准应用中不允许有db的发布单元！"
	}
	// 检查发布单元是否正确
	unit_flag := CheckBasicInfo(fmt.Sprintf("id = %d", conf.UnitId))
	if !unit_flag {
		return false, "基表中没有该发布单元！"
	}
	// 检查网络区域是否正确
	net_flag := CheckBasicInfo(fmt.Sprintf("dumd_comp_en = '%s'", conf.DeployComp))
	if !net_flag {
		return false, "部署租户不正确！"
	}
	return true, ""
}
