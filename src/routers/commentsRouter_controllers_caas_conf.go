package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/caas_conf:CaasDataListController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:CaasDataListController"],
		beego.ControllerComments{
			Method: "GetAll",
			Router: `/caas/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"],
		beego.ControllerComments{
			Method: "CaasAgentList",
			Router: `/caas/agent/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"],
		beego.ControllerComments{
			Method: "CaasAgentCreate",
			Router: `/caas/agent/create`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"],
		beego.ControllerComments{
			Method: "CaasAgentDel",
			Router: `/caas/agent/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"],
		beego.ControllerComments{
			Method: "CaasAgentSurv",
			Router: `/caas/agent/surv`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"],
		beego.ControllerComments{
			Method: "CaasAgentCheck",
			Router: `/caas/agent/check`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"],
		beego.ControllerComments{
			Method: "CaasAgentSearch",
			Router: `/caas/agent/search`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:ManageCaasAgentController"],
		beego.ControllerComments{
			Method: "CaasSyncData",
			Router: `/caas/syncdata`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:McpAgentConfController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:McpAgentConfController"],
		beego.ControllerComments{
			Method: "McpAgentList",
			Router: `/mcp/agent/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:McpAgentConfController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:McpAgentConfController"],
		beego.ControllerComments{
			Method: "McpAgentEdit",
			Router: `/mcp/agent/edit`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/caas_conf:McpAgentConfController"] = append(beego.GlobalControllerRouter["controllers/caas_conf:McpAgentConfController"],
		beego.ControllerComments{
			Method: "McpAgentDel",
			Router: `/mcp/agent/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

}
