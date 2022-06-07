package models

import (
	"fmt"
	"initial"
)

type CaasConf struct {
	Id            int       `gorm:"column:id" json:"id"`
	DeployComp    string    `gorm:"column:deploy_comp" json:"deploy_comp"`
	DeployNetwork string    `gorm:"column:deploy_network" json:"deploy_network"`
	CaasUrl       string    `gorm:"column:caas_url" json:"caas_url"`
	CaasPort      string    `gorm:"column:caas_port" json:"caas_port"`
	CaasToken     string    `gorm:"column:caas_token" json:"caas_token"`
	AgentIp	      string    `gorm:"column:agent_ip" json:"agent_ip"`
	AgentPort     string    `gorm:"column:agent_port" json:"agent_port"`
	DetailSyncTime     string    `gorm:"column:detail_sync_time" json:"detail_sync_time"`
	InsertTime    string    `gorm:"column:insert_time" json:"insert_time"`
	IsDelete      string    `gorm:"column:is_delete" json:"is_delete"`
	AgentSurv     string    `gorm:"column:agent_surv" json:"agent_surv"`
	CaasApiSurv        string    `gorm:"column:caas_api_surv" json:"caas_api_surv"`
	SurvCheckTime      string    `gorm:"column:surv_check_time" json:"surv_check_time"`
}

func (CaasConf) TableName() string {
	return "conf_caas"
}

func (CaasConf)ListByComp(comp string) (agentConf []CaasConf, err error) {
	cond := "is_delete=0"
	if comp != "" && comp != "all"  {
		cond += fmt.Sprintf(" and deploy_comp = '%s'", comp)
	}
	err = initial.DB.Model(&CaasConf{}).Where(cond).Find(&agentConf).Error
	return
}

type CaasConfDetail struct {
	Id            int       `gorm:"column:id" json:"id"`
	CaasId        int       `gorm:"column:caas_id" json:"caas_id"`
	TeamId        string    `gorm:"column:team_id" json:"team_id"`
	TeamName      string    `gorm:"column:team_name" json:"team_name"`
	TeamDesc      string    `gorm:"column:team_desc" json:"team_desc"`
	ClusterId     string    `gorm:"column:cluster_id" json:"cluster_id"`
	ClusterName   string    `gorm:"column:cluster_name" json:"cluster_name"`
	ClusterUuid   string    `gorm:"column:cluster_uuid" json:"cluster_uuid"`
	ClusterDesc   string    `gorm:"column:cluster_desc" json:"cluster_desc"`
	StackId       string    `gorm:"column:stack_id" json:"stack_id"`
	StackUuid     string    `gorm:"column:stack_uuid" json:"stack_uuid"`
	StackName     string    `gorm:"column:stack_name" json:"stack_name"`
	StackDesc     string    `gorm:"column:stack_desc" json:"stack_desc"`
	ServiceId     string    `gorm:"column:service_id" json:"service_id"`
	ServiceUuid   string    `gorm:"column:service_uuid" json:"service_uuid"`
	ServiceName   string    `gorm:"column:service_name" json:"service_name"`
	ServiceNum    int       `gorm:"column:service_num" json:"service_num"`
	InsertTime    string    `gorm:"column:insert_time" json:"insert_time"`
	IsDelete      int       `gorm:"column:is_delete" json:"is_delete"`
}

func (CaasConfDetail) TableName() string {
	return "conf_caas_detail"
}

func (CaasConfDetail)ListLB(caasId int) (routeList []CaasConfDetail,err error) {
	cond := fmt.Sprintf("is_delete = 0 and service_name like '%%-lb'")
	if caasId != 0 {
		cond += fmt.Sprintf(" and caas_id = %d", caasId )
	}
	err = initial.DB.Model(&CaasConf{}).Where(cond).Find(&routeList).Error
	return
}

func (CaasConfDetail)IsExist(caasId int, team, cluster, stack, service string) (err error, isExist bool) {
	cnt := 0
	cond := fmt.Sprintf("is_delete=0 and team_id = '%s' and " +
		"cluster_uuid='%s' and stack_name='%s' and service_name='%s'", team, cluster, stack, service)
	err = initial.DB.Table("conf_caas_detail").Select("id").Where(cond).Count(&cnt).Error
	if err != nil {
		return
	}
	if cnt > 0 {
		isExist = true
	}
	return
}