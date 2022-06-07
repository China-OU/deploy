package info

import (
	"controllers"
	"initial"
	"time"
	"library/mcp"
	"strings"
	"fmt"
	"library/common"
	"controllers/operation"
	"github.com/astaxie/beego"
)

type McpContainerDataController struct {
	controllers.BaseController
}

func (c *McpContainerDataController) URLMapping() {
	// caas容器平台
	c.Mapping("CaasTeamList", c.CaasTeamList)
	c.Mapping("CaasClustList", c.CaasClustList)
	c.Mapping("CaasStackList", c.CaasStackList)
	c.Mapping("CaasServiceList", c.CaasServiceList)
	// rancher容器平台
	c.Mapping("RancherProjectList", c.RancherProjectList)
	c.Mapping("RancherStackList", c.RancherStackList)
	c.Mapping("RancherServiceList", c.RancherServiceList)

	// ocp容器平台
}

// @Description 多容器平台之caas平台，获取租户团队列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	search	query	string	false	"模糊查询"
// @Success 200 {object} []mcp.TeamDataDetail
// @Failure 403
// @router /caas-new/teamlist [get]
func (c *McpContainerDataController) CaasTeamList() {
	deploy_comp := c.GetString("deploy_comp")
	search := c.GetString("search")
	if deploy_comp == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}
	key := fmt.Sprintf("caas_%s_teamlist", deploy_comp)
	if !initial.GetCache.IsExist(key) || common.GetString(initial.GetCache.Get(key)) == "" {
		err, cass_config := operation.GetCaasConfig(deploy_comp)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
			return
		}
		caas := mcp.McpCaasOpr {
			AgentConf: cass_config,
		}
		err, team_list := caas.GetCaasTeamList()
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
		if len(team_list) == 0 {
			c.SetJson(0, "", "没有获取到数据，请联系开发人员定位原因！")
			return
		}
		initial.GetCache.Put(key, team_list, 12*time.Hour)
	}
	origin, ok := (initial.GetCache.Get(key)).([]mcp.TeamDataDetail)
	if !ok {
		c.SetJson(0, "", "数据转换错误!")
		return
	}
	var ret_data []mcp.TeamDataDetail
	for _, v := range origin {
		if strings.Contains(v.Name, "INNER") {
			// 内部接口跳过
			continue
		}
		if strings.Contains(v.Name, search) {
			ret_data = append(ret_data, v)
		}
	}
	if len(ret_data) > 10 {
		ret_data = ret_data[0:10]
	}
	c.SetJson(1, ret_data, "caas的team获取成功！")
}

// @Description 多容器平台之caas平台，获取租户集群列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	search	query	string	false	"模糊查询"
// @Success 200 {object} []mcp.ClustData
// @Failure 403
// @router /caas-new/clustlist [get]
func (c *McpContainerDataController) CaasClustList() {
	deploy_comp := c.GetString("deploy_comp")
	search := c.GetString("search")
	if deploy_comp == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}
	key := fmt.Sprintf("caas_%s_clustlist", deploy_comp)
	if !initial.GetCache.IsExist(key) || common.GetString(initial.GetCache.Get(key)) == "" {
		err, cass_config := operation.GetCaasConfig(deploy_comp)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
			return
		}
		caas := mcp.McpCaasOpr {
			AgentConf: cass_config,
		}
		err, clust_list := caas.GetCaasClustList()
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
		if len(clust_list) == 0 {
			c.SetJson(0, "", "没有获取到数据，请联系开发人员定位原因！")
			return
		}
		initial.GetCache.Put(key, clust_list, 12*time.Hour)
	}
	origin, ok := (initial.GetCache.Get(key)).([]mcp.ClustData)
	if !ok {
		c.SetJson(0, "", "数据转换错误!")
		return
	}
	var ret_data []mcp.ClustData
	for _, v := range origin {
		if strings.Contains(v.Name, search) {
			ret_data = append(ret_data, v)
		}
	}
	if len(ret_data) > 10 {
		ret_data = ret_data[0:10]
	}
	c.SetJson(1, ret_data, "caas的集群获取成功！")
}

