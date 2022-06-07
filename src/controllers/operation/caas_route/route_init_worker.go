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

type RouteInit struct {
	Operator string
	Route models.RouteWebData
	Opr caas.CaasOpr

}

func (c *RouteInit)Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("recover from panic when init route, error:", err)
		}
	}()
	initLog, err := c.InitLog()
	if err != nil {
		beego.Error("创建记录失败：" + err.Error())
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
	err = initLog.UpdateLog(1, costTime, "初始化成功","active")
	//初始化成功，添加到容器服务详情表--描述等信息，前端手动同步触发同步后即可补全
	caasId := c.Opr.AgentConf.Id
	team := caas.TeamDataDetail{Name: c.Route.TeamName, Id:c.Route.TeamId}
	clust := caas.ClustData{Uuid: c.Route.ClusterUuid, Name: c.Route.ClusterName}
	stack := caas.StackDataDetail{Name: c.Route.StackName,}
	serviceDetail := caas.ServiceDataDetail{Name: c.Route.ServiceName, State: "active"}
	err = caas_conf.InsertOrUpdateCaasDetail(caasId, team, clust, stack, serviceDetail)
	if err != nil {
		beego.Error(err.Error())
	}
}


func (c *RouteInit) InitRoute(record *models.ConfCaasRoute) (err error) {
	initData := models.CaasRouteData {ServiceName: c.Route.ServiceName}
	for _, rule := range c.Route.Rule {
		 r :=  models.RouteConfig{
		 	Name: rule.Name,
		 	Protocol: rule.Protocol,
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
		err = errors.New("caas新增路由接口调用失败！" + err.Error())
		return err
	}
	ec := 0
	for {
		ec += 1
		if ec > 20 {
			err = errors.New("初始化等待超时,请联系deploy")
			return
		}
		time.Sleep(30 * time.Second)
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
			if err := record.UpdateLog(1, ec*30,  "初始化中", detail.State); err != nil {
				beego.Error(err.Error())
				return err
			}
		}
	}
}

func (c *RouteInit) InitLog() (record *models.ConfCaasRoute,err error) {
	record = &models.ConfCaasRoute{
		ServiceName: c.Route.ServiceName,
		CaasId:      c.Route.CaasId,
		DeployComp:  c.Route.DeployComp,
		ClusterUuid: c.Route.ClusterUuid,
		ClusterName: c.Route.ClusterName,
		StackName:   c.Route.StackName,
		TeamId:      c.Route.TeamId,
		TeamName:    c.Route.TeamName,
		InsertTime:  time.Now().Format(initial.DatetimeFormat),
		Operator:    c.Operator,
		Status:      0,
	}
	tx := initial.DB.Begin()
	if err = tx.Create(record).Error; err != nil {
		return
	}
	// create rule
	for _, rule := range c.Route.Rule {
		rConfig := models.RouteConfig{
			Name:       rule.Name,
			Protocol:   rule.Protocol,
			RequestPath: rule.RequestPath,
			RequestHost: rule.RequestHost,
			RequestPort: rule.RequestPort,
		}
		r := models.CaasRouteRule {
			RouteConfig : rConfig,
			RouteId: record.Id,
		}
		if err = tx.Create(&r).Error; err != nil {
			return
		}
		for _, config := range rule.Target {
			target := models.Target{
				Weight:        config.Weight,
				TargetService: config.TargetService,
				TargetPath:    config.TargetPath,
				TargetPort:    config.TargetPort,
			}
			cRouteTarget := models.CaasRouteTarget{
				RuleId:   r.Id,
				UnitId:   config.UnitId,
				Target:   target,
			}
			if err = tx.Create(&cRouteTarget).Error; err != nil {
				return
			}
		}
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	return
}
