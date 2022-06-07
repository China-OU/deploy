package info

import (
	"fmt"
	"initial"
	"library/common"
	"controllers/operation"
	"library/mcp"
	"time"
	"strings"
	"github.com/astaxie/beego"
)

// @Description 多容器平台之rancher平台，获取租户项目列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	search	query	string	false	"模糊查询"
// @Success 200 {object} []mcp.RancherData
// @Failure 403
// @router /rancher/project [get]
func (c *McpContainerDataController) RancherProjectList() {
	deploy_comp := c.GetString("deploy_comp")
	search := c.GetString("search")
	if deploy_comp == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}
	key := fmt.Sprintf("rancher_%s_project_list", deploy_comp)
	if !initial.GetCache.IsExist(key) || common.GetString(initial.GetCache.Get(key)) == "" {
		err, cass_config := operation.GetCaasConfig(deploy_comp)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
			return
		}
		rancher := mcp.McpRancherOpr {
			AgentConf: cass_config,
		}
		err, prj_list := rancher.GetRancherProjectList()
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
		if len(prj_list) == 0 {
			c.SetJson(0, "", "没有获取到数据，请联系开发人员定位原因！")
			return
		}
		initial.GetCache.Put(key, prj_list, 12*time.Hour)
	}
	origin, ok := (initial.GetCache.Get(key)).([]mcp.RancherData)
	if !ok {
		c.SetJson(0, "", "数据转换错误!")
		return
	}
	var ret_data []mcp.RancherData
	for _, v := range origin {
		if strings.Contains(v.Name, search) {
			ret_data = append(ret_data, v)
		}
	}
	if len(ret_data) > 10 {
		ret_data = ret_data[0:10]
	}
	c.SetJson(1, ret_data, "rancher的project获取成功！")
}

// @Description 多容器平台之rancher平台，获取租户堆栈列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	project_id	query	string	true	"项目id"
// @Param	search	query	string	false	"模糊查询"
// @Success 200 {object} []mcp.RancherData
// @Failure 403
// @router /rancher/stacklist [get]
func (c *McpContainerDataController) RancherStackList() {
	deploy_comp := c.GetString("deploy_comp")
	project_id := c.GetString("project_id")
	search := c.GetString("search")
	if deploy_comp == "" || project_id == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}
	key := fmt.Sprintf("rancher_%s_%s_stack_list", deploy_comp, project_id)
	if !initial.GetCache.IsExist(key) || common.GetString(initial.GetCache.Get(key)) == "" {
		err, cass_config := operation.GetCaasConfig(deploy_comp)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
			return
		}
		rancher := mcp.McpRancherOpr {
			AgentConf: cass_config,
			ProjectId: project_id,
		}
		err, stack_list := rancher.GetRancherStackList()
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
		if len(stack_list) == 0 {
			c.SetJson(0, "", "没有获取到数据，请联系开发人员定位原因！")
			return
		}
		initial.GetCache.Put(key, stack_list, 12*time.Hour)
	}
	origin, ok := (initial.GetCache.Get(key)).([]mcp.RancherData)
	if !ok {
		c.SetJson(0, "", "数据转换错误!")
		return
	}
	var ret_data []mcp.RancherData
	for _, v := range origin {
		if strings.Contains(v.Name, search) {
			ret_data = append(ret_data, v)
		}
	}
	if len(ret_data) > 10 {
		ret_data = ret_data[0:10]
	}
	c.SetJson(1, ret_data, "rancher的堆栈获取成功！")
}

// @Description 多容器平台之rancher平台，获取租户服务列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	project_id	query	string	true	"项目id"
// @Param	stack_id	query	string	true	"堆栈id"
// @Param	search	query	string	false	"模糊查询，只支持首字母查询"
// @Success 200 {object} []mcp.RancherData
// @Failure 403
// @router /rancher/servicelist [get]
func (c *McpContainerDataController) RancherServiceList() {
	deploy_comp := c.GetString("deploy_comp")
	project_id := c.GetString("project_id")
	stack_id := c.GetString("stack_id")
	search := c.GetString("search")
	if deploy_comp == "" || project_id == "" || stack_id == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}

	err, cass_config := operation.GetCaasConfig(deploy_comp)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
		return
	}
	rancher := mcp.McpRancherOpr {
		AgentConf: cass_config,
		ProjectId: project_id,
		StackId: stack_id,
		Search: search,
	}
	err, service_list := rancher.GetRancherServiceList()
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, service_list, "rancher的服务列表获取成功！")
}
