package models

import (
	"errors"
	"fmt"
	"initial"
	"library/common"
	"regexp"
)

type OprCntrInit struct {
	Id             int    `gorm:"column:id" json:"id"`
	Agent          string `gorm:"column:agent" json:"agent"`
	ClusterUuid    string `gorm:"column:cluster_uuid" json:"cluster_uuid"`
	TeamId         string `gorm:"column:team_id" json:"team_id"`
	StackName      string `gorm:"column:stack_name" json:"stack_name"`
	ServiceName    string `gorm:"column:service_name" json:"service_name"`
	UnitId         int    `gorm:"column:unit_id" json:"unit_id"`
	Image          string `gorm:"column:image" json:"image"`
	InstanceNum    int    `gorm:"column:instance_num" json:"instance_num"`
	Cpu            int    `gorm:"column:cpu" json:"cpu"`
	MemLimit       int    `gorm:"column:mem_limit" json:"mem_limit"` // mb
	Result         int    `gorm:"column:result" json:"result"`
	Message        string `gorm:"column:message" json:"message"`
	OnlineDate     string `gorm:"column:online_date" json:"online_date"`
	CostTime       int    `gorm:"column:cost_time" json:"cost_time"`       // s
	LogConfig      string `gorm:"column:log_config" json:"log_config"`     // 日志配置 保留字段
	HealthCheck    string `gorm:"column:health_check" json:"health_check"` // 健康检查 保留字段
	Environment    string `gorm:"column:environment" json:"environment"`   // 环境变量 保留字段
	Scheduler      string `gorm:"column:scheduler" json:"scheduler"`       // 调度策略
	Volume         string `gorm:"column:volume" json:"volume"`
	Operator       string `gorm:"column:operator" json:"operator"`
	InsertTime     string `gorm:"column:insert_time" json:"insert_time"`
	ClusterName    string `gorm:"column:cluster_name" json:"cluster_name"` // 集群名称
	TeamName       string `gorm:"column:team_name" json:"team_name"`       // 团队名称
	Comp           string `gorm:"column:comp" json:"comp"`                 // 租户
	AppType        string `gorm:"column:app_type" json:"app_type"`
	CpuMemConfigId int    `gorm:"column:cpu_mem_config_id"  json:"cpu_mem_config_id"` // cpu内存配置ID
	IsEdit         bool   `gorm:"column:is_edit" json:"is_edit"`
	State          string `gorm:"column:state" json:"state"` // 服务状态
	IsDelete       bool   `gorm:"column:is_delete" json:"is_delete"`
}

func (OprCntrInit) TableName() string {
	return "opr_cntr_init"
}

