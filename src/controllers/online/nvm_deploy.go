package online

import (
	"controllers"
	"high-conc"
	"initial"
	"models"
	"strings"
)

// @Title 非标虚机应用部署
// @Description 非标虚机应用部署
// @Param	online_id	    query	string	true	"标准容器发布单元的id"
// @Success 200 true or false
// @Failure 403
// @router /nvm/deploy [post]
func (c *NvmOnlineController) NvmDeploy() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	var nvm models.OnlineNvm
	var nvm_conf models.UnitConfNvm
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", online_id).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.OnlineNvm{}).Where("online_id=? and is_delete=0", online_id).First(&nvm).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.UnitConfNvm{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&nvm_conf).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的部署权限！")
		return
	}
	if online.IsSuccess == 1 || online.IsSuccess == 2{
		c.SetJson(0, "", "正在发布中或者已发布完成，不允许再次点击！")
		return
	}
	if nvm.FileAddr == "" {
		c.SetJson(0, "", "要更新的文件地址为空，无法部署！")
		return
	}
	// 正在更新中的应用不允许再次更新
	var cnt int
	initial.DB.Model(models.OnlineAllList{}).Where("is_success = 2 and unit_id = ?", online.UnitId).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "该应用正在升级，不允许再次更新！")
		return
	}

	nvm_deploy := NvmDeploy{
		Conf: nvm_conf,
		Online: online,
		Nvm: nvm,
	}
	high_conc.JobQueue <- &nvm_deploy
	c.SetJson(1, "", "已成功进入队列，请耐心等待容器更新结果！")
}

// @Title 非标虚机的升级日志，无权限认证
// @Description 非标虚机的升级日志，无权限认证
// @Param	online_id	    query	string	true	"非标发布单元的id"
// @Success 200 true or false
// @Failure 403
// @router /nvm/shell-log [get]
func (c *NvmOnlineController) NvmShellLog() {
	online_id, _ := c.GetInt("online_id")
	// 获取发布单元名称
	type InfoName struct {
		Info string `json:"info"`
	}
	var info_name InfoName
	err := initial.DB.Table("online_all_list a").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").
		Select("b.info").Where("a.id = ? and a.is_delete=0", online_id).First(&info_name).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	var nvm models.OnlineNvm
	err = initial.DB.Model(models.OnlineNvm{}).Where("online_id=? and is_delete=0", online_id).First(&nvm).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	ret := map[string]interface{}{
		"unit_info": info_name.Info,
		"log": nvm.ShellLog,
	}
	c.SetJson(1, ret, "shell日志获取成功！")
}
