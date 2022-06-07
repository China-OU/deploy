package caas_conf

import (
	"controllers"
	"strings"
	"fmt"
	"initial"
	"models"
	"github.com/astaxie/beego"
	"library/cfunc"
	"github.com/jinzhu/gorm"
)

type CaasDataListController struct {
	controllers.BaseController
}

func (c *CaasDataListController) URLMapping() {
	c.Mapping("GetAll", c.GetAll)
}

// @Title Get All Caas Service
// @Description 获取所有容器服务列表，展示即可，数据已从caas平台获取。caas管理平台不分网络区域，一个租户只有一套caas管理平台，在每个网络区域都会部署一套k8s
// @Param	deploy_comp	query	string	false	"部署租户"
// @Param	team	query	string	false	"项目团队"
// @Param	cluster	query	string	false	"容器集群"
// @Param	stack	query	string	false	"容器堆栈"
// @Param	service	query	string	false	"容器服务"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} []models.CaasDetailRet
// @Failure 403
// @router /caas/list [get]
func (c *CaasDataListController) GetAll() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	deploy_comp := c.GetString("deploy_comp")
	team := c.GetString("team")
	cluster := c.GetString("cluster")
	stack := c.GetString("stack")
	service := c.GetString("service")
	page, rows := c.GetPageRows()
	cond := " is_delete=0 "
	if deploy_comp != "" {
		var caas_conf models.CaasConf
		var cnt int
		err := initial.DB.Table("conf_caas").Where("deploy_comp=?  and is_delete=0", deploy_comp).
			Count(&cnt).First(&caas_conf).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			c.SetJson(0, "", err.Error())
			return
		}
		if cnt == 0 {
			c.SetJson(0, "", "部署租户下无容器服务!")
			return
		}
		if cnt > 0 {
			cond += fmt.Sprintf(" and caas_id = %d ", caas_conf.Id)
		}
	}
	if strings.TrimSpace(team) != "" {
		cond += fmt.Sprintf(" and team_name like '%%%s%%' ", team)
	}
	if strings.TrimSpace(cluster) != "" {
		cond += fmt.Sprintf(" and cluster_name like '%%%s%%' ", cluster)
	}
	if strings.TrimSpace(stack) != "" {
		cond += fmt.Sprintf(" and stack_name like '%%%s%%' ", stack)
	}
	if strings.TrimSpace(service) != "" {
		cond += fmt.Sprintf(" and service_name like '%%%s%%' ", service)
	}
	var cnt int
	var ulist []models.CaasConfDetail
	err := initial.DB.Model(models.CaasConfDetail{}).Where(cond).Count(&cnt).Offset((page - 1)*rows).Limit(rows).Find(&ulist).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

 	var ret_data []models.CaasDetailRet
 	for _, v := range ulist {
 		cc := cfunc.GetCompNetworkById(v.CaasId)
 		ret_data = append(ret_data, models.CaasDetailRet{
 			cc.DeployComp,
 			cc.DeployNetwork,
 			v,
		})
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": ret_data,
	}
	c.SetJson(1, ret, "数据获取成功！")
}
