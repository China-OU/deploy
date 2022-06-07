package unit_conf

import (
	"models"
	"initial"
)

func GetMcpConfById(mcp_id int) (error, models.UnitConfMcp) {
	var mcp models.UnitConfMcp
	err := initial.DB.Model(models.UnitConfMcp{}).Where("id=? and is_delete=0", mcp_id).First(&mcp).Error
	if err != nil {
		return err, models.UnitConfMcp{}
	}
	return nil, mcp
}

func GetIstioConfDetail(mcp_id int) (error, models.McpConfIstio) {
	var istio models.McpConfIstio
	err := initial.DB.Model(models.McpConfIstio{}).Where("mcp_id=? and is_delete=0", mcp_id).First(&istio).Error
	if err != nil {
		return err, models.McpConfIstio{}
	}
	return nil, istio
}

func GetCaasConfDetail(mcp_id int) (error, models.McpConfCaas) {
	var caas models.McpConfCaas
	err := initial.DB.Model(models.McpConfCaas{}).Where("mcp_id=? and is_delete=0", mcp_id).First(&caas).Error
	if err != nil {
		return err, models.McpConfCaas{}
	}
	return nil, caas
}

func GetRancherConfDetail(mcp_id int) (error, models.McpConfRancher) {
	var rancher models.McpConfRancher
	err := initial.DB.Model(models.McpConfRancher{}).Where("mcp_id=? and is_delete=0", mcp_id).First(&rancher).Error
	if err != nil {
		return err, models.McpConfRancher{}
	}
	return nil, rancher
}
