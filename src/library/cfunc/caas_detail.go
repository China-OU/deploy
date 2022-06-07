package cfunc

import (
	"initial"
	"models"
	"github.com/astaxie/beego"
)

func GetCompNetworkById(caas_id int) models.CaasConf {
	var cc models.CaasConf
	err := initial.DB.Model(models.CaasConf{}).Where("id=?", caas_id).Find(&cc).Error
	if err != nil {
		beego.Error(err.Error())
		return models.CaasConf{}
	}
	return cc
}

func GetAgentConfByComp(comp string) (models.CaasConf, error) {
	var cc models.CaasConf
	err := initial.DB.Model(models.CaasConf{}).Where("is_delete=0 and deploy_comp=?", comp).First(&cc).Error
	if err != nil {
		beego.Error(err.Error())
		return models.CaasConf{}, err
	}
	return cc, nil
}
