package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/jinzhu/gorm"
	"initial"
	"library/common"
	"time"
)

// 路由规则，一条路由规则可以将请求转发给多个目标服务
type RouteConfig struct {
	Name        string   `gorm:"column:name" json:"name"`                // 路由服务名
	Protocol    string   `gorm:"column:protocol" json:"protocol"`        // 协议
	RequestPath string   `gorm:"column:request_path" json:"requestPath"` // 路由请求路径
	RequestHost string   `gorm:"column:request_host" json:"requestHost"` // 路由主机
	RequestPort int      `gorm:"column:request_port" json:"requestPort"` // 路由端口
	Targets     []Target `sql:"-" json:"targets"`
}

type Target struct {
	Weight        int    `gorm:"column:weight" json:"weight"`                // 权重
	TargetService string `gorm:"column:target_service" json:"targetService"` // 目标服务，支持选择
	TargetPath    string `gorm:"column:target_path" json:"targetPath"`       // 目标路径
	TargetPort    int    `gorm:"column:target_port" json:"targetPort"`
}

// 路由规则
type CntrRouteRule struct {
	Id      int `gorm:"column:id" json:"id"`
	RouteId int `gorm:"column:route_id" json:"routeId"`
	UnitId  int `gorm:"column:unit_id" json:"unitId"`
	RouteConfig
}

// 堆栈路由
type ConfCaasRoute struct {
	Id          int       `gorm:"column:id" json:"id"`
	ServiceName string    `gorm:"column:service_name" json:"service_name"`
	CaasId      int       `gorm:"column:caas_id" json:"caas_id"`
	IsDelete    bool      `gorm:"column:is_delete" json:"is_delete"`
	DeployComp  string    `gorm:"column:deploy_comp" json:"deploy_comp"`
	ClusterUuid string    `gorm:"column:cluster_uuid" json:"cluster_uuid"`
	ClusterName string    `gorm:"column:cluster_name" json:"cluster_name"`
	StackName   string    `gorm:"column:stack_name" json:"stack_name"`
	TeamId      int       `gorm:"column:team_id" json:"team_id"`
	TeamName    string    `gorm:"column:team_name" json:"team_name"`
	InsertTime  string    `gorm:"column:insert_time" json:"insert_time"`
	Operator    string    `gorm:"column:operator" json:"operator"`
	Status      int       `json:"column:status" json:"status"`   // 下发状态 0: 更新中 1：成功 2:失败 -1:默认，原来就有的
	State       string    `gorm:"column:state" json:"state"`     // 服务状态
	Message     string    `gorm:"column:message" json:"message"` // 状态提示
	CostTime    int       `gorm:"column:cost_time" json:"cost_time"`
	UpdateTime  time.Time `sql:"-" json:"-"`
}

func (ConfCaasRoute) TableName() string {
	return "conf_caas_route"
}

func (ConfCaasRoute)IsExist(serviceName,stackName,clusterUuid,teamId,deployComp string) bool {
	cnt := 0
	cond := fmt.Sprintf("is_delete=0 and service_name = '%s'" +
		" and stack_name= '%s' and cluster_uuid='%s' and team_id=%s and deploy_comp='%s'" ,serviceName,
		stackName, clusterUuid, teamId, deployComp)
	if err := initial.DB.Table("conf_caas_route").Select("id").Where(cond).Count(&cnt).Error; err != nil {
		beego.Error(err)
		return true
	}
	if cnt > 0 {
		return true
	}
	return false
}

func (ConfCaasRoute) List(limit, page, teamId int, search, comp, cluster, stack string) (err error, total int, confList []*ConfCaasRoute) {
	cond := "is_delete = 0"
	if comp != "" && comp != "all" {
		cond += fmt.Sprintf(" and deploy_comp = '%s'", comp)
	}
	if teamId != 0 {
		cond += fmt.Sprintf(" and team_id = %d", teamId)
	}
	if cluster != "" {
		cond += fmt.Sprintf(" and cluster_uuid = '%s'", cluster)
	}
	if stack != "" {
		cond += fmt.Sprintf(" and stack_name = '%s'", stack)
	}
	if search != "" {
		cond += fmt.Sprintf(" and service_name like '%%%s%%'", search)
	}
	err = initial.DB.Table("conf_caas_route").Where(cond).Count(&total).
		Order("update_time desc").Offset(limit * (page - 1)).Limit(limit).Find(&confList).Error
	return
}

