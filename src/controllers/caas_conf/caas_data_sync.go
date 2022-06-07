package caas_conf

import (
	"time"
	"library/common"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"models"
	"encoding/json"
	"initial"
	"fmt"
	"library/caas"
	"strings"
	"library/datasession"
)

// AllData 获取caas全部数据
// @Title 获取caas全部数据
// @Description 从agent获取各caas平台的数据，记录到数据库中，会实时访问caas平台的接口，通过agent去代理访问
// @Param	comp	query	string	true	"租户，当为all时，为全部租户的数据"
// @Success 200 true or false
// @Failure 403
// @router /caas/syncdata [post]
func (c *ManageCaasAgentController) CaasSyncData() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	comp := c.GetString("comp")
	if comp == "all" {
		last_time, flag := datasession.CaasServiceSyncTime()
		if time.Now().Add(- 30 * time.Minute).Format(initial.DatetimeFormat) < common.GetString(last_time) && flag == 1 {
			c.SetJson(0, "", "Caas服务列表30分钟内只能同步一次，上次同步时间：" + common.GetString(last_time))
			return
		}
	} else {
		last_time, flag := datasession.CaasSingleSyncTime()
		if time.Now().Add(- 5 * time.Minute).Format(initial.DatetimeFormat) < common.GetString(last_time) && flag == 1 {
			c.SetJson(0, "", "Caas单租户服务列表5分钟内只能同步一次，上次同步时间：" + common.GetString(last_time))
			return
		}
	}

	// 获取agent列表
	var ca_conf []models.CaasConf
	cond := ""
	if comp == "all" {
		cond = "is_delete=0"
	} else {
		cond = fmt.Sprintf("is_delete=0 and deploy_comp = '%s' ", comp)
	}
	err := initial.DB.Model(models.CaasConf{}).Where(cond).Find(&ca_conf).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	for _, v := range ca_conf {
		// 不同的网络区域，不同的agent，可以并发，后续再处理
		caas_id := v.Id
		// 获取team的列表
		team_list, err := GetCaasTeamList(v.AgentIp, v.AgentPort)
		if err != nil {
			beego.Error(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}

		// 获取cluster的列表
		clust_list, err := GetCaasClustList(v.AgentIp, v.AgentPort)
		if err != nil {
			beego.Error(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}

		// 获取stack列表
		for _, i := range team_list {
			// 去掉压测数据
			if i.Name == "pressbasedata" || strings.Contains(i.Name, "_INNER") == true {
				continue
			}
			for _, j := range clust_list {
				stack_list, err := GetCaasStackList(common.GetString(i.Id), j.Uuid, v.AgentIp, v.AgentPort)
				if err != nil {
					beego.Error(err.Error())
					c.SetJson(0, "", err.Error())
					return
				}
				for _, k := range stack_list {
					// 录入单元
					service_list, err := GetCaasServiceList(common.GetString(i.Id), j.Uuid, k.Name, v.AgentIp, v.AgentPort)
					if err != nil {
						beego.Error(err.Error())
						c.SetJson(0, "", err.Error())
						return
					}
					for _, t := range service_list {
						err := InsertOrUpdateCaasDetail(caas_id, i, j, k, t)
						if err != nil {
							beego.Error(err.Error())
							c.SetJson(0, "", err.Error())
							return
						}
						time.Sleep(10 * time.Nanosecond)
					}
				}
			}
		}

		// 配置表同步更新时间
		tx := initial.DB.Begin()
		err = tx.Model(models.CaasConf{}).Where("id=?", v.Id).Update("detail_sync_time",
			time.Now().Format(initial.DatetimeFormat)).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
		tx.Commit()

	}
	c.SetJson(1, "", "caas服务列表数据同步成功！")
}

func GetCaasTeamList(ip, port string) ([]caas.TeamDataDetail, error) {
	team_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/info/teamlist", ip, port)
	req_team := httplib.Get(team_url)
	req_team.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req_team.Param("ip_list", ip_list)
	ret_team, err := req_team.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	
	type TeamRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data []caas.TeamDataDetail `json:"data"`
	}
	var team_ret TeamRet
	err = json.Unmarshal(ret_team, &team_ret)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return team_ret.Data, nil
}

