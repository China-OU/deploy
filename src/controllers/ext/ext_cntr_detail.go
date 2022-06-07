package ext

import (
	"controllers/operation"
	"initial"
	"library/cfunc"
	"models"
)

// @Title CntrServiceDetail
// @Description 获取容器平台的应用信息及状态，支持多容器平台
// @Param	unit	query	string	true	"发布单元英文名，比如aml，会查找对应的service，返回service信息"
// @Success 200 true or false
// @Failure 403
// @router /cpds/cntr-detail [get]
func (c *ExtCpdsFuncController) CntrServiceDetail() {
	unit := c.GetString("unit")
	base := cfunc.GetUnitInfoByName(unit)
	if base.Unit == "" {
		c.SetJson(0, nil, "发布单元传参错误")
		return
	}

	var conf models.UnitConfMcp
	err := initial.DB.Model(models.UnitConfMcp{}).Where("unit_id=? and is_delete=0", base.Id).First(&conf).Error
	if err != nil {
		c.SetJson(0, nil, err.Error())
		return
	}
	if conf.ContainerType == "istio" {
		err, data := operation.IstioServiceDetail(conf)
		if err != nil {
			c.SetJson(0, nil, err.Error())
			return
		}
		data.Summary.UnitEn = base.Unit
		data.Summary.UnitCn = base.Name
		c.SetJson(1, data, "istio的服务数据获取成功！")
		return
	}

	if conf.ContainerType == "caas" {
		err, data := operation.CaasServiceDetail(conf)
		if err != nil {
			c.SetJson(0, nil, err.Error())
			return
		}
		data.Summary.UnitEn = base.Unit
		data.Summary.UnitCn = base.Name
		c.SetJson(1, data, "caas的服务数据获取成功！")
		return
	}

	if conf.ContainerType == "rancher" {
		err, data := operation.RancherServiceDetail(conf)
		if err != nil {
			c.SetJson(0, nil, err.Error())
			return
		}
		data.Summary.UnitEn = base.Unit
		data.Summary.UnitCn = base.Name
		c.SetJson(1, data, "caas的服务数据获取成功！")
		return
	}
	if conf.ContainerType == "openshift" {
		c.SetJson(0, nil, "功能开发中，敬请期待！")
		return
	}

	c.SetJson(0, nil, "没有匹配到容器平台类型！")
}
