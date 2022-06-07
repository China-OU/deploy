package ext

import (
	"controllers"
	"controllers/unit_conf"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"models"
	"strings"
)

type ExtPmsFuncController struct {
	controllers.BaseUrlAuthController
}
//用于发布系统获取多容器配置信息
func (c *ExtPmsFuncController) URLMapping() {
	c.Mapping("GetMcpConf", c.GetMcpConf)
}

// @Title GetMcpConf
// @Description 获取多租户平台配置
// @Param	body	body	ext.UnitInfo	true	"发布单元全称"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	false	"调试模式"
// @Success 200 {object} models.UnitConfMcp
// @Failure 403
// @router /pms/mcp/query [post]
func (c *ExtPmsFuncController) GetMcpConf() {
	var input UnitInfo
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		beego.Info(string(c.Ctx.Input.RequestBody))
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	cond := " is_delete=0 "
	if strings.TrimSpace(input.UnitName) != "" {
		cond += fmt.Sprintf(" and b.info = '%s' ", input.UnitName)
	}

	type McpInfo struct {
		models.UnitConfMcp
		Info string `json:"unit"`
		CompName     string    `json:"comp_name"`
		// 下面两个字段跟据不同的容器服务，自主定义，要求重要可读即可
		Namespace    string    `json:"namespace"`
		ServiceName  string    `json:"service_name"`
	}

	var cnt int
	var mcp McpInfo
	err = initial.DB.Table("unit_conf_mcp a").Select("a.*, b.info").
		Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond).Count(&cnt).Find(&mcp).Error
	if cnt == 0 {
		beego.Info(err.Error())
		c.SetJson(0, mcp, input.UnitName+"容器配置信息未维护，请联系部署值班人员处理！")
		return
	}
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, mcp, err.Error())
		return
	}
	mcp.CompName = cfunc.GetCompCnName(mcp.DeployComp)
	mcp.Namespace, mcp.ServiceName = unit_conf.GetContainerInfo(mcp.Id, mcp.ContainerType)
	beego.Info(mcp)
	c.SetJson(1, mcp, "数据获取成功！")
}

type UnitInfo struct {
	UnitName string `json:"unit_name"`
}
