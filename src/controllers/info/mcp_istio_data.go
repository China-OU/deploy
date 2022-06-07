package info

import (
	"controllers"
	"initial"
	"github.com/astaxie/beego"
	"fmt"
	"library/mcp"
	"library/common"
	"time"
	"strings"
	"controllers/operation"
)

type McpIstioDataController struct {
	controllers.BaseController
}

func (c *McpIstioDataController) URLMapping() {
	c.Mapping("IstioNamespace", c.IstioNamespace)
	c.Mapping("IstioDeployment", c.IstioDeployment)
}

// @Title IstioNamespace
// @Description 获取istio的命名空间
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	search	query	string	true	"模糊查询"
// @Success 200 {object} []mcp.NamespaceMetadata
// @Failure 403
// @router /istio/namespace [get]
func (c *McpIstioDataController) IstioNamespace() {
	deploy_comp := c.GetString("deploy_comp")
	search := c.GetString("search")
	if deploy_comp == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}
	key := fmt.Sprintf("istio_%s_namespace", deploy_comp)
	if !initial.GetCache.IsExist(key) || common.GetString(initial.GetCache.Get(key)) == "" {
		err, cass_config := operation.GetCaasConfig(deploy_comp)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
			return
		}
		istio := mcp.McpIstioOpr {
			AgentConf: cass_config,
		}
		err, namespace := istio.GetIstioNamespace()
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
		var ret []mcp.NamespaceMetadata
		for _, v := range namespace.Items {
			if v.Status.Phase == "Active" {
				ret = append(ret, v.Metadata)
			}
		}
		if len(ret) == 0 {
			c.SetJson(0, "", "没有获取到数据，请联系开发人员定位原因！")
			return
		}
		initial.GetCache.Put(key, ret, 1*time.Hour)
	}
	origin, ok := (initial.GetCache.Get(key)).([]mcp.NamespaceMetadata)
	if !ok {
		c.SetJson(0, "", "数据转换错误!")
		return
	}
	var ret_data []mcp.NamespaceMetadata
	for _, v := range origin {
		if strings.Contains(v.Name, search) {
			ret_data = append(ret_data, v)
		}
	}
	if len(ret_data) > 10 {
		ret_data = ret_data[0:10]
	}
	c.SetJson(1, ret_data, "istio的namespace获取成功！")
}

// @Title IstioDeployment
// @Description 获取命名空间下的应用列表
// @Param	deploy_comp	query	string	true	"部署租户"
// @Param	namespace	query	string	true	"命名空间"
// @Param	search	query	string	true	"模糊查询"
// @Success 200 {object} []models.CaasCluster
// @Failure 403
// @router /istio/deployment [get]
func (c *McpIstioDataController) IstioDeployment() {
	deploy_comp := c.GetString("deploy_comp")
	namespace := c.GetString("namespace")
	search := c.GetString("search")
	if deploy_comp == "" || namespace == "" {
		c.SetJson(0, "", "参数有空值！")
		return
	}
	err, cass_config := operation.GetCaasConfig(deploy_comp)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error() + "，请检查agent是否配置！")
		return
	}
	istio := mcp.McpIstioOpr {
		AgentConf: cass_config,
		Namespace: namespace,
	}
	err, deployment := istio.GetIstioDeployment()
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	var ret_data []mcp.DeploymentRet
	for _, v := range deployment {
		if strings.Contains(v.Name, search) {
			ret_data = append(ret_data, v)
		}
	}
	if len(ret_data) > 10 {
		ret_data = ret_data[0:10]
	}
	c.SetJson(1, ret_data, "istio的deployment获取成功！")
}