func (self *OprCntrInit) UpdateLog(result int, msg, state string, cost int) error {
	tx := initial.DB.Begin()
	updateMap := map[string]interface{}{
		"result":    result,
		"message":   msg,
		"cost_time": cost,
		"state":     state,
	}
	if err := tx.Model(&self).Updates(updateMap).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

type OprCntrInitDetail struct {
	OprCntrInit
	Name   string `gorm:"column:name" json:"name"`
	Unit   string `gorm:"column:unit" json:"unit"`
	Leader string `gorm:"column:leader" json:"leader"`
}

func (self OprCntrInit) List(isAdmin bool, leader string, search string, page, limit int) (total int, initList []OprCntrInitDetail, err error) {
	//whereStr := fmt.Sprintf("is_edit = true ")
	whereStr := fmt.Sprintf("a.unit_id = b.id and a.is_delete=0")

	if !isAdmin {
		whereStr += fmt.Sprintf(" and b.leader='%s'", leader)
	}
	if search != "" {
		whereStr += fmt.Sprintf(" and concat(b.unit,b.name) like '%%%s%%'", search)
	}

	db := initial.DB
	err = db.Table("opr_cntr_init as a, unit_conf_list as b").Select(`a.id, a.comp, a.cluster_name, a.stack_name, 
a.service_name, a.instance_num, a.cost_time, a.result, a.message, a.insert_time,a.state, 
a.operator, a.image, a.cpu_mem_config_id, b.unit, b.name, b.leader`).Where(whereStr).
		Count(&total).Order("a.insert_time desc").Scan(&initList).
		Offset((page - 1) * limit).Limit(limit).Find(&initList).Error
	return total, initList, err
}

func (self OprCntrInit) GetOneById(isAdmin bool, leader string, id int) (item OprCntrInitDetail, err error) {
	whereStr := fmt.Sprintf("a.id=%d and a.unit_id = b.id and a.is_delete=0", id)
	if !isAdmin {
		whereStr += fmt.Sprintf(" and b.leader='%s'", leader)
	}
	var itemList []OprCntrInitDetail
	err = initial.DB.Table("opr_cntr_init as a, unit_conf_list as b").Select(
		`a.id, a.comp, a.agent,  a.cluster_name, a.stack_name, a.service_name, a.instance_num, a.health_check,
a.environment, a.log_config, a.volume, a.scheduler, a.app_type, a.image, a.cpu_mem_config_id, a.unit_id, a.cluster_uuid, a.team_id, a.team_name, b.unit, b.name, b.leader`).
		Where(whereStr).Scan(&itemList).Error
	if len(itemList) == 1 {
		item = itemList[0]
	}
	return
}

type HealthCheck struct {
	Enable              string `json:"enable"`              // 是否启用
	Path                string `json:"path"`                // 健康检查路径，HTTP时需要填写
	Protocol            string `json:"protocol"`            // 协议
	ProtocolVersion     string `json:"protocolVersion"`     // 协议版本，HTTP时只能为 1.0 / 1.1
	Port                int    `json:"port"`                // 端口号
	Interval            int    `json:"interval"`            // 检查间隔
	InitializingTimeout int    `json:"initializingTimeout"` // 初始化超时
	HealthyThreshold    int    `json:"healthyThreshold"`    // 健康阈值
	UnhealthyThreshold  int    `json:"unhealthyThreshold"`  // 不健康阈值
	ResponseTimeout     int    `json:"responseTimeout"`     // 检查超时
	Strategy            string `json:"strategy"`            // 不健康时策略，recreate 或置空
}

type LogConfig struct {
	Path     string `json:"path"`     // 日志路径
	Encoding string `json:"encoding"` // 编码
	Filter   string `json:"filter"`   // 过滤条件
	Rule     string `json:"rule"`     // 多行规则
	Negate   string `json:"negate"`   // 是否匹配
	Match    string `json:"match"`    // 匹配行日志位置，before 或 after
	Topic    string `json:"topic"`    // 主题
}

type Volume struct {
	ContainerPath string `json:"containerPath"` // 容器路径
	HostPath      string `json:"hostPath"`      // 主机路径
	Driver        string `json:"driver"`        // 卷类型，2020年2月11日前只支持 local 本地卷
}

func (v *Volume) OK() error {
	if v.Driver != "local" {
		return errors.New("存储类型暂只本地卷（local）")
	}
	if len(v.ContainerPath) == 0 || len(v.ContainerPath) == 0 {
		return errors.New("卷目录不能为空")
	}
	return nil
}

type Environment map[string]string // 环境变量键值对

// 环境变量规范校验
func (self Environment) OK() bool {
	for k, v := range self {
		kRegexp := `^[a-zA-Z0-9_-]*$`
		if len(v) < 2 || len(v) > 256 {
			return false
		}
		kMatched, _ := regexp.Match(kRegexp, []byte(k))
		if !kMatched {
			return false
		}
	}
	return true
}

func (self *HealthCheck) OK() error {
	self.Enable = "true"
	self.Protocol = "TCP"
	self.Strategy = "recreate"
	self.Path = ""
	self.ProtocolVersion = ""
	if self.Port == 0 {
		return errors.New("端口不能为0")
	}
	return nil
}

type Scaling struct {
	DefaultInstances int `json:"defaultInstances"` // 默认实例数
}

type Scheduler struct {
	Policy   string `json:"policy"` // must:必须 mustNot:必须没有 betterNot:最好没有
	Selector string `json:"selector"` // node:主机节点标签, pod: 容器组标签
	Key      string `json:"key"` // 0-9 a-z A-Z
	Value    string `json:"value"` // 0-9 a-z A-Z
}

func (self *Scheduler) OK() error {
	if self.Policy != "mustNot" && self.Policy != "must" && self.Policy != "betterNot" {
		return errors.New("调度规则策略不合法！")
	}
	if self.Selector != "node" && self.Selector != "pod" {
		return errors.New("标签类型不合法，仅支持node和pod！")
	}
	reStr := `^[a-zA-Z0-9]+$`
	matched, err :=  common.RegexpMatched(reStr, self.Key)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("scheduler value invalid")
	}
	matched, err =  common.RegexpMatched(reStr, self.Value)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("scheduler value invalid")
	}
	return nil
}

type OprCntrLog struct {
	Id         int    `gorm:"column:id" json:"id"`
	OprAction  string `gorm:"column:opr_action" json:"oprAction"`
	OldVal     string `gorm:"column:old_val" json:"oldVal"`
	NewVal     string `gorm:"column:new_val" json:"newVal"`
	RelTable   string `gorm:"column:rel_table" json:"relTable"`
	RelId      int    `gorm:"column:rel_id" json:"relId"`
	Operator   string `gorm:"column:operator" json:"operator"`
	InsertTime string `gorm:"column:insert_time" json:"insertTime"`
}

type OprCntrLogDetail struct {
	OprCntrLog
	ServiceName string `gorm:"column:service_name" json:"serviceName"`
}

func (OprCntrLog) TableName() string {
	return "opr_cntr_log"
}

func (OprCntrLog) Find(search string, page, limit int) (total int, logList []OprCntrLogDetail, err error) {
	whereStr := "a.rel_id=b.id"
	if len(search) != 0 {
		whereStr += fmt.Sprintf(" and b.service_name like '%%%s%%'", search)
	}
	err = initial.DB.Table("opr_cntr_log as a, opr_cntr_init as b").Select(`
a.id, a.opr_action, a.old_val, a.new_val, a.operator, a.insert_time, a.rel_id, b.service_name`).Where(whereStr).
		Count(&total).Order("a.insert_time desc").Offset((page - 1) * limit).Limit(limit).Scan(&logList).Error
	return
}
