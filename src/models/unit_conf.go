package models

import (
	"fmt"
	"initial"
)

// 全部发布单元列表
type UnitConfList struct {
	Id             int       `gorm:"column:id" json:"id"`
	Unit           string    `gorm:"column:unit" json:"unit"`
	Name           string    `gorm:"column:name" json:"name"`
	Note           string    `gorm:"column:note" json:"note"`
	Info           string    `gorm:"column:info" json:"info"`
	GitUnit        string    `gorm:"column:git_unit" json:"git_unit"`
	GitId          string    `gorm:"column:git_id" json:"git_id"`
	Leader         string    `gorm:"column:leader" json:"leader"`
	Developer      string    `gorm:"column:developer" json:"developer"`
	Test           string    `gorm:"column:test" json:"test"`
	ReleaseTime    string    `gorm:"column:release_time" json:"release_time"`
	AppSubType     string    `gorm:"column:app_sub_type" json:"app_sub_type"`
	AppType        string    `gorm:"column:app_type" json:"app_type"`
	DumdSubsysId   int       `gorm:"column:dumd_subsys_id" json:"dumd_subsys_id"`
	DumdCompEn     string    `gorm:"column:dumd_comp_en" json:"dumd_comp_en"`
	DumdSubSysname string    `gorm:"column:dumd_sub_sysname" json:"dumd_sub_sysname"`
	DumdSubSysnameCn string  `gorm:"column:dumd_sub_sysname_cn" json:"dumd_sub_sysname_cn"`
	IsOffline      int       `gorm:"column:is_offline" json:"is_offline"`
	InsertTime     string    `gorm:"column:insert_time" json:"insert_time"`
	Operator       string    `gorm:"column:operator" json:"operator"`
}

func (UnitConfList) TableName() string {
	return "unit_conf_list"
}

func (self UnitConfList)Find(search, uid, appType string, page,limit int) (total int, unitList []*UnitConfList,err error){
	if page < 0 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	whereStr := fmt.Sprintf("1=1")
	if search != "" {
		whereStr += fmt.Sprintf(" AND unit like '%%%s%%'", search)
	}
	if uid != "" {
		whereStr += fmt.Sprintf(" AND leader like '%%%s%%'", uid)
	}
	err = initial.DB.Model(&UnitConfList{}).Where(whereStr).Count(&total).Order("unit asc,id").Offset((page-1)*limit).Limit(limit).Find(&unitList).Error
	return total, unitList, err
}

func (self UnitConfList)GetOneById(id int) (unit UnitConfList,err error) {
	err = initial.DB.First(&unit, id).Error
	return
}

// db列表
type UnitConfDb struct {
	Id            int       `gorm:"column:id" json:"id"`
	UnitId        int       `gorm:"column:unit_id" json:"unit_id"`
	Type          string    `gorm:"column:type" json:"type"`
	Username      string    `gorm:"column:username" json:"username"`
	EncryPwd      string    `gorm:"column:encry_pwd" json:"encry_pwd"`
	Host          string    `gorm:"column:host" json:"host"`
	Port          int       `gorm:"column:port" json:"port"`
	Dbname        string    `gorm:"column:dbname" json:"dbname"`
	GitId         string    `gorm:"column:git_id" json:"git_id"`
	GitUnit       string    `gorm:"column:git_unit" json:"git_unit"`
	GitUrl        string    `gorm:"column:git_url" json:"git_url"`
	DeployComp    string    `gorm:"column:deploy_comp" json:"deploy_comp"`
	ConnResult    int       `gorm:"column:conn_result" json:"conn_result"`
	ConnCtime     string    `gorm:"column:conn_ctime" json:"conn_ctime"`
	PwdCtime      string    `gorm:"column:pwd_ctime" json:"pwd_ctime"`
	IsDelete      int       `gorm:"column:is_delete" json:"is_delete"`
	InsertTime    string    `gorm:"column:insert_time" json:"insert_time"`
	Operator      string    `gorm:"column:operator" json:"operator"`
}

