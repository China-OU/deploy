package caas_cntr

import (
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"github.com/jinzhu/gorm"
	"initial"
	"library/caas"
	"library/common"
	"models"
	"strings"
	"time"
)

type CntrEdit struct {
	Operator string // 操作人员
	WebData  caas.InitServiceWebData
	Opr      caas.CaasOpr // 基础配置
}

func (c *CntrEdit) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("Cntr edit Panic error:", err)
		}
	}()
	initLog, err := c.updateLog()
	if err != nil {
		beego.Error("获取记录失败:", err.Error())
		return
	}
	startAt := time.Now()
	err = c.EditService(initLog)
	cost := time.Now().Sub(startAt).Seconds()
	costTime := common.GetInt(cost)
	if err != nil {
		// 超时
		if strings.Contains(err.Error(), "超时") {
			if err := initLog.UpdateLog(1, err.Error(), "deactive", costTime); err != nil {
				beego.Error(err.Error())
			}
		}
		// 失败
		err = initLog.UpdateLog(2, "更新失败："+err.Error(), "", costTime)
		if err != nil {
			beego.Error(err.Error())
		}
		return
	}
	err = initLog.UpdateLog(1, "更新成功", "active", costTime)
	if err != nil {
		beego.Error(err.Error())
	}
}

func (c *CntrEdit) EditService(initLog *models.OprCntrInit) error {
	initData := caas.InitServiceData{
		AgentConf:          c.Opr.AgentConf,
		InitServiceWebData: c.WebData,
	}
	if err := initData.RetryInitService(5); err != nil {
		return errors.New("caas更新配置服务接口调用失败！" + err.Error())
	}
	ec := 0
	for {
		ec += 1
		if ec > 50 {
			return errors.New("更新服务配置等待超时")
		}
		time.Sleep(30 * time.Second)
		err, detail := c.Opr.RetryGetService(5)
		if err != nil {
			beego.Info(err.Error())
			return errors.New("caas获取服务详情接口调用失败！" + err.Error())
		}
		if detail.State == "active" {
			return nil
		} else if len(detail.State) > 0 {
			if err := initLog.UpdateLog(1, "更新中", detail.State, ec*30); err != nil {
				beego.Error(err)
			}
		}
	}
}

const (
	imageOpr       = "镜像"
	instanceNumOpr = "实例数"
	resLimitOpr    = "CPU内存限制"
	envOpr         = "环境变量"
	healthCheckOpr = "健康检查"
	logConfigOpr   = "日志规则"
	volumeOpr      = "存储卷"
	schedulerOpr   = "调度策略"
)

func (c *CntrEdit) updateLog() (*models.OprCntrInit, error) {
	var cntrInit models.OprCntrInit

	if err := initial.DB.Model(&models.OprCntrInit{}).First(&cntrInit, "unit_id=?", c.WebData.UnitId).Error; err != nil {
		return nil, err
	}
	healthCheckJson, _ := json.Marshal(c.WebData.HealthCheck)
	envMapJson, _ := json.Marshal(c.WebData.Environment)
	logConfig, _ := json.Marshal(c.WebData.LogConfig)
	volume, _ := json.Marshal(c.WebData.Volume)
	scheduler, _ := json.Marshal(c.WebData.Scheduler)
	tx := initial.DB.Begin()
	updateMap := make(map[string]interface{}, 7)
	updateMap = map[string]interface{}{
		"cpu":       c.WebData.Cpu * 1000,
		"mem_limit": c.WebData.MemLimit,
		"operator":  c.Operator,
		"result":    0,
		"message":   "更新中",
		"cost_time": 0,
		"state":     "",
		"is_edit":   true,
	}
	if cntrInit.Image != c.WebData.Image {
		if err := c.insertOprRecord(tx, imageOpr, c.WebData.Image, cntrInit.Image, cntrInit.Id); err != nil {
			tx.Rollback()
			return nil, err
		}
		updateMap["image"] = c.WebData.Image
	}

	if cntrInit.InstanceNum != c.WebData.InstanceNum {
		if err := c.insertOprRecord(tx, instanceNumOpr,
			common.GetString(c.WebData.InstanceNum),
			common.GetString(cntrInit.InstanceNum), cntrInit.Id); err != nil {
			tx.Rollback()
			return nil, err
		}
		updateMap["instance_num"] = c.WebData.InstanceNum
	}
	if cntrInit.CpuMemConfigId != c.WebData.CpuMemConfigId {
		cpuMem := CpuMemConf()
		oldVal := cpuMem[cntrInit.CpuMemConfigId]
		newVal := cpuMem[c.WebData.CpuMemConfigId]
		if err := c.insertOprRecord(tx, resLimitOpr, newVal.String(), oldVal.String(), cntrInit.Id); err != nil {
			tx.Rollback()
			return nil, err
		}
		updateMap["cpu_mem_config_id"] = c.WebData.CpuMemConfigId
	}
	if string(healthCheckJson) != cntrInit.HealthCheck {
		if err := c.insertOprRecord(tx, healthCheckOpr, string(healthCheckJson), cntrInit.HealthCheck, cntrInit.Id); err != nil {
			tx.Rollback()
			return nil, err
		}
		updateMap["health_check"] = string(healthCheckJson)
	}
	if string(envMapJson) != cntrInit.Environment {
		if err := c.insertOprRecord(tx, envOpr, string(envMapJson), cntrInit.Environment, cntrInit.Id); err != nil {
			tx.Rollback()
			return nil, err
		}
		updateMap["environment"] = string(envMapJson)
	}
	if string(logConfig) != cntrInit.LogConfig {
		if err := c.insertOprRecord(tx, logConfigOpr, string(logConfig), cntrInit.LogConfig, cntrInit.Id); err != nil {
			tx.Rollback()
			return nil, err
		}
		updateMap["log_config"] = string(logConfig)
	}
	if string(volume) != cntrInit.Volume {
		if err := c.insertOprRecord(tx, volumeOpr, string(volume), cntrInit.Volume, cntrInit.Id); err != nil {
			tx.Rollback()
			return nil, err
		}
		updateMap["volume"] = string(volume)
	}
	if string(scheduler) != cntrInit.Scheduler {
		if err := c.insertOprRecord(tx, schedulerOpr, string(scheduler), cntrInit.Scheduler, cntrInit.Id); err != nil {
			tx.Rollback()
			return nil, err
		}
		updateMap["scheduler"] = string(volume)
	}
	if err := tx.Model(&cntrInit).Updates(updateMap).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &cntrInit, nil
}

// 往操作记录表中插入新数据
func (c *CntrEdit) insertOprRecord(tx *gorm.DB, oprAction, newVal, oldVal string, CntrConfigId int) error {
	var log models.OprCntrLog
	log.Operator = c.Operator
	log.InsertTime = time.Now().Format(initial.DatetimeFormat)
	log.RelTable = "cntr_init_log"
	log.RelId = CntrConfigId
	log.OprAction = "更新" + oprAction
	log.OldVal = oldVal
	log.NewVal = newVal
	return tx.Create(&log).Error
}