// @Description 多容器平台之caas平台，获取租户堆栈列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	team_id	query	string	true	"团队id"
// @Param	clust_uuid	query	string	true	"集群uuid"
// @Param	search	query	string	false	"模糊查询"
// @Success 200 {object} []mcp.StackDataDetail
// @Failure 403
// @router /caas-new/stacklist [get]
func (c *McpContainerDataController) CaasStackList() {
	deploy_comp := c.GetString("deploy_comp")
	team_id := c.GetString("team_id")
	clust_uuid := c.GetString("clust_uuid")
	search := c.GetString("search")
	if deploy_comp == "" || team_id == "" || clust_uuid == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}
	key := fmt.Sprintf("caas_%s_%s_%s_stacklist", deploy_comp, team_id, clust_uuid)
	if !initial.GetCache.IsExist(key) || common.GetString(initial.GetCache.Get(key)) == "" {
		err, cass_config := operation.GetCaasConfig(deploy_comp)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
			return
		}
		caas := mcp.McpCaasOpr {
			AgentConf: cass_config,
			TeamId: team_id,
			ClustUuid: clust_uuid,
		}
		err, stack_list := caas.GetCaasStackList()
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
		if len(stack_list) == 0 {
			c.SetJson(0, "", "没有获取到数据，请联系开发人员定位原因！")
			return
		}
		initial.GetCache.Put(key, stack_list, 10*time.Minute)
	}
	origin, ok := (initial.GetCache.Get(key)).([]mcp.StackDataDetail)
	if !ok {
		c.SetJson(0, "", "数据转换错误!")
		return
	}
	var ret_data []mcp.StackDataDetail
	for _, v := range origin {
		if strings.Contains(v.Name, search) {
			ret_data = append(ret_data, v)
		}
	}
	if len(ret_data) > 10 {
		ret_data = ret_data[0:10]
	}
	c.SetJson(1, ret_data, "caas的堆栈获取成功！")
}

// @Description 多容器平台之caas平台，获取租户服务列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	team_id	query	string	true	"团队id"
// @Param	clust_uuid	query	string	true	"集群uuid"
// @Param	stack_name	query	string	true	"堆栈名"
// @Param	search	query	string	false	"模糊查询"
// @Success 200 {object} []mcp.ServiceDataDetail
// @Failure 403
// @router /caas-new/servicelist [get]
func (c *McpContainerDataController) CaasServiceList() {
	deploy_comp := c.GetString("deploy_comp")
	team_id := c.GetString("team_id")
	clust_uuid := c.GetString("clust_uuid")
	stack_name := c.GetString("stack_name")
	search := c.GetString("search")
	if deploy_comp == "" || team_id == "" || clust_uuid == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}
	key := fmt.Sprintf("caas_%s_%s_%s_%s_servicelist", deploy_comp, team_id, clust_uuid, stack_name)
	if !initial.GetCache.IsExist(key) || common.GetString(initial.GetCache.Get(key)) == "" {
		err, cass_config := operation.GetCaasConfig(deploy_comp)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
			return
		}
		caas := mcp.McpCaasOpr {
			AgentConf: cass_config,
			TeamId: team_id,
			ClustUuid: clust_uuid,
			StackName: stack_name,
		}
		err, service_list := caas.GetCaasServiceList()
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
		if len(service_list) == 0 {
			c.SetJson(0, "", "没有获取到数据，请联系开发人员定位原因！")
			return
		}
		if deploy_comp == "CMRH" && (stack_name == "di1" || stack_name == "st1") {
			initial.GetCache.Put(key, service_list, 10*time.Hour)
		} else {
			initial.GetCache.Put(key, service_list, 5*time.Minute)
		}
	}
	origin, ok := (initial.GetCache.Get(key)).([]mcp.ServiceDataDetail)
	if !ok {
		c.SetJson(0, "", "数据转换错误!")
		return
	}
	var ret_data []mcp.ServiceDataDetail
	var pm mcp.ServiceDataDetail
	for _, v := range origin {
		if v.Name == search {
			pm = v
		} else {
			if strings.Contains(v.Name, search) {
				ret_data = append(ret_data, v)
			}
		}
	}
	// 全匹配放第一位
	var ret_arr []mcp.ServiceDataDetail
	if pm.Name != "" {
		ret_arr = append(ret_arr, pm)
	}
	if len(ret_data) > 10 {
		ret_arr = append(ret_arr, ret_data[0:10]...)
	} else {
		ret_arr = append(ret_arr, ret_data...)
	}
	c.SetJson(1, ret_arr, "caas的服务列表获取成功！")
}