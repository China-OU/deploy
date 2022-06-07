package info

import (
	"controllers"
	"initial"
	"models"
	"github.com/astaxie/beego"
	"fmt"
)

type CaasDetailController struct {
	controllers.BaseController
}

func (c *CaasDetailController) URLMapping() {
	c.Mapping("GetTeamList", c.GetTeamList)
	c.Mapping("GetClustList", c.GetClustList)
	c.Mapping("GetStackList", c.GetStackList)
	c.Mapping("GetServiceList", c.GetServiceList)
}

// @Title GetTeamList
// @Description 根据租户和网络区域，获取团队列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	deploy_network	query	string	true	"部署网络区域"
// @Success 200 {object} []models.CaasTeam
// @Failure 403
// @router /caas/teamlist [get]
func (c *CaasDetailController) GetTeamList() {
	deploy_comp := c.GetString("deploy_comp")
	//deploy_network := c.GetString("deploy_network")
	var caas_conf models.CaasConf
	var cnt int
	err := initial.DB.Table("conf_caas").Where("deploy_comp=? and is_delete=0", deploy_comp).
		Count(&cnt).First(&caas_conf).Error
	if cnt == 0 {
		c.SetJson(0, "", "租户选择有误，请重新选择！")
		return
	}
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	var team_list []models.CaasTeam
	err = initial.DB.Table("conf_caas_detail").Where("caas_id=? and is_delete=0", caas_conf.Id).Select("team_id, team_name").
		Group("team_id, team_name").Limit(10).Find(&team_list).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, team_list, "caas团队获取成功！")
}

// @Title GetClustList
// @Description 根据租户和网络区域，获取集群列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	deploy_network	query	string	true	"部署网络区域"
// @Success 200 {object} []models.CaasCluster
// @Failure 403
// @router /caas/clustlist [get]
func (c *CaasDetailController) GetClustList() {
	deploy_comp := c.GetString("deploy_comp")
	//deploy_network := c.GetString("deploy_network")
	var caas_conf models.CaasConf
	var cnt int
	err := initial.DB.Table("conf_caas").Where("deploy_comp=? and is_delete=0", deploy_comp).
		Count(&cnt).First(&caas_conf).Error
	if cnt == 0 {
		c.SetJson(0, "", "租户选择有误，请重新选择！")
		return
	}
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	var data []models.CaasCluster
	err = initial.DB.Table("conf_caas_detail").Where("caas_id=? and is_delete=0", caas_conf.Id).
		Select("cluster_id, cluster_name, cluster_uuid").Group("cluster_id, cluster_name, cluster_uuid").
		Limit(10).Find(&data).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	var ret []models.CaasCluster
	for _, v := range data {
		// 容器初始化时，会生成clust_id=0的数据，第二天同步后，该数据会更新，此时拉取时需要过滤
		if v.ClusterId != "0" {
			ret = append(ret, v)
		}
	}
	c.SetJson(1, ret, "caas集群获取成功！")
}

// @Title GetStackList
// @Description 根据租户和网络区域，team_id和clust_uuid，获取堆栈列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	deploy_network	query	string	true	"部署网络区域"
// @Param	team_id	query	string	true	"团队id"
// @Param	clust_uuid	query	string	true	"集群uuid"
// @Param	search	query	string	false	"模糊查询"
// @Success 200 {object} []models.CaasStack
// @Failure 403
// @router /caas/stacklist [get]
func (c *CaasDetailController) GetStackList() {
	deploy_comp := c.GetString("deploy_comp")
	//deploy_network := c.GetString("deploy_network")
	team_id := c.GetString("team_id")
	clust_uuid := c.GetString("clust_uuid")
	search := c.GetString("search")
	var caas_conf models.CaasConf
	var cnt int
	err := initial.DB.Table("conf_caas").Where("deploy_comp=? and is_delete=0", deploy_comp).
		Count(&cnt).First(&caas_conf).Error
	if cnt == 0 {
		c.SetJson(0, "", "租户选择有误，请重新选择！")
		return
	}
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	cond := fmt.Sprintf("caas_id=%d and team_id='%s' and cluster_uuid='%s' and is_delete=0", caas_conf.Id, team_id, clust_uuid)
	if search != "" {
		cond += fmt.Sprintf(" and stack_name like '%%%s%%'", search)
	}
	var data []models.CaasStack
	err = initial.DB.Table("conf_caas_detail").Where(cond).Select("stack_id, stack_uuid, stack_name").
		Group("stack_id, stack_uuid, stack_name").Limit(10).Find(&data).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	var ret []models.CaasStack
	for _, v := range data {
		// 容器初始化时，会生成clust_id=0的数据，第二天同步后，该数据会更新，此时拉取时需要过滤
		if v.StackId != "" {
			ret = append(ret, v)
		}
	}
	c.SetJson(1, ret, "caas堆栈获取成功！")
}

// @Title GetServiceList
// @Description 根据租户和网络区域，team_id、clust_uuid和stack_name，获取堆栈列表获取服务列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	deploy_network	query	string	true	"部署网络区域"
// @Param	team_id	query	string	true	"团队id"
// @Param	clust_uuid	query	string	true	"集群uuid"
// @Param	stack_name	query	string	true	"堆栈名"
// @Param	search	query	string	false	"模糊查询"
// @Success 200 {object} []models.CaasService
// @Failure 403
// @router /caas/servicelist [get]
func (c *CaasDetailController) GetServiceList() {
	deploy_comp := c.GetString("deploy_comp")
	//deploy_network := c.GetString("deploy_network")
	team_id := c.GetString("team_id")
	clust_uuid := c.GetString("clust_uuid")
	stack_name := c.GetString("stack_name")
	search := c.GetString("search")
	var caas_conf models.CaasConf
	var cnt int
	err := initial.DB.Table("conf_caas").Where("deploy_comp=? and is_delete=0", deploy_comp).
		Count(&cnt).First(&caas_conf).Error
	if cnt == 0 {
		c.SetJson(0, "", "租户选择有误，请重新选择！")
		return
	}
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	cond := fmt.Sprintf("caas_id=%d and team_id='%s' and cluster_uuid='%s' and stack_name='%s' and is_delete=0",
		caas_conf.Id, team_id, clust_uuid, stack_name)
	if search != "" {
		cond += fmt.Sprintf(" and service_name like '%%%s%%'", search)
	}
	var data []models.CaasService
	err = initial.DB.Table("conf_caas_detail").Where(cond).Select("service_id, service_uuid, service_name").
		Group("service_id, service_uuid, service_name").Limit(50).Find(&data).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, data, "caas服务列表获取成功！")
}
