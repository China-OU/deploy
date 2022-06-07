package operation

import (
	"library/mcp"
	"time"
	"library/common"
	"initial"
	"models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"fmt"
	"encoding/json"
	"strings"
)

// rancher升级进程
type McpRancherUpgrade struct {
	Opr      mcp.McpRancherOpr  // 基础配置
	UnitId   int            // 更新的发布单元
	Image    string         // 更新后的镜像
	Operator string         // 操作人员
	RecordId   int            // 操作记录ID，不需要传入，内部生成
	SourceId string         // 外部来源id
	ParentId   int          // 内部关联id，和部署主记录关联
}

func (c *McpRancherUpgrade) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("Mcp rancher Upgrade Panic error:", err)
		}
	}()
	timeout := time.After(20 * time.Minute)
	run_env := beego.AppConfig.String("runmode")
	if run_env != "prd" {
		// 测试环境缩容器更新短超时时间
		timeout = time.After(8 * time.Minute)
	}
	result_ch := make(chan bool, 1)
	go func() {
		result := c.UpgradeImage()
		result_ch <- result
	}()
	select {
	case <-result_ch:
		beego.Info("执行完成")
		c.UpdateExecResult()
	case <-timeout:
		beego.Info("执行超时")
		c.SaveExecResult(false, run_env + "环境执行超时，容器状态异常，请上caas平台查看", 20 * 60)
		time.Sleep(1*time.Second)
		c.UpdateExecResult()
	}
}

func (c *McpRancherUpgrade) UpgradeImage() bool {
	err, detail := c.Opr.GetRancherService()
	if err != nil {
		// 数据库没记录，无法记录到数据库中
		beego.Error(err.Error())
		return false
	}
	old_image := detail.LaunchConfig["imageUuid"].(string)
	if old_image == "" {
		beego.Error("容器当前镜像无法获取！")
		return false
	}
	id := c.InsertRecord(strings.TrimLeft(old_image, "docker:"))
	if id == 0 {
		beego.Error("caas升级数据录入失败！")
		return false
	}
	c.RecordId = id
	// 判断是否满足升级条件
	if detail.State != "active" {
		c.SaveExecResult(false, fmt.Sprintf("应用%s处于非active状态,无法进行升级!", detail.Name), 0)
		return false
	}
	if detail.Actions.FinishUpgrade != "" {
		c.SaveExecResult(false, fmt.Sprintf("应用%s未完成升级，请先升级完成后再升级!", detail.Name), 0)
		return false
	}
	// 关联外部表
	c.RelRecord()
	// 镜像不管是同名镜像，都可以直接升级，共用一个接口
	c.TheImageUpgrade(detail)
	return true
}

func (c *McpRancherUpgrade) TheImageUpgrade(detail mcp.RancherService) {
	start := time.Now()
	launch_config := detail.LaunchConfig
	launch_config["imageUuid"] = "docker:" + c.Image
	in_service_strategy := map[string]interface{}{
		"launchConfig": launch_config,
	}
	upgrade_data := map[string]interface{}{
		"inServiceStrategy": in_service_strategy,
	}
	upgrade_byte, _ := json.Marshal(upgrade_data)
	err, upgrade_ret := c.Opr.ActionRancherService(detail.Actions.Upgrade, upgrade_byte)
	if err != nil {
		beego.Info(err.Error())
		c.SaveExecResult(false, err.Error(), 0)
		return
	}
	if upgrade_ret.State != "upgrading" {
		c.SaveExecResult(false, "更新的结果错误，应该是upgrading状态", 0)
		return
	}
	// 升级是瞬时操作，要判断是否升级完成
	time.Sleep(40 * time.Second)
	ec := 0
	finish_url := ""
	for {
		ec += 1
		if ec > 50 {
			// 设置执行次数
			cost_time := time.Now().Sub(start).Seconds()
			c.SaveExecResult(false, "执行超时，容器状态异常，请上容器平台检查！", common.GetInt(cost_time))
			return
		}
		beego.Info("还未升级完成，请等待20秒。。。")
		time.Sleep(20 * time.Second)
		err, detail := c.Opr.GetRancherService()
		if err != nil {
			beego.Info(err.Error())
			c.SaveExecResult(false, err.Error(), 0)
			return
		}
		if detail.Actions.FinishUpgrade != "" {
			beego.Info(detail.State)
			finish_url = detail.Actions.FinishUpgrade
			break
		}
	}
	err, finish_ret := c.Opr.ActionRancherService(finish_url, []byte{})
	if err != nil {
		beego.Info(err.Error())
		c.SaveExecResult(false, err.Error(), 0)
		return
	}
	if finish_ret.State != "finishing-upgrade" {
		c.SaveExecResult(false, "完成升级的结果错误，应该是finishing-upgrade状态", 0)
		return
	}
	time.Sleep(5 * time.Second)
	cost_time := time.Now().Sub(start).Seconds()
	c.SaveExecResult(true, detail.Name + "镜像更新成功！", common.GetInt(cost_time))
	return
}

