package unit_conf

import (
	"controllers"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"library/common"
	"models"
	"strconv"
	"strings"
	"time"
)

// 虚机应用发布单元录入
// @Title 虚机应用录入
// @Description 从发布单元列表选取数据，同时作相关信息确认和维护
// @Param body body models.UnitConfVM true "body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /vm/new [post]
func (c *StdVmConfController) New() {
	if strings.Contains(c.Role, "admin") == false && c.Role != "deploy-single" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	var data models.UnitConfVM
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &data)
	if err != nil {
		beego.Error(err)
		c.SetJson(0, "", "数据解析失败！")
		return
	}

	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(data.UnitID, c.UserId) {
		c.SetJson(0, "", "您没有权限编辑此发布单元，请联系发布单元负责人进行操作！")
		return
	}

	// 重复数据校验
	var repeat int
	if err := initial.DB.Model(&data).Where("`unit_id` = ? AND `is_delete` = 0", data.UnitID).
		Count(&repeat).Error; err != nil {
		c.SetJson(0, "", "查询数据时出错！")
		return
	}
	if repeat != 0 {
		c.SetJson(0, "", "该发布单元信息已维护，请勿重复添加！")
		return
	}

	err = InputCheck(&data)
	if err != nil {
		beego.Error(err)
		c.SetJson(0, "", err.Error())
		return
	}

	data.DeployENV = strings.ToLower(beego.AppConfig.String("runmode"))
	var unit models.UnitConfList
	if err := initial.DB.Model(&unit).Where("`id` = ? AND `is_offline` = 0", data.UnitID).First(&unit).Error; err != nil {
		beego.Error(err)
		c.SetJson(0, "", "查询数据时出错！")
		return
	}
	if unit.Id == 0 {
		c.SetJson(0, "", "未查询到该发布单元！")
		return
	}

	if data.AppPath == "" {
		deployType := strings.ToLower(strings.TrimSpace(data.DeployType))
		deployComp := strings.ToLower(strings.TrimSpace(data.DeployComp))
		deployVPC := strings.ToLower(strings.TrimSpace(data.DeployVPC))

		// 发布单元名格式化为 ***-*** 形式
		unitName := strings.Replace(strings.ToLower(unit.Unit), "_", "-", -1)
		appDir := fmt.Sprintf("%s_%s_%s_%s-%s/apps", deployType, unitName, data.DeployENV, deployComp, deployVPC)
		data.AppPath = "/app/appsystems/" + appDir
	} else if !strings.HasPrefix(data.AppPath, "/app/") {
		c.SetJson(0, "", "应用目录不合法，只允许部署在 /app/ 下子目录中")
		return
	}

	data.AppTempPath = "/tmp"
	//data.AppBackupPath = "/app/backup"
	if data.AppBackupPath == "" {
		data.AppBackupPath = "/app/backup"
	}
	data.InsertTime = time.Now().Format("2006-01-02 15:04:05")

	if data.NeedReboot < 0 || data.NeedReboot > 1 {
		c.SetJson(0, "", "参数值错误！")
		return
	}
	if data.NeedReboot == 1 {
		if data.CMDStartup == "" || data.CMDStop == "" {
			c.SetJson(0, "", "需要重启时，启动/停止命令不能为空！")
			return
		}
	}


	tx := initial.DB.Begin()
	if err := tx.Create(&data).Error; err != nil {
		tx.Rollback()
		beego.Error(err)
		c.SetJson(0, "", err.Error())
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	c.SetJson(1, "", "标准虚机发布单元信息维护成功！")
	return
}

