package models

// 多容器平台总表
type UnitConfMcp struct {
	Id             int       `gorm:"column:id" json:"id"`
	UnitId         int       `gorm:"column:unit_id" json:"unit_id"`
	AppType        string    `gorm:"column:app_type" json:"app_type"`
	AppSubType     string    `gorm:"column:app_sub_type" json:"app_sub_type"`
	DeployComp     string    `gorm:"column:deploy_comp" json:"deploy_comp"`
	DeployNetwork  string    `gorm:"column:deploy_network" json:"deploy_network"`
	ContainerType  string    `gorm:"column:container_type" json:"container_type"`
	IsDelete       int       `gorm:"column:is_delete" json:"is_delete"`
	InsertTime     string    `gorm:"column:insert_time" json:"insert_time"`
}

func (UnitConfMcp) TableName() string {
	return "unit_conf_mcp"
}

// k8s-istio应用配置表
type McpConfIstio struct {
	Id             int       `gorm:"column:id" json:"id"`
	McpId          int       `gorm:"column:mcp_id" json:"mcp_id"`
	Namespace      string    `gorm:"column:namespace" json:"namespace"`
	Deployment     string    `gorm:"column:deployment" json:"deployment"`
	Version        string    `gorm:"column:version" json:"version"`
	Container      string    `gorm:"column:container" json:"container"`
	Operator       string    `gorm:"column:operator" json:"operator"`
	IsDelete       int       `gorm:"column:is_delete" json:"is_delete"`
	InsertTime     string    `gorm:"column:insert_time" json:"insert_time"`
}

func (McpConfIstio) TableName() string {
	return "mcp_conf_istio"
}

// mcp-caas 应用配置表
type McpConfCaas struct {
	Id             int       `gorm:"column:id" json:"id"`
	McpId          int       `gorm:"column:mcp_id" json:"mcp_id"`
	TeamId         string    `gorm:"column:team_id" json:"team_id"`
	TeamName       string    `gorm:"column:team_name" json:"team_name"`
	ClusterUuid    string    `gorm:"column:cluster_uuid" json:"cluster_uuid"`
	ClusterName    string    `gorm:"column:cluster_name" json:"cluster_name"`
	StackName      string    `gorm:"column:stack_name" json:"stack_name"`
	ServiceName    string    `gorm:"column:service_name" json:"service_name"`
	Operator       string    `gorm:"column:operator" json:"operator"`
	IsDelete       int       `gorm:"column:is_delete" json:"is_delete"`
	InsertTime     string    `gorm:"column:insert_time" json:"insert_time"`
}

func (McpConfCaas) TableName() string {
	return "mcp_conf_caas"
}

// rancher 应用配置表
type McpConfRancher struct {
	Id             int       `gorm:"column:id" json:"id"`
	McpId          int       `gorm:"column:mcp_id" json:"mcp_id"`
	ProjectId      string    `gorm:"column:project_id" json:"project_id"`
	ProjectName    string    `gorm:"column:project_name" json:"project_name"`
	StackId        string    `gorm:"column:stack_id" json:"stack_id"`
	StackName      string    `gorm:"column:stack_name" json:"stack_name"`
	ServiceId      string    `gorm:"column:service_id" json:"service_id"`
	ServiceName    string    `gorm:"column:service_name" json:"service_name"`
	Operator       string    `gorm:"column:operator" json:"operator"`
	IsDelete       int       `gorm:"column:is_delete" json:"is_delete"`
	InsertTime     string    `gorm:"column:insert_time" json:"insert_time"`
}

func (McpConfRancher) TableName() string {
	return "mcp_conf_rancher"
}

// openshift 应用配置表
// ...


// 多容器平台升级表
type McpUpgradeList struct {
	Id            int       `gorm:"column:id" json:"id"`
	UnitId        int       `gorm:"column:unit_id" json:"unit_id"`
	OldImage      string    `gorm:"column:old_image" json:"old_image"`
	NewImage      string    `gorm:"column:new_image" json:"new_image"`
	Result        int       `gorm:"column:result" json:"result"`
	Operator      string    `gorm:"column:operator" json:"operator"`
	Message       string    `gorm:"column:message" json:"message"`
	OnlineDate    string    `gorm:"column:online_date" json:"online_date"`
	CostTime      int       `gorm:"column:cost_time" json:"cost_time"`
	InsertTime    string    `gorm:"column:insert_time" json:"insert_time"`
	SourceId      string    `gorm:"column:source_id" json:"source_id"`
}

func (McpUpgradeList) TableName() string {
	return "mcp_upgrade_list"
}













// 多容器平台agent配置表
type McpConfAgent struct {
	Id             int       `gorm:"column:id" json:"id"`
	AgentId        int       `gorm:"column:agent_id" json:"agent_id"`
	McpType        string    `gorm:"column:mcp_type" json:"mcp_type"`
	BaseUrl        string    `gorm:"column:base_url" json:"base_url"`
	Token          string    `gorm:"column:token" json:"token"`
	CheckSurv      string    `gorm:"column:check_surv" json:"check_surv"`
	CheckTime      string    `gorm:"column:check_time" json:"check_time"`
	Operator       string    `gorm:"column:operator" json:"operator"`
	IsDelete       int       `gorm:"column:is_delete" json:"is_delete"`
	InsertTime     string    `gorm:"column:insert_time" json:"insert_time"`
}

func (McpConfAgent) TableName() string {
	return "mcp_conf_agent"
}