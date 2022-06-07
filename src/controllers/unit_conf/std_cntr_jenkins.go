package unit_conf

import (
	"models"
	"encoding/json"
	"initial"
	"strings"
	"github.com/astaxie/beego"
	"initial/jenkins_xml"
	"fmt"
	"time"
	"library/cfunc"
	"controllers"
)

// @Title 获取标准容器发布单元的构建xml
// @Description 获取标准容器发布单元的构建xml，前端将前置命令、构建命令和后置命令拼装起来，形成可视化页面
// @Param	id	query	string	false	"cntr配置表的id"
// @Param	unit	query	string	false	"发布单元的英文名"
// @Success 200 true or false
// @Failure 403
// @router /cntr/jenk-xml [get]
func (c *StdCntrConfController) JenkXmlList() {
	id := strings.TrimSpace(c.GetString("id"))
	unit := strings.TrimSpace(c.GetString("unit"))
	if id == "" && unit == "" {
		c.SetJson(0, "", "参数输入有误！")
		return
	}
	var cntr models.UnitConfCntr
	if id != "" {
		err := initial.DB.Model(models.UnitConfCntr{}).Where("id=?", id).First(&cntr).Error
		if err != nil {
			beego.Error(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
	} else {
		ui := cfunc.GetUnitInfoByName(unit)
		err := initial.DB.Model(models.UnitConfCntr{}).Where("unit_id=? AND is_delete = 0", ui.Id).First(&cntr).Error
		if err != nil {
			beego.Error(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
	}

	replase_image_shell := jenkins_xml.PRD_REPLACE_IMAGE_SHELL
	harbor_url := jenkins_xml.PRD_HARBOR_URL
	env := beego.AppConfig.String("runmode")
	if env == "dr" {
		replase_image_shell = jenkins_xml.DR_REPLACE_IMAGE_SHELL
		harbor_url = jenkins_xml.DR_HARBOR_URL
	}
	if env != "prd" && env != "dr" {
		replase_image_shell = jenkins_xml.DI_REPLACE_IMAGE_SHELL
		harbor_url = jenkins_xml.DI_HARBOR_URL
	}
	// 改版后租户信息没有及时传递
	if cntr.DeployComp == "" {
		_, mcf := cfunc.GetContainerTypeByUnitId(cntr.UnitId)
		cntr.DeployComp = mcf.DeployComp
	}
	unit_info := cfunc.GetUnitInfoById(cntr.UnitId)
	job_name := fmt.Sprintf("%s-%s-%s-%s", strings.ToLower(cntr.DeployComp), beego.AppConfig.String("runmode"),
		unit_info.Unit, time.Now().Format(initial.DateFormat))
	repo_name := fmt.Sprintf("%s-library/%s", strings.ToLower(cntr.DeployComp), job_name)
	ret := map[string]interface{}{
		"unit_info": unit_info.Info,
		"before": cntr.BeforeXml,
		"build": cntr.BuildXml,
		"after": cntr.AfterXml,
		"jenkins_node": cntr.JenkinsNode,
		"is_confirm": cntr.IsConfirm,
		"replase_image_shell": replase_image_shell,
		"harbor_url": harbor_url,
		"git_url": cntr.GitUrl,
		"repo_name": repo_name,
		"tag": "latest",
	}
	c.SetJson(1, ret, "jenkins配置获取成功！")
}

// @Title 更新标准容器发布单元的构建xml
// @Description 更新标准容器发布单元的构建xml，只保存前置命令、构建命令和后置命令
// @Param	body	body	unit_conf.CntrJenkinsXmlInput	true	"body形式的数据，传入构建命令"
// @Success 200 true or false
// @Failure 403
// @router /cntr/jenk-xml [post]
func (c *StdCntrConfController) JenkXmlEdit() {
	if strings.Contains(c.Role, "admin") == false && c.Role != "deploy-single" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	var xml CntrJenkinsXmlInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &xml)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	var cntr models.UnitConfCntr
	err = initial.DB.Model(models.UnitConfCntr{}).Where("id=?", xml.Id).First(&cntr).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(cntr.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的jenkins构建配置权限，请联系发布单元负责人进行操作！")
		return
	}

	tx := initial.DB.Begin()
	update_map := map[string]interface{}{
		"before_xml": xml.BeforeXml,
		"build_xml": xml.BuildXml,
		"after_xml": xml.AfterXml,
		"jenkins_node": xml.JenkinsNode,
	}
	err = tx.Model(models.UnitConfCntr{}).Where("id=?", xml.Id).Updates(update_map).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "标准容器的jenkin配置维护成功，可以查看构建配置！")
}

type CntrJenkinsXmlInput struct {
	Id        int `json:"id"`
	BeforeXml string `json:"before_xml"`
	BuildXml  string `json:"build_xml"`
	AfterXml  string `json:"after_xml"`
	JenkinsNode string `json:"jenkins_node"`
}

// @Title 标准容器构建xml的确认，确认后才可以构建发布
// @Description 标准容器构建xml的确认，确认后才可以构建发布
// @Param	id	query	string	true	"标准容器配置的id"
// @Success 200 true or false
// @Failure 403
// @router /cntr/confirm [post]
func (c *StdCntrConfController) JenkXmlConfirm() {
	if strings.Contains(c.Role, "admin") == false && c.Role != "deploy-single" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	id := c.GetString("id")

	var cntr models.UnitConfCntr
	err := initial.DB.Model(models.UnitConfCntr{}).Where("id=?", id).First(&cntr).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(cntr.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的jenkins构建配置权限，请联系发布单元负责人进行操作！")
		return
	}

	tx := initial.DB.Begin()
	err = tx.Model(models.UnitConfCntr{}).Where("id=?", id).Update("is_confirm", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "jenkin配置确认成功，可以正常发布！")
}