// Todo:完善参数检查
func InputCheck(vm *models.UnitConfVM) error {
	// 数据格式化
	appType := strings.ToLower(strings.TrimSpace(vm.AppType))
	appSubType := strings.ToLower(strings.TrimSpace(vm.AppSubType))
	deployType := strings.ToLower(strings.TrimSpace(vm.DeployType))
	deployComp := strings.ToUpper(strings.TrimSpace(vm.DeployComp))
	appBindPort := strings.TrimSpace(vm.AppBindPort)
	appBindProt := strings.ToLower(strings.TrimSpace(vm.AppBindProt))
	appUser := strings.ToLower(strings.TrimSpace(vm.AppUser))

	if vm.UnitID == 0 {
		return errors.New("发布单元ID不能为0")
	}
	appTypes := []string{"app", "web"}
	if !common.InList(appType, appTypes) {
		return errors.New("不支持的应用类型！ " + vm.AppType)
	}
	appSubTypes := []string{"maven", "ant", "gradle", "node", "simple"}
	if !common.InList(appSubType, appSubTypes) {
		return errors.New("不支持的构建类型！ " + vm.AppSubType)
	}
	deployTypes := []string{"jar", "war", "py2", "py3", "ng"}
	if !common.InList(deployType, deployTypes) {
		return errors.New("不支持的部署类型！ " + vm.DeployType)
	}
	if deployComp == "" {
		return errors.New("应用部署租户不能为空！")
	}
	if cfunc.GetCompCnName(strings.ToUpper(deployComp)) == "" {
		return errors.New("不支持的租户！ " + vm.DeployComp)
	}
	if strings.TrimSpace(vm.DeployVPC) == "" {
		return errors.New("应用部署网络域不能为空！")
	}
	if strings.TrimSpace(vm.GitURL) == "" {
		return errors.New("git 地址不能为空！")
	}
	if vm.Hosts == "" {
		return errors.New("IP地址不能为空")
	}
	ips := strings.Split(strings.TrimSpace(vm.Hosts), ";")
	for _, i := range ips {
		if !common.IsValidIP(i) {
			return errors.New("非法的IP地址！ " + i)
		}
	}
	ports := strings.Split(appBindPort, ";")
	for _, v := range ports {
		port, err := strconv.Atoi(v)
		if err != nil ||
			port < 1025 ||
			port > 65535 {
			return errors.New("非法的端口值！ " + v)
		}
	}
	appProts := []string{"tcp", "udp"}
	if !common.InList(appBindProt, appProts) {
		return errors.New("不支持的端口协议！ " + vm.AppBindProt)
	}
	//if appUser != "mwop" && appUser != "deployop" {
	//	return errors.New("应用用户只能是 mwop或deployop")
	//}
	if appUser == "" {
		return errors.New("应用用户不能为空！")
	}
	//vm.CMDPre = cmdParse(vm.CMDPre)
	//vm.CMDStartup = cmdParse(vm.CMDStartup)
	//vm.CMDStop = cmdParse(vm.CMDStop)
	//vm.CMDRear = cmdParse(vm.CMDRear)

	//if strings.TrimSpace(vm.CMDStartup) == "" {
	//	return errors.New("应用启动命令不能为空！")
	//}
	//if strings.TrimSpace(vm.CMDStop) == "" {
	//	return errors.New("应用停止命令不能为空！")
	//}
	return nil
}

//func cmdParse(cmd string) string {
//	cmdPrefix := "source /etc/profile && source ~/.bashrc && "
//	if strings.TrimSpace(cmd) != "" && !strings.HasPrefix(strings.TrimSpace(cmd), cmdPrefix) {
//		return cmdPrefix + strings.TrimSpace(cmd)
//	}
//	return strings.TrimSpace(cmd)
//}

