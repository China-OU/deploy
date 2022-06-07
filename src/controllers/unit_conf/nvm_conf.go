package unit_conf

import (
	"controllers"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/common"
	"models"
	"strings"
	"time"
)

type NoStdVmConfController struct {
	controllers.BaseController
}

func (c *NoStdVmConfController) URLMapping() {
	c.Mapping("NvmConfList", c.NvmConfList)
	c.Mapping("NvmConfEdit", c.NvmConfEdit)
	c.Mapping("NvmConfDel", c.NvmConfDel)
}

// @Title NvmConfList
// @Description 获取非标准虚机的配置列表
// @Param	unit_name	query   string  false   "发布单元英文名，支持模糊搜索"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} models.UnitConfNvm
// @Failure 403
// @router /nvm/list [get]
func (c *NoStdVmConfController) NvmConfList() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	unit_name := c.GetString("unit_name")
	page, rows := c.GetPageRows()
	cond := " a.is_delete=0 "
	if strings.TrimSpace(unit_name) != "" {
		cond += fmt.Sprintf(" and b.unit like '%%%s%%' ", unit_name)
	}

	type NvmInfo struct {
		models.UnitConfNvm
		Unit string `json:"unit"`
		Name string `json:"name"`
	}
	var cnt int
	var nvm []NvmInfo
	err := initial.DB.Table("unit_conf_nvm a").Select("a.*, b.unit, b.name").
		Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond).Count(&cnt).Order("a.id desc").Offset((page - 1)*rows).Limit(rows).Find(&nvm).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	ret := map[string]interface{}{
		"cnt": cnt,
		"data": nvm,
	}
	c.SetJson(1, ret, "获取数据成功！")
}

// @Title 非标虚机应用配置录入和修改
// @Description 非标虚机应用配置录入和修改
// @Param   body   body   models.UnitConfNvm   true   "body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /nvm/edit [post]
func (c *NoStdVmConfController) NvmConfEdit() {
	if c.Role == "guest" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	if beego.AppConfig.String("runmode") == "prd" && strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "生产环境权限收缩，您没有权限操作！")
		return
	}
	var nvm models.UnitConfNvm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &nvm)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	nvm.Hosts = strings.TrimSpace(nvm.Hosts)
	nvm.AppUser = strings.TrimSpace(nvm.AppUser)
	nvm.ShellPath = strings.TrimSpace(nvm.ShellPath)
	nvm.FilePath = strings.TrimSpace(nvm.FilePath)
	for _, v := range strings.Split(nvm.Hosts, ";") {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if !common.CheckIp(v) {
			c.SetJson(0, "", fmt.Sprintf("ip地址%s不误，请重新填写！", v))
			return
		}
	}
	if nvm.UnitId == 0 {
		c.SetJson(0, "", "发布单元为空，请修改！")
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(nvm.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的编辑权限，请联系发布单元负责人进行操作！")
		return
	}
	var cnt int
	initial.DB.Model(models.UnitConfNvm{}).Where("id != ? and unit_id = ? and is_delete = 0", nvm.ID,
		nvm.UnitId).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "此发布单元已经创建，不能重复创建。如果需要创建备份发布单元，请在主列表页录入新发布单元！")
		return
	}
	if nvm.ID > 0 {
		update_map := map[string]interface{}{
			"hosts": nvm.Hosts,
			"app_user" : nvm.AppUser,
			"file_path": nvm.FilePath,
			"shell_content": nvm.ShellContent,
			"shell_path": nvm.ShellPath,
			"path_or_gene": nvm.PathOrGene,
		}
		err := initial.DB.Model(models.UnitConfNvm{}).Where("id=?", nvm.ID).Updates(update_map).Error
		if err != nil {
			beego.Error(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
	} else {
		nvm.InsertTime = time.Now().Format(initial.DatetimeFormat)
		err := initial.DB.Create(&nvm).Error
		if err != nil {
			beego.Error(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
	}
	c.SetJson(1, "", "非标虚机配置成功！")
	return
}

// 虚机应用发布单元配置软删除
// @Title 虚机应用配置软删除
// @Description 配置软删除
// @Param   id	query	string	true	"标准容器配置的id"
// @Success 200 true or false
// @Failure 403
// @router /nvm/delete [delete]
func (c *NoStdVmConfController) NvmConfDel() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作，请联系管理员进行删除！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	var nvm models.UnitConfNvm
	err = initial.DB.Model(models.UnitConfNvm{}).Where("id=?", id).First(&nvm).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.UnitConfNvm{}).Where("id=?", id).Update("is_delete", 1).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, "", "删除成功！")
}