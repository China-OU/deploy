package online

import (
	"time"
	"github.com/astaxie/beego"
	"library/jenkins"
	"initial"
	"models"
	"library/common"
)

type CntrBuild struct {
	Opr           jenkins.JenkOpr
	OnlineId      int
	CntrOnlineId  int
	ImageUrl      string
}

func (c *CntrBuild) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("Cntr Jenkins Build Panic error:", err)
		}
	}()
	timeout := time.After(30 * time.Minute)
	result_ch := make(chan bool, 1)
	go func() {
		result := c.JenkinsBuild()
		result_ch <- result
	}()
	select {
	case <-result_ch:
		beego.Info("cntr构建完成")
	case <-timeout:
		beego.Info("cntr构建超时")
		c.SaveBuildResult(0, 0, 30 * 60, "cntr构建超时", "")
	}
}

func (c *CntrBuild) JenkinsBuild() bool {
	// 初始化
	start := time.Now()
	c.Opr.Init()
	// 删除job
	err := c.Opr.DeleteJob()
	if err != nil {
		beego.Error(err.Error())
		c.SaveBuildResult(0, 0, 0, err.Error(), "")
		return false
	}

	// 新增job
	err = c.Opr.CreateJob()
	if err != nil {
		beego.Error(err.Error())
		c.SaveBuildResult(0, 0, 0, err.Error(), "")
		return false
	}

	// 构建job
	err = c.Opr.BuildJob()
	if err != nil {
		beego.Error(err.Error())
		c.SaveBuildResult(0, 0, 0, err.Error(), "")
		return false
	}

    for {
		// 查询状态
		beego.Info("还未构建完成，请等待20秒。。。")
		time.Sleep(20 * time.Second)
		build, err := c.Opr.GetLastBuild()
		if err != nil {
			beego.Error(err.Error())
			c.SaveBuildResult(0, 0, 0, err.Error(), "")
			return false
		}
		if build.Raw.Result == "" {
			continue
		} else {
			jenk_success := 0
			if build.Raw.Result == "SUCCESS" {
				jenk_success = 1
			}
			cost_time := time.Now().Sub(start).Seconds()
			c.SaveBuildResult(2, jenk_success, common.GetInt(cost_time), "", "")
			return true
		}
	}
	return true
}

func (c *CntrBuild) SaveBuildResult(is_success, jenk_success, cost_time int, err_log, image string) {
	online_update_map := map[string]interface{}{
		"is_success": is_success,
		"error_log": err_log,
	}
	if jenk_success == 1 {
		image = c.ImageUrl
	}
	cntr_online_map := map[string]interface{}{
		"jenkins_name": c.Opr.JobName,
		"jenkins_success": jenk_success,
		"jenkins_image": image,
		"jenkins_cost_time": cost_time,
	}
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineAllList{}).Where("id=?", c.OnlineId).Updates(online_update_map).Error
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Model(models.OnlineStdCntr{}).Where("id=?", c.CntrOnlineId).Updates(cntr_online_map).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}
