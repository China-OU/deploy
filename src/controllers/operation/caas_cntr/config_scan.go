package caas_cntr

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"initial"
	"library/caas"
	"models"
	"strings"
	"time"
)

// 扫描容器服务配置
// @Title CntrConfigGet()
// @Description  扫描堆栈服务配置
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/config/scan [post]
func (c *CntrController) CntrConfigScan() {
	if !strings.Contains(c.Role, "admin") {
		c.SetJson(0, "", "只有管理员才有此权限!")
		return
	}
	var unitCntrList []models.UnitConfCntr
	var caasConfList []models.CaasConf
	con := initial.DB
	if err := con.Table("unit_conf_cntr").Select([]string{`unit_id`, `caas_team`, `caas_cluster`, `caas_stack`, `service_name`, `deploy_comp`}).
		Where("is_delete=0 and service_name <> ''").Find(&unitCntrList).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if err := con.Table("conf_caas").Select([]string{"deploy_comp", "agent_ip", "agent_port"}).
		Where("is_delete = '0'").Find(&caasConfList).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	var caasConf = make(map[string]models.CaasConf, 8)
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
			if comp, ok := caasConf[unit.DeployComp]; !ok {
				continue
			} else {
				caasOpr := caas.CaasOpr{
					AgentConf:   comp,
					TeamId:      unit.CaasTeam,
					ClustUuid:   unit.CaasCluster,
					StackName:   unit.CaasStack,
					ServiceName: unit.ServiceName,
				}
				err, serviceConfig := caasOpr.RetryGetServiceConfig(5)
				if err != nil {
					beego.Error(err)
					continue
				}
				if err := saveConfig(unit.UnitId, serviceConfig); err != nil {
					beego.Error(err)
				}
			}

		}
	}()
	c.SetJson(1, "", "同步配置任务启动，请等待！")
}

func saveConfig(unitId int, data *caas.ServiceConfigAll) error {
	healthCheck, _ := json.Marshal(data.HealthCheck)
	envMap, _ := json.Marshal(data.Environment)
	logConfig, _ := json.Marshal(data.LogConfig)
	volume, _ := json.Marshal(data.Volume)
	scheduler, _ := json.Marshal(data.Scheduler)
	item := models.ConfCaasService{
		UnitId:            unitId,
		InstanceNum:       data.Scaling.DefaultInstances,
		CpuLimit:          data.CpuLimit,
		MemLimit:          data.MemLimit,
		IsAlwaysPullImage: data.AlwaysPullImage,
		Image:             data.Image,
		LogConfig:         string(logConfig),
		Volume:            string(volume),
		Scheduler:         string(scheduler),
		Env:               string(envMap),
		HealthCheck:       string(healthCheck),
		SyncTime:          time.Now().Format(initial.DatetimeFormat),
	}
	return models.ConfCaasService{}.UpdateOrCreate(item)
}
