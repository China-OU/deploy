package ext

import (
	"controllers/unit_conf"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"models"
	"strings"
)

type DeployInfoController struct {
	beego.Controller
}

func(c *DeployInfoController) URLMapping()  {
	c.Mapping("GetUnitDeployConf", c.GetUnitDeployConf)
}

type CntrInfo struct {
	//发布单元信息
	Unit string `json:"unit"`
	Name string `json:"name"`
	Leader string `json:"leader"`
	LeaderName   string    `json:"leader_name"`
	AppType		 string    `json:"app_type"`
	AppSubType   string    `json:"app_sub_type"`
	GitUnit      string    `json:"git_unit"`
	GitUrl		 string	   `json:"git_url"`

	// 取容器部署配置
	Container    string    `json:"container"`
	Team         string    `json:"team"`
	Stack        string    `json:"stack"`
	Service      string    `json:"service"`
	DeployComp   string    `json:"deploy_comp"`
	DeployNetwork string   `json:"deploy_network"`

	//取子系统部分参数
	DumdCompEn string `json:"dumd_comp_en"`
	CompName     string    `json:"comp_name"`
	DumdSubSysname string `json:"dumd_sub_sysname"`
	DumdSubSysnameCn string `json:"dumd_sub_sysname_cn"`
	McpConfId     int       `json:"-"`
}

type VmInfo struct {
	//发布单元信息配置
	Unit string `json:"unit"`
	Name string `json:"name"`
	Leader string `json:"leader"`
	LeaderName   string    `json:"leader_name"`
	AppType		 string    `json:"app_type"`
	AppSubType   string    `json:"app_sub_type"`
	DeployType	 string	   `json:"deploy_type"`
	GitUnit      string    `json:"git_unit"`
	GitUrl		 string	   `json:"git_url"`

	//部署配置信息
	DeployComp   string    `json:"deploy_comp"`
	DeployVpc    string    `json:"deploy_vpc"`
	Artifact	 string	   `json:"artifact"`
	Host	 	 string	   `json:"host"`
	AppUser	 	 string	   `json:"app_user"`
	AppPath	 	 string	   `json:"app_path"`
	AppBindPort	 string	   `json:"app_bind_port"`
	AppBackupPath	 string	   `json:"app_backup_path"`
	CmdStartup	 	 string	   `json:"cmd_startup"`
	CmdStop	 	 string	   `json:"cmd_stop"`

	//取子系统部分参数
	DumdCompEn string `json:"dumd_comp_en"`
	CompName     string    `json:"comp_name"`
	DumdSubSysname string `json:"dumd_sub_sysname"`
	DumdSubSysnameCn string `json:"dumd_sub_sysname_cn"`
}

type DbInfo struct {
	//发布单元信息配置
	Unit string `json:"unit"`
	Name string `json:"name"`
	Leader string `json:"leader"`
	LeaderName   string    `json:"leader_name"`
	Type		 string    `json:"type"`
	GitUnit      string    `json:"git_unit"`
	GitUrl		 string	   `json:"git_url"`

	//部署信息
	Username   	 string    `json:"username"`
	Host	 	 string	   `json:"host"`
	Port	 	 string	   `json:"port"`
	Dbname		 string    `json:"dbname"`
	DeployComp   string    `json:"deploy_comp"`

	//取子系统部分参数
	DumdCompEn string `json:"dumd_comp_en"`
	CompName     string    `json:"comp_name"`
	DumdSubSysname string `json:"dumd_sub_sysname"`
	DumdSubSysnameCn string `json:"dumd_sub_sysname_cn"`
}

