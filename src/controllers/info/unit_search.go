package info

import (
	"controllers"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"models"
)

type UnitDetailController struct {
	controllers.BaseController
}

func (c *UnitDetailController) URLMapping() {
	c.Mapping("UnitSearch", c.UnitSearch)
	c.Mapping("UnitDeployComp", c.UnitDeployComp)
}

// @Title Get Unit Name By Search
// @Description 通过搜索获取发布单元，最多只返回10条数据
// @Param	search	query	string	true	"发布单元名，支持模糊搜索"
// @Success 200 {object} []models.UnitConfList
// @Failure 403
// @router /unit/search [get]
func (c *UnitDetailController) UnitSearch() {
	search := c.GetString("search")
	var unit_list []models.UnitConfList
	cond := fmt.Sprintf("is_offline = 0 and unit like '%%%s%%'", search)
	err := initial.DB.Model(models.UnitConfList{}).Where(cond).Limit(10).Find(&unit_list).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, unit_list, "发布单元获取成功！")
}

// @Title Get Unit Deploy Company
// @Description 获取发布单元可能部署的租户，返回所有数据，后续如果数据过多再分页
// @Success 200 {object} []models.DeployComp
// @Failure 403
// @router /unit/dcomp [get]
func (c *UnitDetailController) UnitDeployComp() {
	var dc []models.DeployComp
	err := initial.DB.Table("unit_conf_list").Select("distinct(dumd_comp_en)").Find(&dc).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 数据清理
	var ret []models.DeployComp
	for _, v := range dc {
		if v.CompEn == "" || v.CompEn == "nil" || v.CompEn == "<nil>" {
			continue
		}
		ret = append(ret, v)
	}
	// 增加扩展网络区域，比如CMHK有cmhk、sy和sdn好几个区域，这在受益人里面只对应一个，需要手动添加
	var vpc []models.VpcExt
	initial.DB.Model(models.VpcExt{}).Where("is_delete=0").Find(&vpc)
	for _, v := range vpc {
		per := models.DeployComp{ CompEn: v.Vpc }
		flag := false
		for _, k := range ret {
			if per.CompEn == k.CompEn {
				flag = true
				break
			}
		}
		if flag == false {
			ret= append(ret, per)
		}
	}
	c.SetJson(1, ret, "部署租户获取成功！")
}
