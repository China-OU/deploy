package caas_route

import (
	"controllers/operation"
	"encoding/json"
	high_conc "high-conc"
	"initial"
	"library/caas"
	"library/common"
	"models"
	"strings"
)

// @Title EditRoute
// @Description 编辑路由服务
// @Param body body models.RouteWebData true "json body"
// @Success 200 {object} {}
// @Failure 403
// @router /route/edit [put]
func (c *CaaSRoute) EditRoute() {
	isAdmin := false
	if strings.Contains(c.Role, "admin") {
		isAdmin = true
	}
	var route models.RouteWebData
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &route); err != nil {
		c.SetJson(0, "", "参数绑定错误！"+err.Error())
		return
	}
	// 唯一性和权限校验
	err, oldRoute := models.RouteWebData{}.GetOneByID(route.Id)
	if err != nil {
		c.SetJson(0, "", "记录不存在")
		return
	}
	// using 2 map to keep the record, compare to web data
	oldRuleMap := make(map[int]models.RouteRule, 2)
	oldTargetMap := make(map[int]models.RouteTarget, 2)
	for _, r := range oldRoute.Rule {
		oldRuleMap[r.Id] = *r
		for _, t := range r.Target {
			oldTargetMap[t.Id] = *t
		}
	}
	ruleMap := make(map[string]bool, 2)
	requestRuleMap := make(map[string]bool, 2)
	tx := initial.DB.Begin()
	for _, r := range route.Rule {
		// 校验请求路径-端口是否重复 -- 唯一性校验
		ruleKey := common.GetString(r.RequestPort) + r.RequestPath
		if _, ok := requestRuleMap[ruleKey]; ok {
			tx.Rollback()
			c.SetJson(0, "", ruleKey+"请求端口路径冲突，请返回检查！")
			return
		} else {
			requestRuleMap[ruleKey] = true
		}
		// 规则名重复校验 -- 唯一性校验
		if _, ok := ruleMap[r.Name]; ok {
			tx.Rollback()
			c.SetJson(0, "", r.Name+"路由规则名称不能重复！")
			return
		} else {
			ruleMap[r.Name] = true
		}
		if r.RequestPort < 7000 || r.RequestPort > 65535 {
			tx.Rollback()
			c.SetJson(0, "", r.Name+"请求端口不合法：请填写[7000,65535]区间整数值！")
			return
		}
		if r.Id == 0 {
			if err := tx.Create(r).Error; err != nil {
				tx.Rollback()
				c.SetJson(0, "", "数据库错误："+err.Error())
				return
			}
		} else {
			if _, ok := oldRuleMap[r.Id]; ok {
				isEqual := models.RouteRule{}.IsEqual(oldRuleMap[r.Id], *r)
				if isEqual {
					// do nothing but can't continue, because has target one to compare
				} else {
					if isAdmin { // 规则只有管理员才能修改，目标服务开发人员可以修改
						if err := tx.Model(&models.RouteRule{}).Where("id = ?", r.Id).Updates(*r).Error; err != nil {
							tx.Rollback()
							c.SetJson(0, "", "数据库错误："+err.Error())
							return
						}
					} else {
						tx.Rollback()
						c.SetJson(0, "", "您不具备修改请求规则的权限，请联系deploy")
						return
					}
				}
			}
		}
		for _, s := range r.Target {
			s.RuleId = r.Id
			isNew := false // 是否新增
			if s.Id == 0 {
				isNew = true
			}
			// 校验目标服务是否存在
			err, isExist := models.CaasConfDetail{}.IsExist(route.CaasId, common.GetString(route.TeamId),
				route.ClusterUuid, route.StackName, s.TargetService)
			if err != nil {
				tx.Rollback()
				c.SetJson(0, "", "SQL查询错误："+err.Error())
				return
			}
			if !isExist {
				tx.Rollback()
				c.SetJson(0, "", s.TargetService+"容器服务不存在，请重选！")
				return
			}
			unit, err := models.UnitConfList{}.GetOneById(s.UnitId)
			if err != nil {
				tx.Rollback()
				c.SetJson(0, "", s.TargetService+"发布单元不存在，请先维护基础信息！")
				return
			}
			// 新增
			if isNew {
				hasAuth := false
				if isAdmin || strings.Contains(unit.Leader, c.UserId) {
					hasAuth = true
				}
				if hasAuth {
					if err := tx.Create(s).Error; err != nil {
						tx.Rollback()
						c.SetJson(0, "", "数据库错误："+err.Error())
						return
					}
				} else {
					tx.Rollback()
					c.SetJson(0, "", s.TargetService+"：只有发布单元负责人才有权限创建此规则！")
					return
				}

			} else { // 修改
				sIsEqual := models.RouteTarget{}.IsEqual(oldTargetMap[s.Id], *s)
				if sIsEqual {
					continue
				}
				hasAuth := false
				oldTarget, err := models.RouteTarget{}.GetOneById(s.Id)
				if err != nil {
					tx.Rollback()
					c.SetJson(0, "", s.TargetService+"发布单元不存在，请先维护基础信息！")
					return
				}
				oldUnit, err := models.UnitConfList{}.GetOneById(oldTarget.UnitId)
				if err != nil {
					tx.Rollback()
					c.SetJson(0, "", "旧的发布单元不存在，请先维护基础信息！")
					return
				}
				if isAdmin || (strings.Contains(oldUnit.Leader, c.UserId) && strings.Contains(unit.Leader, c.UserId)) {
					hasAuth = true
				}
				if hasAuth {
					if err := tx.Model(oldTargetMap[s.Id]).Updates(s).Error; err != nil {
						tx.Rollback()
						c.SetJson(0, "", "数据库错误："+err.Error())
						return
					}
				} else {
					tx.Rollback()
					c.SetJson(0, "", oldUnit.Unit+"、" + unit.Unit + "只有发布单元负责人才有权限覆盖更新此目标服务！")
					return
				}
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		c.SetJson(0, "", "tx error:保存失败")
		return
	}
	err, agentConf := operation.GetCaasConfig(route.DeployComp)
	caasOpr := caas.CaasOpr{
		AgentConf:   agentConf,
		TeamId:      common.GetString(route.TeamId),
		ClustUuid:   route.ClusterUuid,
		StackName:   route.StackName,
		ServiceName: route.ServiceName,
	}
	editWorker := &RouteEdit{
		Operator: c.UserId,
		Route:    route,
		Opr:      caasOpr,
	}
	high_conc.JobQueue <- editWorker
	c.SetJson(1, "", "更新中，请稍后关注更新结果！")
}

// @Title DeleteRouteRule
// @Description 删除路由规则
// @Param id query int true "rule id"
// @Success 200 {object} {}
// @Failure 403
// @router /route/rule/:id [delete]
func (c *CaaSRoute) DeleteRule() {
	id, _ := c.GetInt("id", 0)
	isAdmin := false
	if strings.Contains(c.Role, "admin") {
		isAdmin = true
	}
	if id == 0 {
		c.SetJson(0, "", "id不能为0")
		return
	}
	var rule models.RouteRule
	var broCnt int
	if err := initial.DB.Model(&models.RouteRule{}).Where("is_delete=0").First(&rule, id).Error; err != nil {
		c.SetJson(0, "", "记录不存在："+err.Error())
		return
	}
	if err := initial.DB.Model(&models.RouteRule{}).Where("is_delete=0 and route_id=?", rule.RouteId).Count(&broCnt).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if broCnt < 2 {
		c.SetJson(0, "", "至少保留一条规则！")
		return
	}
	var targetList []models.RouteTarget
	if err := initial.DB.Model(&models.RouteTarget{}).Where("is_delete=0 and rule_id=?", id).Find(&targetList).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	hasAuth := true
	for _, target := range targetList {
		if target.UnitId == 0 {
			c.SetJson(0, "", target.TargetService+"未关联容器服务发布单元，无法删除！")
			return
		}
		unit, err := models.UnitConfList{}.GetOneById(target.UnitId)
		if err != nil {
			c.SetJson(0, "", "发布单元不存在："+err.Error())
			return
		}
		if !strings.Contains(unit.Leader, c.UserId) {
			hasAuth = false
		}
	}
	if !hasAuth && !isAdmin {
		c.SetJson(0, "", "权限不足，请联系管理员！")
		return
	}
	tx := initial.DB.Begin()
	if err := tx.Model(&rule).Updates(map[string]interface{}{"is_delete": 1}).Error; err != nil {
		c.SetJson(0, "", "删除失败："+err.Error())
		return
	}
	if err := tx.Model(&models.RouteTarget{}).Where("rule_id=?", id).Updates(map[string]interface{}{"is_delete": 1}).Error; err != nil {
		c.SetJson(0, "", "删除失败："+err.Error())
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.SetJson(0, "", "事务提交失败")
		return
	}
	c.SetJson(1, "", "删除成功")
}

// @Title DeleteRouteTarget
// @Description 删除路由规则目标服务
// @Param id query int true "target id"
// @Success 200 {object} models.RouteWebData
// @Failure 403
// @router /route/target/:id [delete]
func (c *CaaSRoute) DeleteTarget() {
	id, _ := c.GetInt("id", 0)
	if id == 0 {
		c.SetJson(0, "", "id不能为0")
		return
	}
	isAdmin := false
	if strings.Contains(c.Role, "admin") {
		isAdmin = true
	}
	hasAuth := true
	var target models.RouteTarget
	var cnt int // 兄弟 target 计数器
	if err := initial.DB.Model(models.RouteTarget{}).Where("is_delete=0").First(&target, id).Error; err != nil {
		c.SetJson(0, "", "目标服务记录不存在："+err.Error())
		return
	}
	if err := initial.DB.Model(models.RouteTarget{}).Where("is_delete=0 and rule_id=?", target.RuleId).Count(&cnt).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if cnt == 1 {
		c.SetJson(0, "", "规则必须包含一条目标服务！")
		return
	}
	if target.UnitId == 0 {
		c.SetJson(0, "", target.TargetService+"未关联容器服务发布单元，无法删除！")
		return
	}
	unit, err := models.UnitConfList{}.GetOneById(target.UnitId)
	if err != nil {
		c.SetJson(0, "", "删除失败，发布单元不存在："+err.Error())
		return
	}
	if !isAdmin {
		if !strings.Contains(unit.Leader, c.UserId) {
			hasAuth = false
		}
	}
	if !hasAuth {
		c.SetJson(0, "", "权限不足，请联系管理员！")
		return
	}
	tx := initial.DB.Begin()
	if err := tx.Model(&models.RouteTarget{}).Updates(map[string]interface{}{"is_delete": 1}).Error; err != nil {
		c.SetJson(0, "", "删除失败："+err.Error())
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.SetJson(0, "", "事务提交失败")
		return
	}
	c.SetJson(1, "", "删除成功")
}

// @Title DeleteRoute
// @Description 删除路由规则目标服务
// @Param id query int true "route id"
// @Success 200 {object} {}
// @Failure 403
// @router /route/:id [delete]
func (c *CaaSRoute) DeleteRoute() {
	id, _ := c.GetInt("id", 0)
	if !strings.Contains(c.Role, "admin") {
		c.SetJson(0, "", "权限不足，请联系管理员！")
	}
	if id == 0 {
		c.SetJson(0, "", "id不能为0")
		return
	}
	var route models.ConfCaasRoute
	var ruleList []models.RouteRule
	updateMap := map[string]interface{}{"is_delete": 1}
	if err := initial.DB.Model(&models.ConfCaasRoute{}).Where("is_delete=0").First(&route, id).Error; err != nil {
		c.SetJson(0, "", "记录不存在！")
		return
	}
	if err := initial.DB.Model(&models.RouteRule{}).Where("is_delete=0 and route_id=?", id).Find(&ruleList).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	tx := initial.DB.Begin()
	// 级联删除
	if err := tx.Model(&route).Updates(updateMap).Error; err != nil {
		c.SetJson(0, "", "删除失败："+err.Error())
		return
	}
	if err := tx.Model(models.RouteRule{}).Where("is_delete=0 and route_id=?", id).Updates(updateMap).Error; err != nil {
		c.SetJson(0, "", "删除失败："+err.Error())
		return
	}
	for _, rule := range ruleList {
		if err := tx.Model(&models.RouteTarget{}).Where("rule_id=?", rule.Id).Updates(updateMap).Error; err != nil {
			c.SetJson(0, "", "删除失败："+err.Error())
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		c.SetJson(0, "", "事务提交失败")
		return
	}
	c.SetJson(1, "", "删除成功")
}
