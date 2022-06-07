package models

import "time"

type UnitConfVM struct {
	ID				int			`gorm:"column:id" json:"id"`
	AppType			string		`gorm:"column:app_type" json:"app_type"`
	AppSubType		string		`gorm:"column:app_sub_type" json:"app_sub_type"`
	UnitID			int			`gorm:"column:unit_id" json:"unit_id"`
	GitID			string		`gorm:"column:git_id" json:"git_id"`
	GitUnit			string		`gorm:"column:git_unit" json:"git_unit"`
	GitURL			string		`gorm:"column:git_url" json:"git_url"`
	DeployType		string		`gorm:"column:deploy_type;comment:'应用部署类型，如jar'" json:"deploy_type"`
	DeployComp		string		`gorm:"column:deploy_comp;comment:'所属租户'" json:"deploy_comp"`
	DeployENV		string		`gorm:"column:deploy_env;comment:'部署环境，PRD,DEV,SIT,UAT等'" json:"deploy_env"`
	DeployVPC		string		`gorm:"column:deploy_vpc;comment:'网络安全区域，CORE,DMZ,PTR等'" json:"deploy_vpc"`
	Artifact		string		`gorm:"column:artifact;comment:'应用包文件名'" json:"artifact"`
	Hosts			string		`gorm:"column:hosts;comment:'应用主机'" json:"hosts"`
	AppUser			string		`gorm:"column:app_user;comment:'应用用户'" json:"app_user"`
	AppPath			string		`gorm:"column:app_path;comment:'应用路径'" json:"app_path"`
	AppBindProt		string		`gorm:"column:app_bind_prot;comment:'应用传输协议，如TCP,UDP'" json:"app_bind_prot"`
	AppBindPort		string		`gorm:"column:app_bind_port;comment:'应用监听端口，;分割'" json:"app_bind_port"`
	AppTempPath		string		`gorm:"column:app_temp_path;comment:'应用临时路径'" json:"app_temp_path"`
	AppBackupPath	string		`gorm:"column:app_backup_path;comment:'应用备份路径'" json:"app_backup_path"`
	NeedReboot		int			`gorm:"column:need_reboot;comment:'是否需要重启应用或中间件'" json:"need_reboot"`
	CMDPre			string		`gorm:"column:cmd_pre;comment:'启动时前置命令'" json:"cmd_pre"`
	CMDStop			string		`gorm:"column:cmd_stop;comment:'停止应用命令'" json:"cmd_stop"`
	CMDStartup		string		`gorm:"column:cmd_startup;comment:'应用启动命令'" json:"cmd_startup"`
	CMDRear			string		`gorm:"column:cmd_rear;comment:'启动时后置命令'" json:"cmd_rear"`
	BeforeXml		string		`gorm:"column:before_xml;comment:'构建时前置配置'" json:"before_xml"`
	BuildXml		string		`gorm:"column:build_xml;comment:'构建配置'" json:"build_xml"`
	AfterXml		string		`gorm:"column:after_xml;comment:'构建时后置配置'" json:"after_xml"`
	JenkinsNode		string		`gorm:"column:jenkins_node;comment:'构建节点'" json:"jenkins_node"`
	ArtifactPath	string		`gorm:"column:artifact_path;comment:'制品包路径'" json:"artifact_path"`
	IsConfirm		int			`gorm:"column:is_confirm;comment:'开发是否确认'" json:"is_confirm"`
	InsertTime		string		`gorm:"column:insert_time" json:"insert_time"`
	UpdateTime		time.Time	`gorm:"column:update_time;default:current_time on update current_time" json:"update_time"`
	IsDelete		int			`gorm:"column:is_delete" json:"is_delete"`
	//PKGPath         string      `gorm:"column:pkg_path" json:"pkg_path"`
}

func (UnitConfVM) TableName() string {
	return "unit_conf_vm"
}