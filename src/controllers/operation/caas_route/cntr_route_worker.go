package caas_route

import (
	"github.com/astaxie/beego"
	"library/caas"
	"math/rand"
	"models"
	"sync"
	"time"
)

type CntrRouteSync struct {
	Comp string
}

func (c *CntrRouteSync) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("CaaSRouteSync Panic error:", err)
		}
	}()
	// 1. 获取 agent 配置并存储到 map（或者缓存）
	agentList, err := models.CaasConf{}.ListByComp(c.Comp)
	if err != nil {
		return
	}
	agentMap := make(map[int]models.CaasConf, 4)
	for _, agent := range agentList {
		agentMap[agent.Id] = agent
	}
	// 2. 获取每一个 lb 的详细配置
	caasId := 0
	if len(agentList) == 1 {
		caasId = agentList[0].Id
	}
	routeList, err := models.CaasConfDetail{}.ListLB(caasId)
	if err != nil {
		return
	}
	// 3. 同步
	wg := sync.WaitGroup{}
	wg.Add(len(routeList))
	for _, service := range routeList {
		go func(service models.CaasConfDetail) {
			defer func() {
				if err := recover(); err != nil {
					beego.Error("CaaSRouteSync Panic error:", err)
				}
			}()
			defer wg.Done()
			randInt := rand.Intn(9)
			time.Sleep(time.Duration(randInt+1) * time.Second)
			if _, ok := agentMap[service.CaasId]; !ok {
				beego.Error(service.ServiceName + "的agentConfId配置错误")
				return
			}
			caasOpr := caas.CaasOpr{
				AgentConf:   agentMap[service.CaasId],
				TeamId:      service.TeamId,
				ClustUuid:   service.ClusterUuid,
				StackName:   service.StackName,
				ServiceName: service.ServiceName,
			}
			err, caasRouteConfig := caasOpr.RetryGetRouteConfig(5)
			if err != nil {
				beego.Error(caasOpr.ServiceName + "路由配置获取失败！" + err.Error())
				return
			}
			beego.Info(caasRouteConfig.ServiceName, "路由配置获取成功")
			err = caasRouteConfig.InsertOrUpdate(service, agentMap[service.CaasId])
			if err != nil {
				beego.Error(err.Error())
			}
		}(service)

	}
	wg.Wait()
	beego.Info("同步完成")
}
