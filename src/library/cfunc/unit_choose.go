package cfunc

import (
	"models"
	"initial"
	"github.com/astaxie/beego"
)

// 通过unit_id搜索发布单元信息
func GetUnitInfoById(id int) models.UnitConfList {
	var cc models.UnitConfList
	err := initial.DB.Model(models.UnitConfList{}).Where("id=?", id).First(&cc).Error
	if err != nil {
		beego.Error(err.Error())
		return models.UnitConfList{}
	}
	return cc
}

// 通过发布单元英文名搜索发布单元信息
func GetUnitInfoByName(unit string) models.UnitConfList {
	var cc models.UnitConfList
	err := initial.DB.Model(models.UnitConfList{}).Where("`unit`= ? AND `is_offline` = 0", unit).First(&cc).Error
	if err != nil {
		beego.Error(err.Error())
		return models.UnitConfList{}
	}
	return cc
}


//// 通过负责人搜索获取发布单元列表
func GetContainerTypeByUnitId(unit_id int) (error, models.UnitConfMcp) {
	var conf models.UnitConfMcp
	err := initial.DB.Model(models.UnitConfMcp{}).Where("unit_id=? and is_delete=0", unit_id).First(&conf).Error
	if err != nil {
		return err, models.UnitConfMcp{}
	}
	return nil, conf
}
