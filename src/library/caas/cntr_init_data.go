package caas

import (
	"models"
)

// 首次部署，初始化容器服务, web 接口
type InitServiceWebData struct {
	UnitId        int         `json:"unit_id"`      // 发布单元ID
	AppType       string      `json:"app_type"`     // 发布单元类型
	TeamId        string      `json:"team_id"`      // 团队ID
	ClusterUuid   string      `json:"cluster_uuid"` // 集群uuid
	StackName     string      `json:"stack_name"`   // 堆栈名称
	ServiceName   string      `json:"service_name"` // 服务名称
	InstanceNum   int         `json:"instance_num"` // 实例数
	MemLimit      int         `json:"mem_limit"`    // 内存限制
	Cpu           int         `json:"cpu"`          // CPU 核数
	CpuMemConfigId int        `json:"cpu_mem_config_id"` // cpu内存配置ID
	Image         string      `json:"image"`        // 初始化镜像
	ClusterName   string      `json:"cluster_name"` // 集群名称
	TeamName      string      `json:"team_name"`    // 团队名称
	Comp          string      `json:"comp"`         // 租户
	Environment   *models.Environment `json:"environment"`  // 环境变量
	LogConfig     []*models.LogConfig `json:"logConfig"`
	HealthCheck   *models.HealthCheck `json:"healthCheck"`   // 健康检查详细
	Volume        []*models.Volume `json:"volume"` // 存储卷
	Scheduler     []*models.Scheduler `json:"scheduler"`
}

type InitServiceData struct {
	AgentConf models.CaasConf
	InitServiceWebData
}

type InitServiceAgentData struct {
	Image           string      `json:"image"`           // 镜像地址，必填
	ServiceName     string             `json:"serviceName"`     // 服务名称，必填
	//MemLimit        int                `json:"memLimit"`        // 内存限制，选填
	//CpuLimit        int                `json:"cpuLimit"`             // cpu核数，选填
	AlwaysPullImage string             `json:"alwaysPullImage"` // default "true"
	Environment     models.Environment `json:"environment"`
	LogConfig       []*models.LogConfig `json:"logConfig"`
	HealthCheck     *models.HealthCheck `json:"healthCheck"`
	Scaling         *models.Scaling     `json:"scalings"`
	Volume          []*models.Volume    `json:"volume"`
	Scheduler       []*models.Scheduler `json:"scheduler"`
}



type InitServiceAgentDataWithMemLimit struct {
	InitServiceAgentData
	MemLimit        int                `json:"memLimit"`        // 内存限制，选填

}

type ServiceConfigAll struct {
	InitServiceAgentDataWithMemLimit
	CpuLimit        int                 `json:"cpuLimit"`
}
