package online

import (
	"github.com/astaxie/beego"
	"initial"
	"library/common"
	"library/jenkins"
	"models"
	"time"
)

type VMBuild struct {
	Opr        jenkins.JenkOpr
	OnlineId   int
	VMOnlineId int // buildLogId
	ArtifactUrl   string
}

func (c *VMBuild) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("VM_APP Jenkins Build Panic error:", err)
		}
	}()
	timeout := time.After(30 * time.Minute)
	resultCh := make(chan bool, 1)
	go func() {
		result := c.JenkinsBuild()
		resultCh <- result
	}()
	select {
	case <-resultCh:
		beego.Info("VM_APP构建完成")
	case <-timeout:
		beego.Info("VM_APP构建超时")
		c.SaveBuildResult( 2, 0, 30*60, "VM_APP构建超时")
	}
}

func (c *VMBuild) JenkinsBuild() bool {
	// 初始化
	start := time.Now()
	c.Opr.Init()
	// 删除job
	err := c.Opr.DeleteJob()
	if err != nil {
		beego.Error(err.Error())
		c.SaveBuildResult(2, 0, 0, err.Error())
		return false
	}

	// 新增job
	err = c.Opr.CreateJob()
	if err != nil {
		beego.Error(err.Error())
		c.SaveBuildResult( 2, 0, 0, err.Error())
		return false
	}

	// 构建job
	err = c.Opr.BuildJob()
	if err != nil {
		beego.Error(err.Error())
		c.SaveBuildResult( 2, 0, 0, err.Error())
		return false
	}

	for {
		// 查询状态
		beego.Info("还未构建完成，请等待20秒。。。")
		time.Sleep(20 * time.Second)
		build, err := c.Opr.GetLastBuild()
		if err != nil {
			beego.Error(err.Error())
			c.SaveBuildResult(2, 0, 0, err.Error())
			return false
		}
		if build.Raw.Result == "" {
			costTime := time.Now().Sub(start).Seconds()
			c.SaveBuildResult(2, 2, common.GetInt(costTime), "")
			continue
		} else {
			isSuccess := 0
			if build.Raw.Result == "SUCCESS" {
				isSuccess = 1
			}
			costTime := time.Now().Sub(start).Seconds()
			c.SaveBuildResult(2, isSuccess, common.GetInt(costTime), "")
			return true
		}
	}
}

func (c *VMBuild) SaveBuildResult(isSuccess, buildStatus, duration int, errLog string) {
	onlineUpdateMap := map[string]interface{}{
		"is_success": isSuccess,
		"error_log": errLog,
	}
	vmOnlineMap := map[string]interface{}{
		"jenkins_name": c.Opr.JobName,
		"build_status": buildStatus,
		"build_duration": duration,
		"artifact_url": c.ArtifactUrl,
	}
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineAllList{}).Where("id = ?", c.OnlineId).Updates(onlineUpdateMap).Error
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Model(models.OnlineStdVM{}).Where("id = ?", c.VMOnlineId).Updates(vmOnlineMap).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}
