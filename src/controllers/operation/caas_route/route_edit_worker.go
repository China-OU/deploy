package caas_route

import (
	"controllers/caas_conf"
	"errors"
	"github.com/astaxie/beego"
	"initial"
	"library/caas"
	"library/common"
	"models"
	"strings"
	"time"
)

type RouteEdit struct {
	Operator string
	Route    models.RouteWebData
	Opr      caas.CaasOpr
}

func (c *RouteEdit) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("recover from panic when edit route, error:", err)
		}
	}()
	initLog, err := c.InitLog()
	if err != nil {
		beego.Error("更新记录失败：" + err.Error())
		return
	}
	startAt := time.Now()
	err = c.InitRoute(initLog)
	cost := time.Now().Sub(startAt).Seconds()
	costTime := common.GetInt(cost)
	if err != nil {
		// 超时
		if strings.Contains(err.Error(), "超时") {
			if err := initLog.UpdateLog(3, costTime, err.Error(), ""); err != nil {
				beego.Error(err.Error())
			}
		}
		// 失败
		err = initLog.UpdateLog(2, costTime, err.Error(), "")
		if err != nil {
			beego.Error(err.Error())
		}
		return
	}
	err = initLog.UpdateLog(1, costTime, "更新成功", "active")
	//初始化成功，添加到容器服务详情表--描述等信息，前端手动同步触发同步后即可补全
	caasId := c.Opr.AgentConf.Id
	team := caas.TeamDataDetail{Name: c.Route.TeamName, Id: c.Route.TeamId}
	clust := caas.ClustData{Uuid: c.Route.ClusterUuid, Name: c.Route.ClusterName}
	stack := caas.StackDataDetail{Name: c.Route.StackName}
	serviceDetail := caas.ServiceDataDetail{Name: c.Route.ServiceName, State: "active"}
	err = caas_conf.InsertOrUpdateCaasDetail(caasId, team, clust, stack, serviceDetail)
	if err != nil {
		beego.Error(err.Error())
	}
}

func (c *RouteEdit) InitRoute(record *models.ConfCaasRoute) (err error) {
	initData := models.CaasRouteData{ServiceName: c.Route.ServiceName}
	for _, rule := range c.Route.Rule {
		r := models.RouteConfig{
			Name:        rule.Name,
			Protocol:    rule.Protocol,
			RequestPort: rule.RequestPort,
			RequestPath: rule.RequestPath,
			RequestHost: rule.RequestHost,
		}
		for _, target := range rule.Target {
			var t models.Target
			t.TargetService = target.TargetService
			t.TargetPath = target.TargetPath
			t.TargetPort = target.TargetPort
			t.Weight = target.Weight
			r.Targets = append(r.Targets, t)
		}
		initData.RouteConfig = append(initData.RouteConfig, r)
	}
	if err := c.Opr.RetryEditRoute(initData, 5); err != nil {
		err = errors.New("caas更新路由接口调用失败！" + err.Error())
		return err
	}
	ec := 0
	for {
		ec += 1
		if ec > 20 {
			err = errors.New("更新等待超时,请联系deploy")
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
		retry := 5
		err, detail := c.Opr.RetryGetService(retry)
		if err != nil {
			beego.Info(err.Error())
			err = errors.New("caas获取服务详情接口调用失败！" + err.Error())
			return err
		}
		if detail.State == "active" {
			return nil
		} else if len(detail.State) > 0 {
			if err := record.UpdateLog(1, ec*30, "更新中", detail.State); err != nil {
				beego.Error(err.Error())
				return err
			}
		}
	}
}

func (c *RouteEdit) InitLog() (record *models.ConfCaasRoute, err error) {
	updateMap := map[string]interface{}{
		"state":    "",
		"operator": c.Operator,
		"status":   0,
		"message":  "更新中",
	}
	tx := initial.DB.Begin()
	if err := tx.Model(&c.Route.ConfCaasRoute).Updates(updateMap).Error; err != nil {
		return record, err
	}
	if err = tx.Commit().Error; err != nil {
		return record, err
	}
	record = &c.Route.ConfCaasRoute
	return record, err
}
