package mcp

// rancher接口返回数据结构差不多
type RancherRet struct {
	Type            string            `json:"type"`
	ResourceType    string            `json:"resourceType"`
	Actions         map[string]string `json:"actions"`
	Data            []RancherData      `json:"data"`
	Pagination      RancherPaginatio  `json:"pagination"`
}

type RancherData struct {
	Id            string             `json:"id"`
	Type          string             `json:"type"`
	Actions       map[string]string  `json:"actions"`
	Name          string 	         `json:"name"`
	State         string             `json:"state"`
	Kind          string             `json:"kind"`
	BaseType      string             `json:"baseType"`
	Created       string             `json:"created"`
	// service接口专有
	CurrentScale  int                `json:"currentScale"`
	Scale         int                `json:"scale"`
	// instance接口专有
	ImageUuid       string  `json:"imageUuid"`
	PrimaryIpAddress string   `json:"primaryIpAddress"`
}

type RancherPaginatio struct {
	Previous     string     `json:"previous"`
	Next         string     `json:"next"`
	Limit        int        `json:"limit"`
}

type RancherService struct {
	Id            string             `json:"id"`
	Type          string             `json:"type"`
	Actions       AcctionDetail      `json:"actions"`
	Name          string 	         `json:"name"`
	State         string             `json:"state"`
	Kind          string             `json:"kind"`
	BaseType      string             `json:"baseType"`
	Created       string             `json:"created"`
	CurrentScale  int                `json:"currentScale"`
	Scale         int                `json:"scale"`
	LaunchConfig  map[string]interface{}       `json:"launchConfig"`
}

type AcctionDetail struct {
	Upgrade        string `json:"upgrade"`
	FinishUpgrade  string `json:"finishupgrade"`
	Restart        string `json:"restart"`
}