func (UnitConfDb) TableName() string {
	return "unit_conf_db"
}

// db密码历史记录列表
type RcDbPwd struct {
	Id            int       `gorm:"column:id" json:"id"`
	DbConfId      int       `gorm:"column:db_conf_id" json:"db_conf_id"`
	Username      string    `gorm:"column:username" json:"username"`
	EncryPwd      string    `gorm:"column:encry_pwd" json:"encry_pwd"`
	Host          string    `gorm:"column:host" json:"host"`
	Port          int       `gorm:"column:port" json:"port"`
	Dbname        string    `gorm:"column:dbname" json:"dbname"`
	InsertTime    string    `gorm:"column:insert_time" json:"insert_time"`
	Operator    string    `gorm:"column:operator" json:"operator"`
}

func (RcDbPwd) TableName() string {
	return "rc_db_pwd"
}

// 标准容器应用列表，网络租户+网络区域+租户+集群+堆栈+服务，才能完全确定一个服务
type UnitConfCntr struct {
	Id            int       `gorm:"column:id" json:"id"`
	UnitId        int       `gorm:"column:unit_id" json:"unit_id"`
	ServiceName   string    `gorm:"column:service_name" json:"service_name"`
	AppType       string    `gorm:"column:app_type" json:"app_type"`
	AppSubType    string    `gorm:"column:app_sub_type" json:"app_sub_type"`
	GitId         string    `gorm:"column:git_id" json:"git_id"`
	GitUnit       string    `gorm:"column:git_unit" json:"git_unit"`
	GitUrl        string    `gorm:"column:git_url" json:"git_url"`
	CaasTeam      string    `gorm:"column:caas_team" json:"caas_team"`
	CaasCluster   string    `gorm:"column:caas_cluster" json:"caas_cluster"`
	CaasStack     string    `gorm:"column:caas_stack" json:"caas_stack"`
	DeployComp    string    `gorm:"column:deploy_comp" json:"deploy_comp"`
	DeployNetwork string    `gorm:"column:deploy_network" json:"deploy_network"`
	IsDelete      int       `gorm:"column:is_delete" json:"is_delete"`
	InsertTime    string    `gorm:"column:insert_time" json:"insert_time"`
	BeforeXml     string    `gorm:"column:before_xml" json:"before_xml"`
	BuildXml      string    `gorm:"column:build_xml" json:"build_xml"`
	AfterXml      string    `gorm:"column:after_xml" json:"after_xml"`
	JenkinsNode   string    `gorm:"column:jenkins_node" json:"jenkins_node"`
	IsConfirm     int       `gorm:"column:is_confirm" json:"is_confirm"`
	CpdsFlag      int       `gorm:"column:cpds_flag" json:"cpds_flag"`
	McpConfId     int       `gorm:"column:mcp_conf_id" json:"mcp_conf_id"`
}

func (UnitConfCntr) TableName() string {
	return "unit_conf_cntr"
}


// 非标准虚机应用列表
type UnitConfNvm struct {
	ID              int         `gorm:"column:id" json:"id"`
	UnitId          int         `gorm:"column:unit_id" json:"unit_id"`
	Hosts           string      `gorm:"column:hosts;comment:'应用主机'" json:"hosts"`
	AppUser         string      `gorm:"column:app_user;comment:'应用用户'" json:"app_user"`
	FilePath        string      `gorm:"column:file_path" json:"file_path"`
	ShellContent    string      `gorm:"column:shell_content" json:"shell_content"`
	ShellPath       string      `gorm:"column:shell_path" json:"shell_path"`
	PathOrGene      string      `gorm:"column:path_or_gene" json:"path_or_gene"`
	InsertTime      string      `gorm:"column:insert_time" json:"insert_time"`
	IsDelete        int         `gorm:"column:is_delete" json:"is_delete"`
}

func (UnitConfNvm) TableName() string {
	return "unit_conf_nvm"
}
