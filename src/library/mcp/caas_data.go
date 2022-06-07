package mcp

// caas接口团队列表
type TeamList struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       TeamData     `json:"data"`
}

type TeamData struct {
	CurrentPage     int      `json:"currentPage"`
	PageSize        int      `json:"pageSize"`
	TotalSize       int      `json:"totalSize"`
	Data            []TeamDataDetail  `json:"data"`
}

type TeamDataDetail struct {
	Id           int      `json:"id"`
	Name         string   `json:"name"`
	MemberCount  int      `json:"memberCount"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
}

// caas接口集群列表
type CaasClustList struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       []ClustData  `json:"data"`
}

type ClustData struct {
	Id     int      `json:"id"`
	Name   string   `json:"name"`
	Uuid   string   `json:"uuid"`
	Des    string   `json:"des"`
	State  string   `json:"state"`
}

// caas接口堆栈列表
type CaasStackList struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       StackData    `json:"data"`
}

type StackData struct {
	CurrentPage     int      `json:"currentPage"`
	PageSize        int      `json:"pageSize"`
	TotalSize       int      `json:"totalSize"`
	Data            []StackDataDetail  `json:"data"`
}

type StackDataDetail struct {
	Id           string  `json:"id"`
	Uuid         string  `json:"uuid"`
	Description  string  `json:"description"`
	TenantId     string  `json:"tenantId"`
	State        string  `json:"state"`
	Name         string  `json:"name"`
	Ns           string  `json:"ns"`
}

// caas接口服务列表
type CaasServiceList struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       ServiceData    `json:"data"`
}

type ServiceData struct {
	CurrentPage     int      `json:"currentPage"`
	PageSize        int      `json:"pageSize"`
	TotalSize       int      `json:"totalSize"`
	Data            []ServiceDataDetail  `json:"data"`
}

type ServiceDataDetail struct {
	Id           string  `json:"id"`
	Uuid         string  `json:"uuid"`
	State        string  `json:"state"`
	Name         string  `json:"name"`
	Image        string  `json:"image"`
	Wise2cServiceType  string  `json:"wise2cServiceType"`
}

// caas接口 服务实例列表
type CaasInstanceList struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       []InstanceDataDetail    `json:"data"`
}

type InstanceDataDetail struct {
	Id           string  `json:"id"`
	Uuid         string  `json:"uuid"`
	Ip           string  `json:"ip"`
	Host         string  `json:"host"`
	State        string  `json:"state"`
	Name         string  `json:"name"`
	WarnMsg      string  `json:"warn_msg"`
}

// caas接口 service状态
type CaasServiceStatus struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       ServiceStatusDetail    `json:"data"`
}

type ServiceStatusDetail struct {
	Uuid         string  `json:"uuid"`
	State        string  `json:"state"`
	Name         string  `json:"name"`
	Image        string  `json:"image"`
	WarnMsg      string  `json:"warn_msg"`
}

// caas服务升级时，传给接口的参数
type ServiceUpgradeInput struct {
	StartFirst     string     `json:"startFirst"`
	LaunchConfig   LConfig    `json:"launchconfig"`
	InitContainers []string   `json:"initContainers"`
}

type LConfig struct {
	ImageUuid    string     `json:"imageUuid"`
}

// caas-api 升级时返回数据
type CaasApiOprRet struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       map[string]string     `json:"data"`
}