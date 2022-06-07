package caas_route

import (
	"controllers"
	"controllers/operation"
	"encoding/json"
	high_conc "high-conc"
	"initial"
	"library/caas"
	"library/common"
	"library/datasession"
	"models"
	"strings"
	"time"
)

type CaaSRoute struct {
	controllers.BaseController
}


func (c *CaaSRoute) URLMapping() {
	c.Mapping("InitRoute", c.InitRoute)
	c.Mapping("ListRoute", c.ListRoute)
	c.Mapping("ListRouteLog", c.ListRouteLog)
	c.Mapping("FetchRouteService", c.FetchRouteService)
}
// @Title InitRoute
// @Description 新增路由服务
// @Param body body models.RouteWebData true "json body"
// @Success 200 {object} {}
// @Failure 403
// @router /route/int [post]
func (c *CaaSRoute) InitRoute() {
	isAdmin := false
	if strings.Contains(c.Role, "admin") {
		isAdmin = true
	}
	var route models.RouteWebData
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &route); err != nil {
		c.SetJson(0, "", "参数绑定错误！" + err.Error())
		return
	}
	// 校验 lb 服务是否已经创建
	if !strings.HasSuffix(route.ServiceName, "-lb") {
		c.SetJson(0, "", "路由服务名必须以-lb结尾！")
	}
	err, isExist := models.CaasConfDetail{}.IsExist(route.CaasId, common.GetString(route.TeamId),
		route.ClusterUuid, route.StackName, route.ServiceName)
	if err != nil {
		c.SetJson(0, "", "SQL查询错误：" + err.Error())
		return
	}
	if isExist {
		c.SetJson(0, "", "路由服务已存在，请勿重复创建！")
		return
	}
	// 唯一性和权限校验
	isExist = models.ConfCaasRoute{}.IsExist(route.ServiceName, route.StackName,
		route.ClusterUuid, common.GetString(route.TeamId), route.DeployComp)
	if isExist  {
		c.SetJson(0, "", "重复添加：请在原记录上操作！")
		return
	}

	ruleMap := make(map[string]bool,2)
	for _, r := range route.Rule {
		// 规则名重复校验
		if _, ok := ruleMap[r.Name]; ok {
			c.SetJson(0, "", r.Name + "路由规则名称不能重复！")
			return
		}
		// LB 端口范围校验
		// Todo: LB Rule Request PORT 校验，同一集群下，不能重复，否则会冲突
		if r.RequestPort < 7000 || r.RequestPort > 65535 {
			c.SetJson(0, "", r.Name + "请求端口不合法：请填写[7000,65535]区间整数值！")
			return
		}
		for _, s := range r.Target {
			// 校验目标服务是否存在
			err, isExist := models.CaasConfDetail{}.IsExist(route.CaasId, common.GetString(route.TeamId),
				route.ClusterUuid, route.StackName, s.TargetService)
			if err != nil {
				c.SetJson(0, "", "SQL查询错误：" + err.Error())
				return
			}
			if !isExist {
				c.SetJson(0, "", s.TargetService + "容器服务不存在，请重选！")
				return
			}
			// 检验操作人是否有权限创建
			if !isAdmin {
				unit, err := models.UnitConfList{}.GetOneById(s.UnitId)
				if err != nil {
					c.SetJson(0, "", s.TargetService + "发布单元不存在，请先维护基础信息！")
					return
				}
				if !strings.Contains(unit.Leader, c.UserId) {
					c.SetJson(0, "", "您不是该发布单元负责人，没有权限创建！")
					return
				}
			}
		}
	}
	err, agentConf := operation.GetCaasConfig(route.DeployComp)
	caasOpr := caas.CaasOpr{
		AgentConf:   agentConf,
		TeamId:      common.GetString(route.TeamId),
		ClustUuid:   route.ClusterUuid,
		StackName:   route.StackName,
		ServiceName: route.ServiceName,
	}
	initWorker := &RouteInit{
		Operator: c.UserId,
		Route:    route,
		Opr: caasOpr,
	}
	high_conc.JobQueue <- initWorker
	c.SetJson(1, "", "创建中，请稍后关注创建结果！")
}


// @Title GetRoute
// @Description 获取路由详情
// @Param id query int true "route id"
// @Success 200 {object} models.RouteWebData
// @Failure 403
// @router /route/:id/detail [get]
func (c *CaaSRoute) GetRoute() {
	id, _ := c.GetInt("id", 0)
	if id == 0 {
		c.SetJson(0, "", "id不能为0")
		return
	}
	err, route := models.RouteWebData{}.GetOneByID(id)
	if err != nil {
		c.SetJson(0, "", "获取数据失败："+ err.Error())
		return
	}
	c.SetJson(1, route, "获取成功")
}

// @Title ListRoute
// @Description 查询路由规则列表
// @Param limit query int true "记录数"
// @Param page  query int true "页码"
// @Param search query string false "搜索关键字"
// @Param comp  query string false "租户"
// @Param stack_name query string false "堆栈名"
// @Param cluster_uuid query string false "集群uuid"
// @Success 200 {object} {}
// @Failure 403
// @router /route/list [get]
func (c *CaaSRoute) ListRoute() {
	limit, _ := c.GetInt("limit", 40)
	page, _ := c.GetInt("page", 1)
	search := c.GetString("search", "")
	comp := c.GetString("comp", "")
	clusterUuid := c.GetString("cluster_uuid", "")
	stackName := c.GetString("stack_name", "")
	teamId, _ := c.GetInt("team_id", 0)
	err, total, routeList := models.ConfCaasRoute{}.List(limit, page, teamId, search, comp, clusterUuid, stackName)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, map[string]interface{}{"total":total, "data":routeList}, "")
}

// @Title ListRouteInitLog
// @Description 获取创建路由日志列表
// @Param limit query int true "记录数"
// @Param page  query int true "页码"
// @Param search query string true "搜索关键字"
// @Success 200 {object} {}
// @Failure 403
// @router /route/log/list [get]
func (c *CaaSRoute) ListRouteLog() {

}

// @Title FetchRouteService
// @Description 同步路由服务详情
// @Param comp query string true "租户英文名大写或all（全部租户）"
// @Success 200 {object} {}
// @Failure 403
// @router /route/sync-caas/:comp [get]
func (c *CaaSRoute) FetchRouteService() {
	isAdmin := false
	if strings.Contains(c.Role, "admin") {
		isAdmin = true
	}
	if !isAdmin {
		c.SetJson(0, "", "权限不足")
		return
	}
	comp := c.GetString("comp")
	if comp == "all" {
		lastTime, flag := datasession.CaasRouteSyncTime()
		if time.Now().Add(- 30 * time.Minute).Format(initial.DatetimeFormat) < common.GetString(lastTime) && flag == 1 {
			c.SetJson(0, "", "Caas路由列表30分钟内只能同步一次，上次同步时间：" + common.GetString(lastTime))
			return
		}
	} else {
		lastTime, flag := datasession.CaasSingleSyncTime()
		if time.Now().Add(- 5 * time.Minute).Format(initial.DatetimeFormat) < common.GetString(lastTime) && flag == 1 {
			c.SetJson(0, "", "Caas单租户路由列表5分钟内只能同步一次，上次同步时间：" + common.GetString(lastTime))
			return
		}
	}
	routeSync := &CntrRouteSync{Comp: comp}
	high_conc.JobQueue <- routeSync
	c.SetJson(1, "", "路由服务同步中！")
}