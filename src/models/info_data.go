package models

type UnitType struct {
	AppType string `gorm:"column:app_type" json:"app_type"`
}

type UnitSubType struct {
	AppSubType string `gorm:"column:app_sub_type" json:"app_sub_type"`
}

type DeployComp struct {
	CompEn string `gorm:"column:dumd_comp_en" json:"dumd_comp_en"`
}

type CaasTeam struct {
	TeamId   string `gorm:"column:team_id" json:"team_id"`
	TeamName string `gorm:"column:team_name" json:"team_name"`
}

type CaasCluster struct {
	ClusterId    string `gorm:"column:cluster_id" json:"cluster_id"`
	ClusterName  string `gorm:"column:cluster_name" json:"cluster_name"`
	ClusterUuid  string `gorm:"column:cluster_uuid" json:"cluster_uuid"`
}

type CaasStack struct {
	StackId   string `gorm:"column:stack_id" json:"stack_id"`
	StackName string `gorm:"column:stack_name" json:"stack_name"`
	StackUuid string `gorm:"column:stack_uuid" json:"stack_uuid"`
}

type CaasService struct {
	ServiceId   string `gorm:"column:service_id" json:"service_id"`
	ServiceName string `gorm:"column:service_name" json:"service_name"`
	ServiceUuid string `gorm:"column:service_uuid" json:"service_uuid"`
}
