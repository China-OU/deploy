package master

import (
	"controllers"
	"library/common"
	"strings"
	"models"
	"initial"
)

type CaasBaseConfController struct {
	controllers.BaseAgentController
}

func (c *CaasBaseConfController) URLMapping() {
	c.Mapping("CaasConfList", c.CaasConfList)
	c.Mapping("DBConfList", c.DBConfList)
	c.Mapping("McpConfList", c.McpConfList)
}

// CaasConf方法
// @Title CaasConf
// @Description  获取caas的配置。
// @Success 200 {object} true or false
// @Failure 403
// @router /caas/conf [get]
func (c *CaasBaseConfController) CaasConfList() {
	// 先解密token
	caas_origin_token := common.AesDecrypt(c.CaasConf.CaasToken)
	// 解密发送
	caas_conf := map[string]string {
		"caas_url": c.CaasConf.CaasUrl,
		"caas_token": common.AgentEncrypt(caas_origin_token),
		"caas_port": c.CaasConf.CaasPort,
	}
	c.SetJson(1, caas_conf, "配置获取成功！")
}

// DBConf方法
// @Title DBConf方法
// @Description  获取数据库的配置。
// @Param	dbname	query	string	true	"数据库名"
// @Success 200 {object} true or false
// @Failure 403
// @router /database/conf [get]
func (c *CaasBaseConfController) DBConfList() {
	dbname := c.GetString("dbname")
	dtype := c.GetString("dtype")
	db_host := c.GetString("db_host")
	db_conf := make(map[string]interface{})
	if strings.TrimSpace(dbname) == "" {
		c.SetJson(0, db_conf, "数据库名为空！")
		return
	}

	var conf models.UnitConfDb
	err := initial.DB.Model(models.UnitConfDb{}).Where("is_delete=0 and dbname=? and type=? and host=?", dbname, dtype, db_host).First(&conf).Error
	if err != nil {
		c.SetJson(0, db_conf, err.Error())
		return
	}

	db_conf["host"] = conf.Host
	db_conf["port"] = conf.Port
	db_conf["dbname"] = conf.Dbname
	db_conf["username"] = conf.Username
	db_conf["password"] = common.AgentEncrypt(common.AesDecrypt(conf.EncryPwd))
	c.SetJson(1, db_conf, "数据库配置获取成功！")
}

// @Title 获取多容器平台的配置
// @Description  获取多容器平台的配置。
// @Param	container_type	query	string	true	"容器平台类型"
// @Success 200 {object} true or false
// @Failure 403
// @router /mcp/conf [get]
func (c *CaasBaseConfController) McpConfList() {
	container_type := c.GetString("container_type")
	if strings.TrimSpace(container_type) == "" {
		c.SetJson(0, nil, "容器类型为空！")
		return
	}

	var conf models.McpConfAgent
	err := initial.DB.Model(models.McpConfAgent{}).Where("agent_id=? and mcp_type=? and is_delete=0", c.CaasConf.Id,
		container_type).First(&conf).Error
	if err != nil {
		c.SetJson(0, conf, err.Error())
		return
	}

	origin_token := common.AesDecrypt(conf.Token)
	// 解密发送
	mcp_conf := map[string]string {
		"base_url": conf.BaseUrl,
		"token": common.AgentEncrypt(origin_token),
	}
	c.SetJson(1, mcp_conf, "容器平台配置获取成功！")
}