// 数据记录与更新
func (c *McpRancherUpgrade) InsertRecord(old_image string) int {
	var record models.McpUpgradeList
	record.UnitId = c.UnitId
	record.OldImage = old_image
	record.NewImage = c.Image
	record.Result = 2
	record.Operator = c.Operator
	now := time.Now()
	today := now.Format(initial.DateFormat)
	if now.Hour() < 4 {
		today = now.AddDate(0, 0, -1).Format(initial.DateFormat)
	}
	record.OnlineDate = today
	record.CostTime = 0
	record.InsertTime = now.Format(initial.DatetimeFormat)
	record.SourceId = c.SourceId
	tx := initial.DB.Begin()
	err := tx.Create(&record).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return 0
	}
	tx.Commit()
	return record.Id
}

func (c *McpRancherUpgrade) SaveExecResult(result bool, msg string, cost int) {
	int_result := 0
	if result {
		int_result = 1
	}
	update_map := map[string]interface{}{
		"result": int_result,
		"message": msg,
		"cost_time": cost,
	}
	tx := initial.DB.Begin()
	err := tx.Model(models.McpUpgradeList{}).Where("id=?", c.RecordId).Updates(update_map).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()
}

// 关联到升级父表
func (c *McpRancherUpgrade) RelRecord() {
	if c.ParentId > 0 {
		tx := initial.DB.Begin()
		err := tx.Model(models.OnlineStdCntr{}).Where("id=?", c.ParentId).Update("opr_cntr_id", c.RecordId).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			return
		}
		tx.Commit()
	}
}

// 结果返回到pms或者更新到主表
func (c *McpRancherUpgrade) UpdateExecResult() {
	if c.ParentId > 0 {
		// 更新主表
		tx := initial.DB.Begin()
		var sub_info models.OnlineStdCntr
		err := tx.Model(models.OnlineStdCntr{}).Where("id=?", c.ParentId).First(&sub_info).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			return
		}

		var opr models.McpUpgradeList
		err = tx.Model(models.McpUpgradeList{}).Where("id=?", c.RecordId).First(&opr).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			return
		}
		update_map := map[string]interface{}{
			"is_success": opr.Result,
			"error_log": opr.Message,
		}
		err = tx.Model(models.OnlineAllList{}).Where("id=?", sub_info.OnlineId).Updates(update_map).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			return
		}
		tx.Commit()

		// 回推发布管理系统。通过发布管理系统拉取时，只更新主表的source_id即可。
		// 升级子表的source_id，用于cpds或者devops直接调用。两者不宜混淆
		var record models.OnlineAllList
		err = initial.DB.Model(models.OnlineAllList{}).Where("id=?", sub_info.OnlineId).First(&record).Error
		if err != nil {
			beego.Error(err.Error())
			return
		}
		if record.SourceId != "" && record.SourceId != "0" {
			req := httplib.Get(beego.AppConfig.String("pms_baseurl") + "/mdp/release/result")
			req.Header("Authorization", "Basic mdeploy_d8c8680d046b1c60e63657deb3ce6d89")
			req.Header("Content-Type", "application/json")
			req.Param("record_id", record.SourceId)
			req.Param("result", common.GetString(opr.Result))
			_, err := req.String()
			if err != nil {
				beego.Error(err.Error())
			}
		}
	}
}