// GetUnitDeployConf 方法
// @Title GetUnitDeployConf
// @Description 获取所有发布单元部署配置信息
// @Param	en_name	query	string	true	"发布单元英文名，唯一匹配"
// @Success 200 {object} ext.CntrInfo
// @Failure 403
// @router /unit/deployconf/search [get]
func (c *DeployInfoController) GetUnitDeployConf() {
	//静态token校验
	header := c.Ctx.Request.Header
	auth := ""
	if header["Authorization"] != nil && len(header["Authorization"]) > 0 {
		auth = header["Authorization"][0]
	} else {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "没有header!","data":[]MultiCntrInfo{}}
		c.ServeJSON()
		return
	}
	if strings.Replace(auth, "Basic ", "", -1) != "mdeploy_jrTQrrldodKV81FZ4gqrE3CaTLqqvyG7" {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "header校验失败!","data":[]MultiCntrInfo{}}
		c.ServeJSON()
		return
	}

	en_name := c.GetString("en_name")
	cond := " is_delete=0 "
	if strings.TrimSpace(en_name) != "" {
		cond += fmt.Sprintf(" and b.unit = '%s' ", en_name)
	}
	//发布单元校验
	var unit_cnt int
	var unit_list models.UnitConfList
	err := initial.DB.Model(models.UnitConfList{}).Where("is_offline=0 and unit = ?",en_name).Count(&unit_cnt).Find(&unit_list).Error
	if err != nil && err.Error() != "record not found" {
		beego.Info(err.Error())
		c.Data["json"] = map[string]interface{}{"code": 0, "message": err.Error(),"data":""}
		c.ServeJSON()
		return
	}
	if unit_cnt == 0 {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "部署平台无此发布单元。","data":""}
		c.ServeJSON()
		return
	}
	var cntr_cnt int
	var cntr CntrInfo
	err = initial.DB.Table("unit_conf_cntr a").Select("a.*, b.unit, b.name, b.leader, b.dumd_comp_en," +
		" b.dumd_sub_sysname, b.dumd_sub_sysname_cn").Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond+" and ( a.mcp_conf_id > 0 or a.service_name = '') ").Count(&cntr_cnt).Find(&cntr).Error
	if err != nil && err.Error() != "record not found" {
		beego.Info(err.Error())
		c.Data["json"] = map[string]interface{}{"code": 0, "message": err.Error(),"data":""}
		c.ServeJSON()
		return
	}
	if cntr_cnt != 0 {
		cntr.LeaderName = cfunc.GetUserCnName(cntr.Leader)
		cntr.CompName = cfunc.GetCompCnName(cntr.DumdCompEn)
		// 获取关联id
		cntr = getCntrDetail(cntr)
		ret := map[string]interface{}{
			"deploy_type": "container",
			"data": cntr,
		}
		c.Data["json"] = map[string]interface{}{"code": 1, "message": "数据获取成功！","data":ret}
		c.ServeJSON()
		return
	}
	var vm_cnt int
	var vm VmInfo
	err = initial.DB.Table("unit_conf_vm a").Select("a.*, b.unit, b.name, b.leader, b.dumd_comp_en," +
		" b.dumd_sub_sysname, b.dumd_sub_sysname_cn").Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond).Count(&vm_cnt).Find(&vm).Error
	if err != nil && err.Error() != "record not found" {
		beego.Info(err.Error())
		c.Data["json"] = map[string]interface{}{"code": 0, "message": err.Error(),"data":""}
		c.ServeJSON()
		return
	}
	if vm_cnt != 0 {
		vm.LeaderName = cfunc.GetUserCnName(vm.Leader)
		vm.CompName = cfunc.GetCompCnName(vm.DumdCompEn)
		ret := map[string]interface{}{
			"deploy_type": "vm",
			"data": vm,
		}
		c.Data["json"] = map[string]interface{}{"code": 1, "message": "数据获取成功！","data":ret}
		c.ServeJSON()
		return
	}
	var db_cnt int
	var db DbInfo
	err = initial.DB.Table("unit_conf_db a").Select("a.*, b.unit, b.name, b.leader, b.dumd_comp_en," +
		" b.dumd_sub_sysname, b.dumd_sub_sysname_cn").Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond).Count(&db_cnt).Find(&db).Error
	if err != nil && err.Error() != "record not found" {
		beego.Info(err.Error())
		c.Data["json"] = map[string]interface{}{"code": 0, "message": err.Error(),"data":""}
		c.ServeJSON()
		return
	}
	if db_cnt != 0 {
		db.LeaderName = cfunc.GetUserCnName(vm.Leader)
		db.CompName = cfunc.GetCompCnName(vm.DumdCompEn)
		ret := map[string]interface{}{
			"deploy_type": "db",
			"data": db,
		}
		c.Data["json"] = map[string]interface{}{"code": 1, "message": "数据获取成功！","data":ret}
		c.ServeJSON()
		return
	}
	if cntr_cnt == 0 && vm_cnt == 0 && db_cnt == 0 {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "该发布单元未配置部署信息。","data":""}
		c.ServeJSON()
		return
	}

}

func getCntrDetail(info CntrInfo) CntrInfo {
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
