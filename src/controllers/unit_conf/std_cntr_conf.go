package unit_conf

import (
	"controllers"
	"strings"
	"fmt"
	"initial"
	"models"
	"github.com/astaxie/beego"
	"library/cfunc"
	"library/common"
	"encoding/json"
)

type StdCntrConfController struct {
	controllers.BaseController
}

func (c *StdCntrConfController) URLMapping() {
	c.Mapping("CntrList", c.CntrList)
	c.Mapping("CntrEdit", c.CntrEdit)
	c.Mapping("CntrDel", c.CntrDel)
	c.Mapping("JenkXmlList", c.JenkXmlList)
	c.Mapping("JenkXmlEdit", c.JenkXmlEdit)
	c.Mapping("JenkXmlConfirm", c.JenkXmlConfirm)
	c.Mapping("CpdsConfig", c.CpdsConfig)
}

type CntrInfo struct {
	models.UnitConfCntr
	Unit string `json:"unit"`
	Name string `json:"name"`
	Leader string `json:"leader"`
	// 取反射值
	LeaderName   string    `json:"leader_name"`
	CompName     string    `json:"comp_name"`
	// 取容器部分参数
	Container    string    `json:"container"`
	Team         string    `json:"team"`
	Stack        string    `json:"stack"`
	Service      string    `json:"service"`
}

// GetAll 方法
// @Title Get All
// @Description 获取所有发布单元列表
// @Param	cntr_sub_type	query	string	false	"MAVEN/ANT/GRADLE/NODE/SIMPLE"
// @Param	en_name	query	string	false	"发布单元英文名，支持模糊搜索"
// @Param	cpds_flag	query	string	false	"已接入/未接入，值分别为1/0"
// @Param	config_style	query	string	false	"caas配置/多容器配置，值分别为caas/mcp"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} models.UnitConfList
// @Failure 403
// @router /cntr/list [get]
func (c *StdCntrConfController) CntrList() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	cntr_sub_type := c.GetString("cntr_sub_type")
	en_name := c.GetString("en_name")
	cpds_flag := c.GetString("cpds_flag")
	config_style := c.GetString("config_style")
	page, rows := c.GetPageRows()
	cond := " is_delete=0 "
	if strings.TrimSpace(en_name) != "" {
		cond += fmt.Sprintf(" and b.unit like '%%%s%%' ", en_name)
	}
	if cntr_sub_type != "" {
		cond += fmt.Sprintf(" and a.app_sub_type = '%s' ", cntr_sub_type)
	}
	if cpds_flag != "" {
		cond += fmt.Sprintf(" and a.cpds_flag = %d ", common.GetInt(cpds_flag))
	}
	if config_style == "caas" {
		cond += fmt.Sprintf(" and a.mcp_conf_id = 0 and a.service_name != '' ")
	}
	if config_style == "mcp" {
		cond += fmt.Sprintf(" and ( a.mcp_conf_id > 0 or a.service_name = '') ")
	}

	var cnt int
	var cntr []CntrInfo
	err := initial.DB.Table("unit_conf_cntr a").Select("a.*, b.unit, b.name, b.leader").
		Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond).Count(&cnt).Order("a.id desc").Offset((page - 1)*rows).Limit(rows).Find(&cntr).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	for i:=0; i<len(cntr); i++ {
		cntr[i].LeaderName = cfunc.GetUserCnName(cntr[i].Leader)
		cntr[i].CompName = cfunc.GetCompCnName(cntr[i].DeployComp)
		if cntr[i].McpConfId == 0 {
			cntr[i].Container = "caas"
			cntr[i].Team = cfunc.GetTeamCnName(cntr[i].CaasTeam, cntr[i].DeployComp)
			cntr[i].Stack = cntr[i].CaasStack
			cntr[i].Service = cntr[i].ServiceName
		} else {
			// 获取关联id
			cntr[i] = GetCntrInfo(cntr[i])
		}
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": cntr,
	}
	c.SetJson(1, ret, "数据获取成功！")
}

func GetCntrInfo(info CntrInfo) CntrInfo {
	var mcp models.UnitConfMcp
	initial.DB.Model(models.UnitConfMcp{}).Where("id=?", info.McpConfId).First(&mcp)
	info.AppType = mcp.AppType
	info.AppSubType = mcp.AppSubType
	info.DeployComp = mcp.DeployComp
	info.Container = mcp.ContainerType

	if mcp.ContainerType == "istio" {
		err, istio := GetIstioConfDetail(info.McpConfId)
		if err != nil {
			info.Service = "error"
			return info
		}
		info.Team = "无"
		info.Stack = istio.Namespace
		info.Service = istio.Deployment
		return info
	}

	if mcp.ContainerType == "caas" {
		err, caas := GetCaasConfDetail(info.McpConfId)
		if err != nil {
			info.Service = "error"
			return info
		}
		info.Team = caas.TeamName
		info.Stack = caas.StackName
		info.Service = caas.ServiceName
		return info
	}

	if mcp.ContainerType == "openshift" {
		return info
	}

	if mcp.ContainerType == "rancher" {
		err, rancher := GetRancherConfDetail(info.McpConfId)
		if err != nil {
			info.Service = "error"
			return info
		}
		info.Team = rancher.ProjectName
		info.Stack = rancher.StackName
		info.Service = rancher.ServiceName
		return info
	}
	return info
}

// @Title DelCntr
// @Description 删除cntr的配置
// @Param	id	query	string	true	"数据列的id"
// @Success 200 true or false
// @Failure 403
// @router /cntr/del [post]
func (c *StdCntrConfController) CntrDel() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作，请联系管理员进行删除！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	tx := initial.DB.Begin()
	err = tx.Model(models.UnitConfCntr{}).Where("id=?", id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "标准容器的配置删除成功！")
}

// @Title 配置自助部署
// @Description 配置自助部署
// @Param	body	body	models.UnitConfCntr	true	"body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /cntr/cpds/config [post]
func (c *StdCntrConfController) CpdsConfig() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作，请联系管理员进行删除！")
		return
	}

	var cntr models.UnitConfCntr
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cntr)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	tx := initial.DB.Begin()
	err = tx.Model(models.UnitConfCntr{}).Where("id=?", cntr.Id).Update("cpds_flag", cntr.CpdsFlag).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "自助部署配置成功！")
}
