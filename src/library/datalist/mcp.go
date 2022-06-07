package datalist

import "models"

type McpCommonInput struct {
	BaseInfo  models.UnitConfMcp `json:"base_info"`
}

type IstioConfInput struct {
	BaseInfo  models.UnitConfMcp `json:"base_info"`
	IstioInfo models.McpConfIstio `json:"istio_info"`
}

type CaasConfInput struct {
	BaseInfo  models.UnitConfMcp `json:"base_info"`
	CaasInfo  models.McpConfCaas `json:"caas_info"`
}

type RancherConfInput struct {
	BaseInfo     models.UnitConfMcp    `json:"base_info"`
	RancherInfo  models.McpConfRancher `json:"rancher_info"`
}
