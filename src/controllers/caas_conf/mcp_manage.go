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
	"library/common"
	"time"
)

type McpAgentConfController struct {
	controllers.BaseController
}

func (c *McpAgentConfController) URLMapping() {
	c.Mapping("McpAgentList", c.McpAgentList)
	c.Mapping("McpAgentEdit", c.McpAgentEdit)
	c.Mapping("McpAgentDel", c.McpAgentDel)
}

// @Title 获取多租户的url和token等信息
// @Description 获取多租户的url和token等信息
// @Param	mcp_type	query	string	true	"多容器平台类型"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200  {object} []models.McpConfAgent
// @Failure 403
// @router /mcp/agent/list [get]
func (c *McpAgentConfController) McpAgentList() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	mcp_type := c.GetString("mcp_type")
	page, rows := c.GetPageRows()
	type McpAgentInfo struct {
		models.McpConfAgent
		DeployComp   string  `json:"deploy_comp"`
		AgentIp      string  `json:"agent_ip"`
		AgentPort    int     `json:"agent_port"`
	}
	var cnt int
	var mcp_agent []McpAgentInfo
	cond := " a.is_delete=0 "
	if mcp_type != "" {
		cond += fmt.Sprintf(" and a.mcp_type='%s' ", mcp_type)
	}
	err := initial.DB.Table("mcp_conf_agent a").Select("b.deploy_comp, b.agent_ip, b.agent_port, a.*").
		Joins("LEFT JOIN conf_caas b ON a.agent_id = b.id").Where(cond).Count(&cnt).Order("a.id desc").
		Offset((page - 1)*rows).Limit(rows).Find(&mcp_agent).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 超管显示明文密码；普通管理员显示 ********
	for i:=0; i<len(mcp_agent); i++ {
		if c.Role == "super-admin" {
			mcp_agent[i].Token = common.WebPwdEncrypt(common.AesDecrypt(mcp_agent[i].Token))
		} else {
			mcp_agent[i].Token = common.WebPwdEncrypt("********")
		}
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": mcp_agent,
	}
	c.SetJson(1, ret, "数据获取成功！")
}

// @Title 编辑多租户的url和token等信息
// @Description 编辑多租户的url和token等信息
// @Param	body	body	models.McpConfAgent	true	"body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /mcp/agent/edit [post]
func (c *McpAgentConfController) McpAgentEdit() {
	if c.Role != "super-admin" {
		c.SetJson(0, "", "您没有权限操作，只有超级管理员才能操作！")
		return
	}

	var mcp_conf models.McpConfAgent
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &mcp_conf)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 基础校验
	var cnt int
	err = initial.DB.Model(models.CaasConf{}).Where("id=?", mcp_conf.AgentId).Count(&cnt).Error
	if cnt == 0 {
		c.SetJson(0, "", "没有此agent信息!")
		return
	}

	mcp_conf.Token = common.AesEncrypt(common.WebPwdDecrypt(mcp_conf.Token))
	mcp_conf.BaseUrl = strings.TrimSpace(mcp_conf.BaseUrl)
	tx := initial.DB.Begin()
	input_id := mcp_conf.Id
	if mcp_conf.Id > 0 {
		// 只更新五个字段
		update_map := map[string]interface{}{
			"base_url": mcp_conf.BaseUrl,
			"token": mcp_conf.Token,
			"operator" : c.UserId,
		}
		err = tx.Model(models.McpConfAgent{}).Where("id=?", mcp_conf.Id).Updates(update_map).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	} else {
		var cnt int
		err = initial.DB.Model(models.McpConfAgent{}).Where("agent_id=? and mcp_type=? and is_delete=0",
			mcp_conf.AgentId, mcp_conf.McpType).Count(&cnt).Error
		if cnt > 0 {
			c.SetJson(0, "", "已录入此容器类型数据，不能重复录入，可以进行编辑!")
			return
		}

		mcp_conf.InsertTime = time.Now().Format(initial.DatetimeFormat)
		mcp_conf.Operator = c.UserId
		mcp_conf.IsDelete = 0
		err = tx.Create(&mcp_conf).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	}
	tx.Commit()
	msg := "mcp信息新增成功！"
	if input_id > 0 {
		msg = "mcp信息修改成功！"
	}
	c.SetJson(1, "", msg)
}

// @Title 删除多租户的url和token等信息
// @Description 删除多租户的url和token等信息
// @Param	id	query	string	true	"删除id"
// @Success 200 true or false
// @Failure 403
// @router /mcp/agent/del [post]
func (c *McpAgentConfController) McpAgentDel() {
	if c.Role != "super-admin" {
		c.SetJson(0, "", "您没有权限操作，只有超级管理员才能操作！")
		return
	}

	id := c.GetString("id")
	tx := initial.DB.Begin()
	err := tx.Model(models.McpConfAgent{}).Where("id=?", id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "mcp信息删除成功！")
}