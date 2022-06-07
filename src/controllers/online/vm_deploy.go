package online

import (
	"controllers"
	"github.com/astaxie/beego"
	"github.com/jinzhu/gorm"
	high_conc "high-conc"
	"initial"
	"library/common"
	"models"
	"strings"
	"fmt"
)

// @Title 虚机应用更新
// @Description 虚机应用更新
// @Param online_id query string true "vm上线ID"
// @Success 200 true or false
// @Failure 403
// @router /vm/upgrade [post]
func (c *StdVmOnlineController) VmUpgrade() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	onlineID := c.GetString("online_id")
	onlineID = strings.TrimSpace(onlineID)
	if onlineID == "" {
		c.SetJson(0, "", "虚机应用上线ID为空！")
		return
	}
	var (
		err  error
		onlineAll models.OnlineAllList
		onlineVM models.OnlineStdVM
		vmConf models.UnitConfVM
	)
	var msg string
	if err = initial.DB.Model(models.OnlineAllList{}).Where("id = ? AND is_delete = 0",onlineID).First(&onlineAll).Error ; err != nil {
		if err == gorm.ErrRecordNotFound {
			msg = "发布任务不存在！"
		}else{
			msg = err.Error()
		}
		c.SetJson(0, "", msg)
		return
	}
	if err = initial.DB.Model(&models.OnlineStdVM{}).Where("online_id = ? AND is_delete = 0", onlineAll.Id).First(&onlineVM).Error ; err != nil {
		if err == gorm.ErrRecordNotFound {
			msg = "虚机应用发布记录不存在！"
		}else{
			msg = err.Error()
		}
		c.SetJson(0, "", msg)
		return
	}
	if err = initial.DB.Model(&models.UnitConfVM{}).Where("is_delete = 0 AND unit_id = ?", onlineAll.UnitId).First(&vmConf).Error ; err != nil {
		if err == gorm.ErrRecordNotFound {
			msg = "发布单元配置不存在！"
		}else{
			msg = err.Error()
		}
		c.SetJson(0, "", msg)
		return
	}

	if c.Role == "deploy-single" {
		auth_flag := controllers.CheckUnitSingleAuth(onlineAll.UnitId, c.UserId)
		if !auth_flag {
			c.SetJson(0, "", "您没有权限升级此发布单元，只有此发布单元的负责人、开发人员和测试人员才可以升级！")
			return
		}
	}

	if onlineAll.OnlineTime != "" && beego.AppConfig.String("runmode") == "prd" {
		// 只有生产环境才验证发布时间
		ready := common.ReadyToRelease(onlineAll.OnlineTime, initial.DateSepLine)
		if !ready {
			c.SetJson(0, "", fmt.Sprintf("发布时间为%s,请在指定时间进行更新!", onlineAll.OnlineTime))
			return
		}
	}

	if onlineVM.BuildStatus != 1 {
		c.SetJson(0, "", "未构建成功的任务不允许升级！")
		return
	}

	if onlineVM.UpgradeStatus == 2 {
		c.SetJson(0, "", "该虚机应用正在升级中，请稍等！")
		return
	}

	if onlineVM.UpgradeStatus == 1 && onlineAll.IsSuccess == 1 {
		c.SetJson(0, "", "已升级成功的任务不允许重复发布！")
		return
	}

	var count int
	filtration := fmt.Sprintf("a.is_delete = 0 AND a.upgrade_status = 2 AND b.unit_id = %d",onlineAll.UnitId)
	if err = initial.DB.Table("online_std_vm a").Joins("LEFT JOIN online_all_list b ON a.online_id = b.id").Select("a.*").Where(filtration).Count(&count).Error ; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if count > 0 {
		c.SetJson(0, "", "该发布单元已有发布任务在进行，请勿重复发布！")
		return
	}

	dpVM := VmDeploy{
		Conf:		vmConf,
		OnlineAll:	onlineAll,
		OnlineVM:	onlineVM,
	}

	high_conc.JobQueue <- &dpVM
	dataAll := map[string]interface{}{
		"is_success": 2,
		"error_log": "",
	}
	dataVm := map[string]interface{}{
		"upgrade_status": 2,
		"upgrade_logs": "",
		"upgrade_duration": 0,
	}
	if err := SaveVmDpRest(onlineAll, dataAll, onlineVM, dataVm); err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", "状态更新失败！")
		return
	}
	c.SetJson(1, "", "任务进入发布队列，请稍等！")
	return
}

type LogData struct {
	TaskID	string	`json:"task_id"`
	Logs	string	`json:"logs"`
}
// @Title 查看虚机发布日志
// @Description 查看虚机发布日志
// @Param task_id path int true "发布任务ID"
// @Success 200 true or false
// @Failure 403
// @router /vm/log/:task_id [get]
func (c *StdVmOnlineController) VMLog() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	taskID := c.Ctx.Input.Param(":task_id")
	var taskVm models.OnlineStdVM
	if err := initial.DB.Model(&taskVm).Where("online_id = ? AND is_delete = 0", taskID).First(&taskVm).Error ; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	var data LogData
	data.TaskID = taskID
	data.Logs = taskVm.UpgradeLogs
	c.SetJson(1, data, "")
	return
}