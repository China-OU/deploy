package info

import (
	"controllers"
	"initial"
	"models"
	"github.com/astaxie/beego"
)

type UnitTypeInfoController struct {
	controllers.BaseController
}

func (c *UnitTypeInfoController) URLMapping() {
	c.Mapping("GetType", c.GetType)
	c.Mapping("GetSubType", c.GetSubType)
}

// @Title Get Unit Type Info
// @Description 获取发布单元类型，app/web等，最多只返回10条数据
// @Success 200 {object} []models.UnitType
// @Failure 403
// @router /unit/type [get]
func (c *UnitTypeInfoController) GetType() {
	var unit_type_list []models.UnitType
	err := initial.DB.Table("unit_conf_list").Select("distinct(app_type)").Limit(10).Find(&unit_type_list).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 数据清理
	var ret []models.UnitType
	for _, v := range unit_type_list {
		if v.AppType == "" || v.AppType == "nil" || v.AppType == "<nil>" {
			continue
		}
		ret = append(ret, v)
	}
	c.SetJson(1, ret, "发布单元类型获取成功！")
}

// @Title Get Unit Type Info
// @Description 获取发布单元子类型，app/web等，最多只返回10条数据
// @Param	app_type	query	string	true	"发布单元类型"
// @Success 200 {object} []models.UnitSubType
// @Failure 403
// @router /unit/subtype [get]
func (c *UnitTypeInfoController) GetSubType() {
	app_type := c.GetString("app_type")
	var sub_type_list []models.UnitSubType
	err := initial.DB.Table("unit_conf_list").Select("distinct(app_sub_type)").Where("app_type = ?", app_type).
		Limit(10).Find(&sub_type_list).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 数据清理
	var ret []models.UnitSubType
	for _, v := range sub_type_list {
		if v.AppSubType == "" || v.AppSubType == "nil" || v.AppSubType == "<nil>" {
			continue
		}
		ret = append(ret, v)
	}
	c.SetJson(1, sub_type_list, "发布单元类型获取成功！")
}
