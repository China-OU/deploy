package caas_conf

import (
	"github.com/astaxie/beego"
	"models"
	"initial"
	"strings"
	"controllers"
	"github.com/jinzhu/gorm"
	"encoding/json"
	"fmt"
	"controllers/unit_conf"
	"library/common"
	"time"
	"library/caas"
)

type ManageCaasAgentController struct {
	controllers.BaseController
}

func (c *ManageCaasAgentController) URLMapping() {
	c.Mapping("CaasAgentList", c.CaasAgentList)
	c.Mapping("CaasAgentCreate", c.CaasAgentCreate)
	c.Mapping("CaasAgentDel", c.CaasAgentDel)
	c.Mapping("CaasAgentSurv", c.CaasAgentSurv)
	c.Mapping("CaasSyncData", c.CaasSyncData)
	c.Mapping("CaasAgentCheck", c.CaasAgentCheck)
	c.Mapping("CaasAgentSearch", c.CaasAgentSearch)
}

// @Title 获取caas的agent列表信息
// @Description 获取caas的agent列表信息，包括agent的ip和port，caas管理平台的token等信息。
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200  {object} []models.CaasConf
// @Failure 403
// @router /caas/agent/list [get]
func (c *ManageCaasAgentController) CaasAgentList() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	page, rows := c.GetPageRows()
	var cnt int
	var agent_list []models.CaasConf
	err := initial.DB.Model(models.CaasConf{}).Count(&cnt).Order("is_delete asc, id desc").Offset((page - 1)*rows).
		Limit(rows).Find(&agent_list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 超管显示明文密码；普通管理员显示 ********
	for i:=0; i<len(agent_list); i++ {
		if c.Role == "super-admin" {
			agent_list[i].CaasToken = common.WebPwdEncrypt(common.AesDecrypt(agent_list[i].CaasToken))
		} else {
			agent_list[i].CaasToken = common.WebPwdEncrypt("********")
		}
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": agent_list,
	}
	c.SetJson(1, ret, "数据获取成功！")
}

// @Title caas的agent信息新增和修改接口
// @Description caas的agent信息新增和修改接口，只有超管才能操作
// @Param	body	body	models.CaasConf	true	"body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /caas/agent/create [post]
func (c *ManageCaasAgentController) CaasAgentCreate() {
	if c.Role != "super-admin" {
		c.SetJson(0, "", "您没有权限操作，只有超级管理员才能操作！")
		return
	}

	var agent_conf models.CaasConf
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &agent_conf)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 基础校验
	net_flag := unit_conf.CheckBasicInfo(fmt.Sprintf("dumd_comp_en = '%s'", agent_conf.DeployComp))
	if !net_flag {
		c.SetJson(0, "", "部署租户不正确！")
		return
	}
	if common.CheckIp(agent_conf.AgentIp) == false {
		c.SetJson(0, "", "agent的ip填写不正确！")
		return
	}
	if common.GetInt(agent_conf.CaasPort) > 65535 || common.GetInt(agent_conf.CaasPort) < 1024 ||
		common.GetInt(agent_conf.AgentPort) > 65535 || common.GetInt(agent_conf.AgentPort) < 1024 {
			c.SetJson(0, "", "端口地址有误，请选择1024到65535之间的端口！")
			return
	}
	if strings.Contains(agent_conf.CaasUrl, "http") {
		c.SetJson(0, "", "caas的地址不需要带http参数！")
		return
	}

	// 新增或者修改
	var cnt int
	initial.DB.Model(models.CaasConf{}).Where("id != ? and (deploy_comp = ? or agent_ip = ?)",
		agent_conf.Id, agent_conf.DeployComp, agent_conf.AgentIp).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "部署租户或者agent的ip已存在，请重新填写！")
		return
	}

	// 信息加解密处理
	agent_conf.DeployNetwork = ""
	agent_conf.CaasToken = common.AesEncrypt(common.WebPwdDecrypt(agent_conf.CaasToken))

	tx := initial.DB.Begin()
	input_id := agent_conf.Id
	if agent_conf.Id > 0 {
		// 只更新五个字段
		update_map := map[string]interface{}{
			"caas_url": agent_conf.CaasUrl,
			"caas_port": agent_conf.CaasPort,
			"caas_token" : agent_conf.CaasToken,
			"agent_ip": agent_conf.AgentIp,
			"agent_port": agent_conf.AgentPort,
			"is_delete": agent_conf.IsDelete,
		}
		err = tx.Model(models.CaasConf{}).Where("id=?", agent_conf.Id).Updates(update_map).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	} else {
		agent_conf.IsDelete = "0"
		agent_conf.InsertTime = time.Now().Format(initial.DatetimeFormat)
		err = tx.Create(&agent_conf).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	}
	tx.Commit()
	msg := "agent信息新增成功！"
	if input_id > 0 {
		// &符后，id会重新赋值，必须保存起来
		msg = "agent信息修改成功！"
	}
	c.SetJson(1, "", msg)
}