// 虚机应用发布单元配置更新
// @Title 虚机应用配置更新
// @Description 配置更新
// @Param body body models.UnitConfVM true "body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /vm/update [post]
func (c *StdVmConfController) Update() {
	// Test Pass
	if strings.Contains(c.Role, "admin") == false && c.Role != "deploy-single" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	var data models.UnitConfVM
	//var old models.UnitConfVM
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &data)
	if err != nil {
		beego.Error(err)
		c.SetJson(0, "", "数据解析失败！")
		return
	}

	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(data.UnitID, c.UserId) {
		c.SetJson(0, "", "您没有权限编辑此发布单元，请联系发布单元负责人进行操作！")
		return
	}

	err = InputCheck(&data)
	if err != nil {
		beego.Error(err)
		c.SetJson(0, "", err.Error())
		return
	}
	var count int
	if err := initial.DB.Model(&data).Where("`unit_id` = ? AND `is_delete` = 0", data.UnitID).Count(&count).Error; err != nil {
		beego.Error(err)
		c.SetJson(0, "", "查询数据时出错！")
		return
	}
	if count == 0 {
		c.SetJson(0, "", "未查询到该记录，更新失败！")
		return
	}
	var unit models.UnitConfList
	if err := initial.DB.Model(&unit).Where("`id` = ? AND `is_offline` = 0", data.UnitID).First(&unit).Error; err != nil {
		beego.Error(err)
		c.SetJson(0, "", "查询数据时出错！")
		return
	}
	if unit.Id == 0 {
		c.SetJson(0, "", "未查询到该发布单元！")
		return
	}

	if data.AppPath == "" {
		// 重新拼接应用目录
		unitName := strings.Replace(strings.ToLower(unit.Unit), "_", "-", -1)
		deployType := strings.ToLower(strings.TrimSpace(data.DeployType))
		deployComp := strings.ToLower(strings.TrimSpace(data.DeployComp))
		deployVPC := strings.ToLower(strings.TrimSpace(data.DeployVPC))
		data.DeployENV = strings.ToLower(beego.AppConfig.String("runmode"))
		appDir := fmt.Sprintf("%s_%s_%s_%s-%s", deployType, unitName, data.DeployENV, deployComp, deployVPC)
		data.AppPath = "/app/appsystems/" + appDir
	}

	if data.NeedReboot < 0 || data.NeedReboot > 1 {
		c.SetJson(0, "", "参数值错误！")
		return
	}
	if data.NeedReboot == 1 {
		if data.CMDStartup == "" || data.CMDStop == "" {
			c.SetJson(0, "", "需要重启时，启动/停止命令不能为空！")
			return
		}
	}

	updates := map[string]interface{}{
		"app_type": data.AppType,
		"app_sub_type": data.AppSubType,
		"deploy_type": data.DeployType,
		"git_id": data.GitID,
		"git_unit": data.GitUnit,
		"git_url": data.GitURL,
		"deploy_comp": data.DeployComp,
		"deploy_vpc": data.DeployVPC,
		"artifact": data.Artifact,
		"hosts": data.Hosts,
		"app_user": data.AppUser,
		"app_path": data.AppPath,
		"app_bind_prot": data.AppBindProt,
		"app_bind_port": data.AppBindPort,
		"app_backup_path": data.AppBackupPath,
		"cmd_pre": data.CMDPre,
		"cmd_stop": data.CMDStop,
		"cmd_startup": data.CMDStartup,
		"cmd_rear": data.CMDRear,
		"need_reboot": data.NeedReboot,
	}

	tx := initial.DB.Begin()
	if err := tx.Model(&data).Where("`unit_id` = ? AND `is_delete` = 0", data.UnitID).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, "", "标准虚机发布单元信息更新成功！")
	return
}

// 虚机应用发布单元配置软删除
// @Title 虚机应用配置软删除
// @Description 配置软删除
// @Param id	query	string	true	"标准容器配置的id"
// @Success 200 true or false
// @Failure 403
// @router /vm/delete [delete]
func (c *StdVmConfController) Delete() {
	// TestPass
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作，请联系管理员进行删除！")
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
		c.SetJson(0, "", "您没有此发布单元的Jenkins构建配置权限，请联系发布单元负责人进行操作！")
		return
	}

	tx := initial.DB.Begin()
	err = tx.Model(models.UnitConfVM{}).Where("id=?", id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "删除成功！")
}