func GetCaasClustList(ip, port string) ([]caas.ClustData, error) {
	clust_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/info/clustlist", ip, port)
	req_clust := httplib.Get(clust_url)
	req_clust.Header("agent-auth", initial.AgentToken)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req_clust.Param("ip_list", ip_list)
	ret_clust, err := req_clust.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	type ClustRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data []caas.ClustData `json:"data"`
	}
	var clust_ret ClustRet
	err = json.Unmarshal(ret_clust, &clust_ret)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return clust_ret.Data, nil
}

func GetCaasStackList(team_id, clust_uuid, ip, port string) ([]caas.StackDataDetail, error) {
	stack_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/info/stacklist", ip, port)
	req_stack := httplib.Get(stack_url)
	req_stack.Header("agent-auth", initial.AgentToken)
	req_stack.Param("team_id", team_id)
	req_stack.Param("clust_uuid", clust_uuid)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req_stack.Param("ip_list", ip_list)
	ret_stack, err := req_stack.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	type StackRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data []caas.StackDataDetail `json:"data"`
	}
	var stack_ret StackRet
	err = json.Unmarshal(ret_stack, &stack_ret)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return stack_ret.Data, nil
}

func GetCaasServiceList(team_id, clust_uuid, stack_name, ip, port string) ([]caas.ServiceDataDetail, error) {
	service_url := "http://" + fmt.Sprintf("%s:%s/agent/v1/info/servicelist", ip, port)
	req_service := httplib.Get(service_url)
	req_service.SetTimeout(30 * time.Minute, 30 * time.Minute)
	req_service.Header("agent-auth", initial.AgentToken)
	req_service.Param("team_id", team_id)
	req_service.Param("clust_uuid", clust_uuid)
	req_service.Param("stack_name", stack_name)
	ip_list := strings.Join(common.GetLocalIp(), ",")
	req_service.Param("ip_list", ip_list)
	ret_service, err := req_service.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	type ServiceRet struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data []caas.ServiceDataDetail `json:"data"`
	}
	var service_ret ServiceRet
	err = json.Unmarshal(ret_service, &service_ret)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return service_ret.Data, nil
}

func InsertOrUpdateCaasDetail(caas_id int, team caas.TeamDataDetail, clust caas.ClustData, stack caas.StackDataDetail,
	service caas.ServiceDataDetail) error {
		tx := initial.DB.Begin()
		var cnt int
		var detail models.CaasConfDetail
		tx.Model(models.CaasConfDetail{}).Where("caas_id=? and team_id=? and cluster_uuid=? and stack_name=? and service_name=?",
			caas_id, team.Id, clust.Uuid, stack.Name, service.Name).Count(&cnt).First(&detail)
		var d models.CaasConfDetail
		d.CaasId = caas_id
		d.TeamId = common.GetString(team.Id)
		d.TeamName = team.Name
		d.TeamDesc = team.Description
		d.ClusterId = common.GetString(clust.Id)
		d.ClusterUuid = clust.Uuid
		d.ClusterName = clust.Name
		d.ClusterDesc = clust.Des
		d.StackId = stack.Id
		d.StackName = stack.Name
		d.StackUuid = stack.Uuid
		d.StackDesc = stack.Description
		d.ServiceId = service.Id
		d.ServiceName = service.Name
		d.ServiceUuid = service.Uuid
		// 后续才会有这个数据
		d.ServiceNum = 0

		if cnt > 0 {
			// 更新
			d.Id = detail.Id
			d.InsertTime = detail.InsertTime
			d.IsDelete = 0
			err := tx.Save(&d).Error
			if err != nil {
				tx.Rollback()
				beego.Error(err.Error())
				return err
			}
		} else {
			// 录入
			d.InsertTime = time.Now().Format(initial.DatetimeFormat)
			d.IsDelete = 0
			err := tx.Create(&d).Error
			if err != nil {
				tx.Rollback()
				beego.Error(err.Error())
				return err
			}
		}
		tx.Commit()
		return nil
}