// @Title caas的agent信息删除
// @Description caas的agent信息删除
// @Param	id	query	string	true	"删除id"
// @Success 200 true or false
// @Failure 403
// @router /caas/agent/del [post]
func (c *ManageCaasAgentController) CaasAgentDel() {
	if c.Role != "super-admin" {
		c.SetJson(0, "", "您没有权限操作，只有超级管理员才能操作！")
		return
	}

	id := c.GetString("id")
	tx := initial.DB.Begin()
	err := tx.Model(models.CaasConf{}).Where("id=?", id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "agent信息删除成功！")
}

// @Title 检查agent和caas的api是否存活
// @Description 检查agent和caas的api是否存活
// @Param	id	query	string	true	"agent的id，为all时检测所有agent的存活，为数字时检测对应id的agent存活"
// @Success 200 true or false
// @Failure 403
// @router /caas/agent/surv [get]
func (c *ManageCaasAgentController) CaasAgentSurv() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	id := c.GetString("id")
	cond := "is_delete = 0"
	if id != "all" {
		cond = fmt.Sprintf("id = '%s' and is_delete = 0", id)
	}
	var cc []models.CaasConf
	err := initial.DB.Model(models.CaasConf{}).Where(cond).Find(&cc).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	// 检测存活状态
	for _, v := range cc {
		caas_opr := caas.CaasOpr{
			AgentConf: v,
		}
		_, status := caas_opr.CheckAgentStatus()
		tx := initial.DB.Begin()
		update_map := map[string]interface{}{
			"agent_surv": status.AgentStatus,
			"caas_api_surv": status.CaasApiStatus,
			"surv_check_time" : time.Now().Format(initial.DatetimeFormat),
		}
		err := tx.Model(models.CaasConf{}).Where("id=?", v.Id).Updates(update_map).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
		tx.Commit()
	}
	c.SetJson(1, "", "agent状态检查完成！")
}

// @Title 校验租户是否已添加
// @Description 校验租户是否已添加
// @Param	comp	query	string	true	"租户"
// @Success 200  true or false
// @Failure 403
// @router /caas/agent/check [get]
func (c *ManageCaasAgentController) CaasAgentCheck() {
	comp := c.GetString("comp")
	var cnt int
	err := initial.DB.Model(models.CaasConf{}).Where("deploy_comp=?", comp).Count(&cnt).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if cnt == 0 {
		c.SetJson(1, "", "无该租户数据，可以新增！")
	}
	c.SetJson(0, "", "已有该租户数据，不能重复添加！")
}

// @Description 查询agent信息
// @Param	search	query	string	true	"查询参数"
// @Success 200  {object} []models.CaasConf
// @Failure 403
// @router /caas/agent/search [get]
func (c *ManageCaasAgentController) CaasAgentSearch() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	cond := " is_delete=0 "
	search := c.GetString("search")
	if search != "" {
		cond += fmt.Sprintf(" and deploy_comp like '%%%s%%' ", search)
	}

	var agent_list []models.CaasConf
	err := initial.DB.Model(models.CaasConf{}).Where(cond).Order("id desc").Limit(10).Find(&agent_list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 超管显示明文密码；普通管理员显示 ********
	for i:=0; i<len(agent_list); i++ {
		agent_list[i].CaasToken = "********"
	}
	c.SetJson(1, agent_list, "数据获取成功！")
}