func (c *ConfCaasRoute) UpdateLog(status, cost int, msg, state string) error {
	tx := initial.DB.Begin()
	if err := tx.Model(&c).Updates(map[string]interface{}{
		"status":    status,
		"cost_time": cost,
		"message":   msg,
		"state":     state,
	}).Error; err != nil {
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

type CaasRouteRule struct {
	Id int `gorm:"column:id" json:"id"`
	RouteConfig
	RouteId  int `gorm:"column:route_id" json:"route_id"`
	IsDelete int `gorm:"column:is_delete" json:"is_delete"`
}

func (CaasRouteRule) TableName() string {
	return "caas_route_rule"
}

type CaasRouteTarget struct {
	Id       int  `gorm:"column:id" json:"id"`
	RuleId   int  `gorm:"column:rule_id" json:"rule_id"`
	IsDelete bool `gorm:"column:is_delete" json:"is_delete"`
	UnitId   int  `gorm:"column:unit_id" json:"unit_id"`
	Target
}

func (CaasRouteTarget) TableName() string {
	return "caas_route_target"
}

// caas 路由配置信息接口返回数据格式
type CaasRouteData struct {
	ServiceName string        `json:"serviceName"`
	RouteConfig []RouteConfig `json:"routeConfig"`
}

// 此种同步策略，无法同步下线的 lb
func (c *CaasRouteData) InsertOrUpdate(service CaasConfDetail, conf CaasConf) (err error) {
	var route ConfCaasRoute
	db := initial.DB
	cond := fmt.Sprintf("is_delete = 0 and " +
		"stack_name = '%s' and cluster_uuid = '%s' and team_id = %s and deploy_comp = '%s' and service_name = '%s'",
		 service.StackName, service.ClusterUuid, service.TeamId, conf.DeployComp, c.ServiceName,)
	err = db.Table("conf_caas_route").Where(cond).First(&route).Error
	tx := initial.DB.Begin()
	route.ServiceName = c.ServiceName
	route.CaasId = service.CaasId
	route.DeployComp = conf.DeployComp
	route.StackName = service.StackName
	route.ClusterUuid = service.ClusterUuid
	route.TeamName = service.TeamName
	route.TeamId = common.GetInt(service.TeamId)
	route.ClusterName = service.ClusterName
	route.InsertTime = time.Now().Format(initial.DatetimeFormat)
	if route.Status != 1 {
		route.Status = -1
	}
	route.State = "active"
	if gorm.IsRecordNotFoundError(err) { // 全量插入
		err = nil
		if err = tx.Create(&route).Error; err != nil {
			tx.Rollback()
			return
		}
	} else if err != nil {
		return
	}
	if err = tx.Save(&route).Error; err != nil {
		tx.Rollback()
		return
	}
	for _, routeConfig := range c.RouteConfig {
		var rule CaasRouteRule
		err = db.Table("caas_route_rule").Where("is_delete = 0 and route_id =? and name = ?", route.Id, routeConfig.Name).First(&rule).Error
		rule.RouteConfig = routeConfig
		rule.RouteId = route.Id
		if gorm.IsRecordNotFoundError(err) {
			err = tx.Create(&rule).Error
			if err != nil {
				tx.Rollback()
				return
			}
		} else if err != nil {
			return
		}
		// update
		err = tx.Save(&rule).Error
		if err != nil {
			tx.Rollback()
			return
		}
		for _, target := range routeConfig.Targets {
			var t CaasRouteTarget
			t.RuleId = rule.Id
			err = db.Table("caas_route_target").Where("is_delete=0 and target_service = ? and rule_id = ?", target.TargetService, rule.Id).First(&t).Error
			var cntrUnit UnitConfCntr
			err = db.Table("unit_conf_cntr").Where("service_name=? and caas_cluster = ? and caas_stack =? and caas_team = ?",
				target.TargetService, service.ClusterUuid, service.StackName, service.TeamId).First(&cntrUnit).Error
			if gorm.IsRecordNotFoundError(err) {
				err = nil
				t.UnitId = 0
			} else {
				t.UnitId = cntrUnit.UnitId
			}
			t.Target = target
			if gorm.IsRecordNotFoundError(err) {
				err = tx.Create(&t).Error
				if err != nil {
					tx.Rollback()
					return
				}
			} else if err != nil {
				return
			}
			err = tx.Save(&t).Error
			if err != nil {
				tx.Rollback()
				return
			}

		}
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return
	}
	return
}

type RouteWebData struct {
	ConfCaasRoute
	Rule []*RouteRule `sql:"-" json:"route_config"`
}

type RouteRule struct {
	Id          int            `gorm:"column:id" json:"id"`
	Name        string         `gorm:"column:name" json:"name"`                 // 路由服务名
	Protocol    string         `gorm:"column:protocol" json:"protocol"`         // 协议
	RequestPath string         `gorm:"column:request_path" json:"request_path"` // 路由请求路径
	RequestHost string         `gorm:"column:request_host" json:"request_host"` // 路由主机
	RequestPort int            `gorm:"column:request_port" json:"request_port"` // 路由端口
	RouteId     int            `gorm:"column:route_id" json:"route_id"`
	IsDelete    int            `gorm:"column:is_delete" json:"is_delete"`
	Target      []*RouteTarget `sql:"-" json:"targets"`
}
func (RouteRule)TableName() string {
	return "caas_route_rule"
}


func (RouteRule)GetOneById(id, routeId int) (routeRule RouteRule, err error) {
	err = initial.DB.Table("caas_route_rule").Where("is_delelte = 0 and route_id=?", routeId).First(&routeRule, id).Error
	return
}

func (RouteRule)IsEqual(a,b RouteRule) bool{
	return a.Name == b.Name && a.Protocol == b.Protocol && a.RouteId == b.RouteId && a.RequestPort == b.RequestPort && a.RequestPath == b.RequestPath && a.RequestHost == b.RequestHost
}

type RouteTarget struct {
	Id            int    `gorm:"column:id" json:"id"`
	RuleId        int    `gorm:"column:rule_id" json:"rule_id"`
	IsDelete      int   `gorm:"column:is_delete" json:"is_delete"`
	UnitId        int    `gorm:"column:unit_id" json:"unit_id"`
	Weight        int    `gorm:"column:weight" json:"weight"`                 // 权重
	TargetService string `gorm:"column:target_service" json:"target_service"` // 目标服务，支持选择
	TargetPath    string `gorm:"column:target_path" json:"target_path"`       // 目标路径
	TargetPort    int    `gorm:"column:target_port" json:"target_port"`
}

func (RouteTarget)TableName() string {
	return "caas_route_target"
}

func (RouteTarget)GetOneById(id int) (routeTarget RouteTarget, err error) {
	err = initial.DB.Table("caas_route_target").Where("is_delete = 0").First(&routeTarget, id).Error
	return
}

func (RouteTarget)IsEqual(a, b RouteTarget) bool {
	return a.UnitId == b.UnitId && a.RuleId == b.RuleId && a.Weight == b.Weight && a.TargetPort == b.TargetPort && a.TargetPath == b.TargetPath && a.TargetService == b.TargetService
}
// 复杂度：O(sum（config） * Log(n))
func (RouteWebData) GetOneByID(id int) (err error, route RouteWebData) {
	err = initial.DB.Table("conf_caas_route").Where("is_delete=0").First(&route, id).Error
	if err != nil {
		return
	}
	var ruleList []*RouteRule
	err = initial.DB.Table("caas_route_rule").Where("is_delete=0 and route_id=?", id).Find(&ruleList).Error
	if err != nil {
		return
	}

	for _, rule := range ruleList {
		var tList []*RouteTarget
		err = initial.DB.Table("caas_route_target").Where("is_delete=0 and rule_id=?", rule.Id).Find(&tList).Error
		if err != nil {
			return
		}
		rule.Target = tList
	}
	route.Rule = ruleList
	return
}
