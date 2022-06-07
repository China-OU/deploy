package unit_conf

import (
	"controllers"
	"encoding/json"
	"initial"
	"library/cfunc"
	"models"
	"strings"
)

type VMJenkinsXMLInput struct {
	ID			int		`json:"id"`
	JenkinsNode	string	`json:"jenkins_node"`
	BeforeXml	string	`json:"before_xml"`
	BuildXml	string	`json:"build_xml"`
	AfterXml	string	`json:"after_xml"`
	PkgPath		string	`json:"pkg_path"`
} 

// @Title 查看主机应用构建配置
// @Description 查看主机应用构建配置详细信息
// @Param	unit	query	string	false	"发布单元英文名"
// @Success	200	true or false
// @Failure	403
// @router /vm/jenk/xml [get]
func (c *StdVmConfController) JenkinsXmlCheck() {
	unitName := strings.TrimSpace(c.GetString("unit"))
	if unitName == "" {
		c.SetJson(0, "", "参数不能为空！")
		return
	}

	var unitConf models.UnitConfVM
	conn := initial.DB
	unit := cfunc.GetUnitInfoByName(unitName)
	if err := conn.Model(&unitConf).Where("`unit_id` = ? AND `is_delete` = 0", unit.Id).First(&unitConf).Error;
	err != nil || unitConf.ID == 0 {
		c.SetJson(0, "", "未查询到该发布单元记录！")
		return
	}
	unit = cfunc.GetUnitInfoById(unitConf.UnitID)
	unitName += "(" + unit.Name + ")"
	jenkinsConf := map[string]interface{}{
		"jenkins_node": unitConf.JenkinsNode,
		"git_url": unitConf.GitURL,
		"before_xml": unitConf.BeforeXml,
		"build_xml": unitConf.BuildXml,
		"after_xml": unitConf.AfterXml,
		"pkg_path": unitConf.ArtifactPath,
		"is_confirm": unitConf.IsConfirm,
		"unit_name": unitName,
	}
	c.SetJson(1, jenkinsConf, "Jenkins配置获取成功！")
	return
}

// @Title 虚机构建确认
// @Description 构建配置确认，确认后才可以构建发布
// @Param	id	query	string	true	"虚机配置id"
// @Success 200 true or false
// @Failure 403
// @router /vm/jenk/confirm [post]
func (c *StdVmConfController) JenkinsXmlConfirm() {
	// Test Pass
	if strings.Contains(c.Role, "admin") == false && c.Role != "deploy-single" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	id := c.GetString("id")

	var vm models.UnitConfVM
	err := initial.DB.Model(models.UnitConfVM{}).Where("id=?", id).First(&vm).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(vm.UnitID, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的jenkins构建配置权限，请联系发布单元负责人进行操作！")
		return
	}

	if vm.ArtifactPath == "" {
		c.SetJson(0, "", "Jenkins构建配置错误，请检查！")
		return
	}

	path := strings.Trim(vm.ArtifactPath, " ")
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".war"){
		pathArr := strings.Split(path, "/")
		if pathArr[len(pathArr) - 1] != vm.Artifact {
			c.SetJson(0, "", "Jenkins构建配置错误，请检查！")
			return
		}
	}
	tx := initial.DB.Begin()
	err = tx.Model(models.UnitConfVM{}).Where("id=?", id).Update("is_confirm", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "Jenkins配置确认成功，可以正常发布！")
}

// @Title 更新虚机应用发布单元的构建xml
// @Description 更新虚机应用发布单元的构建xml，只保存前置命令、构建命令和后置命令
// @Param	body	body	unit_conf.CntrJenkinsXmlInput	true	"body形式的数据，传入构建命令"
// @Success 200 true or false
// @Failure 403
// @router /vm/jenk/update [post]
func (c *StdVmConfController) JenkinsXmlEdit() {
	// Test Pass
	if strings.Contains(c.Role, "admin") == false && c.Role != "deploy-single" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	var xml VMJenkinsXMLInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &xml)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	var vm models.UnitConfVM
	err = initial.DB.Model(models.UnitConfVM{}).Where("id=?", xml.ID).First(&vm).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(vm.UnitID, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的jenkins构建配置权限，请联系发布单元负责人进行操作！")
		return
	}

	if xml.PkgPath == "" {
		c.SetJson(0, "", "jenkins构建配置中应用包位置不能为空！")
		return
	}

	path := strings.Trim(xml.PkgPath, " ")
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".war"){
		pathArr := strings.Split(path, "/")
		if pathArr[len(pathArr) - 1] != vm.Artifact {
			c.SetJson(0, "", "Jenkins构建配置中工作空间包路径包含的文件名和配置中指定的发布文件名不一致！")
			return
		}
	}

	tx := initial.DB.Begin()
	updateMap := map[string]interface{}{
		"before_xml":   xml.BeforeXml,
		"build_xml":    xml.BuildXml,
		"after_xml":    xml.AfterXml,
		"jenkins_node": xml.JenkinsNode,
		"artifact_path": xml.PkgPath,
	}
	err = tx.Model(models.UnitConfVM{}).Where("id=?", xml.ID).Updates(updateMap).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "虚机应用Jenkins配置维护成功，可以查看构建配置！")
}