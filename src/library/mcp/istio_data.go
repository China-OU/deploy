package mcp

// 通用数据返回
type CommonRet struct {
	Code  int      `json:"code"`
	Msg   string   `json:"msg"`
	Data  string   `json:"data"`
}

// 命名空间
type NamespaceData struct {
	Kind        string           `json:"kind"`
	ApiVersion  string           `json:"apiVersion"`
	Items       []NamespaceItem  `json:"items"`
}

type NamespaceItem struct {
	Metadata NamespaceMetadata `json:"metadata"`
	Status   NamespacesSatus `json:"status"`
}

type NamespaceMetadata struct {
	Name  string  `json:"name"`
	Uid   string  `json:"uid"`
}

type NamespacesSatus struct {
	Phase  string   `json:"phase"`
}

// 部署数据结构
type DeploymentMetadata struct {
	Name   string             `json:"name"`
	Uid    string             `json:"uid"`
	Labels DeploymentLabel    `json:"labels"`
}
type DeploymentLabel struct {
	App        string    `json:"app"`
	Service    string    `json:"io.wise2c.service"`
	Version    string    `json:"version"`
}

type SpecData struct {
	Replicas   int   `json:"replicas"`
	Template  SpecMetadata    `json:"template"`
}
type SpecMetadata struct {
	Spec  SpecContainer `json:"spec"`
}
type SpecContainer struct {
	Containers  []SpecContainerInfo   `json:"containers"`
}
type SpecContainerInfo struct {
	Name  string  `json:"name"`
	Image string  `json:"image"`
}

type DeploymentRet struct {
	Name   string        `json:"name"`
	App        string    `json:"app"`
	Version    string    `json:"version"`
	Container  string    `json:"container"`
}

// pod状态
type PodStatus struct {
	Phase string `json:"phase"`
	HostIP string `json:"hostIP"`
	PodIP string `json:"podIP"`
	StartTime string `json:"startTime"`
}

type PodRet struct {
	PodName   string        `json:"pod_name"`
	Image     string        `json:"image"`
	PodIP     string        `json:"pod_ip"`
	StartTime string        `json:"start_time"`
	Status    string        `json:"status"`
}

// status返回
type StatusData struct {
	Spec     SpecData   `json:"spec"`
	Status   StatusRep   `json:"status"`
}

type StatusRep struct {
	Replicas        int `json:"replicas"`
	UpdatedReplicas int `json:"updatedReplicas"`
}

// 镜像升级的body参数
type UpgradeInput struct {
	Spec SpecInput `json:"spec"`
}
type SpecInput struct {
	Template SpecMetadata `json:"template"`
}





// istio通用数据结构返回
type IstioData struct {
	Kind        string           `json:"kind"`
	ApiVersion  string           `json:"apiVersion"`
	Items       []ItemData  `json:"items"`
}

type ItemData struct {
	Metadata interface{}   `json:"metadata"`
	Spec     interface{}   `json:"spec"`
	Status   interface{}   `json:"status"`
}