package caas_cntr

import (
	"github.com/astaxie/beego"
	"library/caas"
	"models"
	"strings"
)

// 同步初始化任务配置
// @Title CntrConfigSync()
// @Description 同步镜像和实例数、下发状态
// @Param id query int false "容器初始化记录 id"
// @Success 200 {object} models.OprCntrInit
// @Failure 403
// @router /cntr/init/:id/sync [get]
func (c *CntrController)SyncCntrInitStatus() {
	id, _ := c.GetInt("id", 0)
	if id == 0 {
		c.SetJson(0, "", "id 错误")
		return
	}
	isAdmin := false
	if strings.Contains(c.Role, "admin") {
		isAdmin = true
	}
	item, err := models.OprCntrInit{}.GetOneById(isAdmin, c.UserId, id)
	if err != nil {
		c.SetJson(0, "", "记录不存在或权限不足:"+err.Error())
		return
	}
	var comp models.CaasConf
	comp.DeployComp = item.Comp
	hostStringList := strings.Split(item.Agent, ":")
	if len(hostStringList) != 2 {
		c.SetJson(0, "", "获取 agent 配置错误")
		return
	}
	comp.AgentIp, comp.AgentPort = hostStringList[0], hostStringList[1]
	caasOpr := caas.CaasOpr{
		AgentConf:   comp,
		TeamId:      item.TeamId,
		ClustUuid:   item.ClusterUuid,
		StackName:   item.StackName,
		ServiceName: item.ServiceName,
	}
	err, serviceConfig := caasOpr.RetryGetServiceConfig(5)
	if err != nil {
		beego.Error(err)
		c.SetJson(0, "", err.Error() )
		return
	}
	err, serviceStatus := caasOpr.RetryGetService(5)
	if err != nil {
		beego.Error(err)
		c.SetJson(0, "", err.Error() )
		return
	}
	err, updatedCntrConfig  := UpdateOrCreateConfig(item.UnitId, serviceConfig, serviceStatus, &caasOpr)
	if err != nil {
		beego.Error(err)
		c.SetJson(0, "", err.Error() )
		return
	}
	c.SetJson(1, updatedCntrConfig, "同步成功！")
}
