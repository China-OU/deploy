package caas_cntr

import (
	"controllers/operation"
	"github.com/astaxie/beego"
	"initial"
	"library/caas"
	"models"
	"strings"
)

// 编辑更新现有服务：日志收集规则、环境变量、健康检查、弹性伸缩等
// @Title EditService
// @Description 更新堆栈服务配置,此接口保留，今后可以根据需求灵活开启
// @Param	body	body	caas.InitServiceWebData 	true	"json body"
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/config [post]
func (c *CntrController) CntrConfigUpdate() {
	c.SetJson(0, "", "暂未开放此接口")
	//c.IsEdit = true
	//c.InitService()
}

// 获取现有服务的所有配置
// @Title CntrConfigGet()
// @Description  输入发布单元ID，返回堆栈服务最新的配置
// @Param unit_id query int false "容器服务详情ID"
// @Success 200 {object} caas.InitServiceAgentData
// @Failure 403
// @router /cntr/config [get]
func (c *CntrController) CntrConfigGet() {
	if !strings.Contains(c.Role, "admin")  {
		c.SetJson(0, "", "只有管理员才有此权限!")
		return
	}
	unitId := c.GetString("unit_id")
	err, unitCntr := operation.GetCntrConfig(unitId)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if unitCntr.ServiceName == "" {
		c.SetJson(0, "", "该发布单元还未关联容器服务请先关联！")
		return
	}
	err, caasConf := operation.GetCaasConfig(unitCntr.DeployComp)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error()+ "，请检查agent的配置是否正确！")
		return
	}
	caasOpr := caas.CaasOpr{
		AgentConf:   caasConf,
		TeamId:      unitCntr.CaasTeam,
		ClustUuid:   unitCntr.CaasCluster,
		StackName:   unitCntr.CaasStack,
		ServiceName: unitCntr.ServiceName,
	}
	err, data := caasOpr.RetryGetServiceConfig(5)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, data, "")
}

// 同步现有服务的所有配置到本地
// @Title CntrConfigSync()
// @Description 同步堆栈服务最新的实例数和镜像地址配置
// @Success 200 {object} caas.InitServiceAgentData
// @Failure 403
// @router /cntr/config/sync [get]
func (c *CntrController) CntrConfigSync() {
	if !strings.Contains(c.Role, "admin")  {
		c.SetJson(0, "", "只有管理员才有此权限!")
		return
	}
	var unitCntrList []models.OprCntrInit
	var caasConfList []models.CaasConf
	con := initial.DB
	if err := con.Table("opr_cntr_init").Select([]string{`unit_id`, `team_id`, `cluster_uuid`, `stack_name`, `service_name`, `comp`}).
		Where("result = 1").Find(&unitCntrList).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if err := con.Table("conf_caas").Select([]string{"deploy_comp", "agent_ip", "agent_port"}).
		Where("is_delete = '0'").Find(&caasConfList).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	var caasConf = make(map[string]models.CaasConf,8)
	for _, conf := range caasConfList {
		caasConf[conf.DeployComp] = conf
	}
	go func() {
		// goroutine panic catcher
		defer func() {
			if err := recover(); err != nil {
				beego.Error("Panic error:", err)
			}
		}()
		for _, unit := range unitCntrList {
			if comp, ok := caasConf[unit.Comp]; !ok  {
				continue
			} else {
				caasOpr := caas.CaasOpr{
					AgentConf:   comp,
					TeamId:      unit.TeamId,
					ClustUuid:   unit.ClusterUuid,
					StackName:   unit.StackName,
					ServiceName: unit.ServiceName,
				}
				err, serviceConfig := caasOpr.RetryGetServiceConfig(5)
				if err != nil {
					beego.Error(err)
					continue
				}
				err, serviceStatus := caasOpr.RetryGetService(5)
				if err != nil {
					beego.Error(err)
					continue
				}
				if err, _ := UpdateOrCreateConfig(unit.UnitId, serviceConfig, serviceStatus, &caasOpr); err != nil {
					beego.Error(err)
				}
			}

		}
	}()
	c.SetJson(1, "", "同步配置任务启动，请等待！")
}


func UpdateOrCreateConfig(unitId int, serviceConfig *caas.ServiceConfigAll,
	serviceStatus *caas.ServiceStatusDetail,caas  *caas.CaasOpr) (error, *models.OprCntrInit) {
	var cntrConfig models.OprCntrInit
	if err := initial.DB.Model(models.OprCntrInit{}).First(&cntrConfig,"is_delete = 0 and unit_id=?", unitId).Error; err != nil {
		return err, nil
	}
	// 只更新容器镜像地址和实例数
	updateMap := make(map[string]interface{}, 5)
	if serviceStatus.State == "active" {
		updateMap["result"] = 1
		updateMap["message"] = "下发成功"
	}
	updateMap["state"] = serviceStatus.State
	updateMap["image"] = serviceConfig.Image
	updateMap["instance_num"] = serviceConfig.Scaling.DefaultInstances
	tx := initial.DB.Begin()
	if err := tx.Model(&cntrConfig).Updates(updateMap).Error; err != nil {
		tx.Rollback()
		return err, nil
	}
	tx.Commit()
	return nil, &cntrConfig
}