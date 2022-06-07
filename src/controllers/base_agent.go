package controllers

import (
	"github.com/astaxie/beego"
	"initial"
	"strings"
	"library/common"
	"models"
)

// 用于agent访问master节点做的权限校验
type BaseAgentController struct {
	beego.Controller
	CaasConf models.CaasConf
}

func (c *BaseAgentController) Prepare() {
	header := c.Ctx.Input.Header("master-auth")
	// 判断是否有header
	if header == "" {
		c.SetJson(0, "", "没有带token！")
		return
	}
	if header != initial.MasterToken {
		c.SetJson(0, "", "token不对！")
		return
	}
	// 判断ip是否从agent节点过来的
	ip := c.GetString("ip")
	if ip == "" {
		c.SetJson(0, "", "需要传入ip数据！")
		return
	}
	ip_list := common.MasterDecrypt(c.GetString("ip"))
	if ip_list == "" {
		c.SetJson(0, "", "传入的数据有误！")
		return
	}
	ip_arr := strings.Split(ip_list, ",")
	var cnt int
	var caas_conf models.CaasConf
	for _, v := range ip_arr {
		initial.DB.Model(models.CaasConf{}).Where("agent_ip=? and is_delete=0", v).Count(&cnt).First(&caas_conf)
		if cnt > 0 {
			break
		}
	}
	if cnt == 0 {
		c.SetJson(0, "", "数据库中没有该配置，请联系管理员添加！")
		return
	}
	c.CaasConf = caas_conf
	//beego.Info(strings.Split(c.Ctx.Request.RemoteAddr, ":")[0])
	if strings.Split(c.Ctx.Request.RemoteAddr, ":")[0] != c.CaasConf.AgentIp {
		c.SetJson(0, "", "无权发送请求！")
		return
	}
}


func (c *BaseAgentController) SetJson(code int, data interface{}, Msg string) {
	c.Data["json"] = map[string]interface{}{"code": code, "msg": Msg, "data": data}
	c.ServeJSON()
}