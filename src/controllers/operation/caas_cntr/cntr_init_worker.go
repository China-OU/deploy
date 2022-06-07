package caas_cntr

import (
	"controllers/caas_conf"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"initial"
	"library/caas"
	"library/common"
	"models"
	"strings"
	"time"
)

type CntrInit struct {
	Operator string // 操作人员
	WebData  caas.InitServiceWebData
	Opr      caas.CaasOpr // 基础配置

}

func (c *CntrInit) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("Cntr Init Panic error:", err)
		}
	}()
	initLog, err := c.InitLog()
	if err != nil {
		beego.Error(err.Error())
		return
	}

	startAt := time.Now()
	err = c.InitService(initLog)
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
		err = initLog.UpdateLog(2, "初始化失败："+err.Error(), "", costTime)
		if err != nil {
			beego.Error(err.Error())
		}
		return
	}
	err = initLog.UpdateLog(1, "初始化成功", "active", costTime)
	//初始化成功，添加到容器服务详情表--描述等信息，前端手动同步触发同步后即可补全
	caasId := c.Opr.AgentConf.Id
	team := caas.TeamDataDetail{
		Name: c.WebData.TeamName,
		Id:   common.StrToInt(c.WebData.TeamId)}
	clust := caas.ClustData{Uuid: c.WebData.ClusterUuid,
		Name: c.WebData.ClusterName}
	stack := caas.StackDataDetail{
		Name: c.WebData.StackName,
	}
	serviceDetail := caas.ServiceDataDetail{Name: c.WebData.ServiceName,
		State: "active",
		Image: c.WebData.Image,
	}
	err = caas_conf.InsertOrUpdateCaasDetail(caasId, team, clust, stack, serviceDetail)
	if err != nil {
		beego.Error(err.Error())
	}
}

func (c *CntrInit) InitService(initLog *models.OprCntrInit) error {
	initData := caas.InitServiceData{
		AgentConf:          c.Opr.AgentConf,
		InitServiceWebData: c.WebData,
	}
	if err := initData.RetryInitService(5); err != nil {
		return errors.New("caas新增服务接口调用失败！" + err.Error())
	}
	ec := 0
	for {
		ec += 1
		if ec > 50 {
			return errors.New("初始化等待超时,请联系deploy")
		}
		time.Sleep(30 * time.Second)
		retry := 5
		err, detail := c.Opr.RetryGetService(retry)
		if err != nil {
			beego.Info(err.Error())
			return errors.New("caas获取服务详情接口调用失败！" + err.Error())
		}
		if detail.State == "active" {
			return nil
		} else if len(detail.State) > 0 {
			if err := initLog.UpdateLog(1, "初始化中", detail.State, ec*30); err != nil {
				beego.Error(err.Error())
			}
		}
	}
}

func (c *CntrInit) InitLog() (*models.OprCntrInit, error) {
	now := time.Now()
	healthCheckJson, _ := json.Marshal(c.WebData.HealthCheck)
	envMapJson, _ := json.Marshal(c.WebData.Environment)
	logConfig, _ := json.Marshal(c.WebData.LogConfig)
	volume, _ := json.Marshal(c.WebData.Volume)
	scheduler, _ := json.Marshal(c.WebData.Scheduler)
	cntrInit := &models.OprCntrInit{
		Agent:          c.Opr.AgentConf.AgentIp + ":" + c.Opr.AgentConf.AgentPort,
		ClusterUuid:    c.WebData.ClusterUuid,
		TeamId:         c.WebData.TeamId,
		StackName:      c.WebData.StackName,
		ServiceName:    c.WebData.ServiceName,
		UnitId:         c.WebData.UnitId,
		Image:          c.WebData.Image,
		InstanceNum:    c.WebData.InstanceNum,
		Cpu:            c.WebData.Cpu * 1000,
		MemLimit:       c.WebData.MemLimit,
		ClusterName:    c.WebData.ClusterName,
		TeamName:       c.WebData.TeamName,
		Comp:           c.WebData.Comp,
		OnlineDate:     now.Format(initial.DateFormat),
		Operator:       c.Operator,
		InsertTime:     now.Format(initial.DatetimeFormat),
		Result:         0,
		Message:        "初始化中",
		HealthCheck:    string(healthCheckJson),
		Environment:    string(envMapJson),
		LogConfig:      string(logConfig),
		CpuMemConfigId: c.WebData.CpuMemConfigId,
		AppType:        c.WebData.AppType,
		IsEdit:         false,
		State:          "",
		Volume:         string(volume),
		Scheduler:      string(scheduler),
	}
	tx := initial.DB.Begin()
	if err := tx.Create(&cntrInit).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return cntrInit, nil
}
