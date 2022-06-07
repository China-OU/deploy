package ext

import (
	"controllers/unit_conf"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"library/common"
	"models"
	"strings"
	"time"
)

type McpInfoController struct {
	beego.Controller
}

func (c *McpInfoController) URLMapping() {
	c.Mapping("MultiContainerList", c.MultiContainerList)
}

type MultiCntrInfo struct {
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
	//取子系统部分参数
	DumdCompEn string `json:"dumd_comp_en"`
	DumdSubSysname string `json:"dumd_sub_sysname"`
	DumdSubSysnameCn string `json:"dumd_sub_sysname_cn"`
	McpConfId     int       `json:"mcp_conf_id"`
}

// GetAll 方法
// @Title Get All
// @Description 获取所有发布单元容器服务列表
// @Param	en_name	query	string	false	"发布单元英文名，支持模糊搜索"
// @Success 200 {object} ext.MultiCntrInfo
// @Failure 403
// @router /cntr/list [get]
func (c *McpInfoController) MultiContainerList() {
	header := c.Ctx.Request.Header
	auth := ""
	if header["Authorization"] != nil && len(header["Authorization"]) > 0 {
		auth = header["Authorization"][0]
	} else {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "没有header!","data":[]MultiCntrInfo{}}
		c.ServeJSON()
		return
	}
	if strings.Replace(auth, "Basic ", "", -1) != "mdeploy_9BwLCSIYMSSbIFPEFNmZI1znSMUC0VaV" {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "header校验失败!","data":[]MultiCntrInfo{}}
		c.ServeJSON()
		return
	}

	last_time, flag := getLastCallTime()
	if time.Now().Add(- 15 * time.Minute).Format("2006-01-02 15:04:05") < common.GetString(last_time) && flag == 1 {
		c.Data["json"] = map[string]interface{}{"code": "0","message": "15分钟内只允许访问一次！上次访问时间：" + common.GetString(last_time),"data":[]MultiCntrInfo{}}
		c.ServeJSON()
		return
	}

	en_name := c.GetString("en_name")
	cond := " is_delete=0 "
	if strings.TrimSpace(en_name) != "" {
		cond += fmt.Sprintf(" and b.unit like '%%%s%%' ", en_name)
	}
	//config_style := "mcp"
	cond += fmt.Sprintf(" and ( a.mcp_conf_id > 0 or a.service_name = '') ")

	var cnt int
	var cntr []MultiCntrInfo
	err := initial.DB.Table("unit_conf_cntr a").Select("a.*, b.unit, b.name, b.leader, b.dumd_comp_en," +
		" b.dumd_sub_sysname, b.dumd_sub_sysname_cn").Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond).Count(&cnt).Order("a.id desc").Find(&cntr).Error
	if err != nil {
		beego.Info(err.Error())
		c.Data["json"] = map[string]interface{}{"code": 0, "message": err.Error(),"data":""}
		c.ServeJSON()
		return
	}

	for i:=0; i<len(cntr); i++ {
		cntr[i].LeaderName = cfunc.GetUserCnName(cntr[i].Leader)
		cntr[i].CompName = cfunc.GetCompCnName(cntr[i].DumdCompEn)
		// 获取关联id
		cntr[i] = getCntrInfo(cntr[i])
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": cntr,
	}
	c.Data["json"] = map[string]interface{}{"code": 1, "message": "数据获取成功！","data":ret}
	c.ServeJSON()
}

func getCntrInfo(info MultiCntrInfo) MultiCntrInfo {
	var mcp models.UnitConfMcp
	initial.DB.Model(models.UnitConfMcp{}).Where("id=?", info.McpConfId).First(&mcp)
	info.Container = mcp.ContainerType

	if mcp.ContainerType == "istio" {
		err, istio := unit_conf.GetIstioConfDetail(info.McpConfId)
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
		err, caas := unit_conf.GetCaasConfDetail(info.McpConfId)
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
		err, rancher := unit_conf.GetRancherConfDetail(info.McpConfId)
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

func getLastCallTime() (interface{}, int) {
	flag := 1
	if !initial.GetCache.IsExist("get_mcp_list_api_req") || common.GetString(initial.GetCache.Get("get_mcp_list_api_req")) == "" {
		initial.GetCache.Put("get_mcp_list_api_req", time.Now().Format("2006-01-02 15:04:05"), 15*time.Minute)
		flag = 0
	}
	return initial.GetCache.Get("get_mcp_list_api_req"), flag
}
