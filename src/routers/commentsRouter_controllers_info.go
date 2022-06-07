package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/info:CaasDetailController"] = append(beego.GlobalControllerRouter["controllers/info:CaasDetailController"],
		beego.ControllerComments{
			Method: "GetTeamList",
			Router: `/caas/teamlist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:CaasDetailController"] = append(beego.GlobalControllerRouter["controllers/info:CaasDetailController"],
		beego.ControllerComments{
			Method: "GetClustList",
			Router: `/caas/clustlist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:CaasDetailController"] = append(beego.GlobalControllerRouter["controllers/info:CaasDetailController"],
		beego.ControllerComments{
			Method: "GetStackList",
			Router: `/caas/stacklist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:CaasDetailController"] = append(beego.GlobalControllerRouter["controllers/info:CaasDetailController"],
		beego.ControllerComments{
			Method: "GetServiceList",
			Router: `/caas/servicelist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:GitInfoController"] = append(beego.GlobalControllerRouter["controllers/info:GitInfoController"],
		beego.ControllerComments{
			Method: "SearchInfo",
			Router: `/search`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:GitInfoController"] = append(beego.GlobalControllerRouter["controllers/info:GitInfoController"],
		beego.ControllerComments{
			Method: "GetBranchList",
			Router: `/branch/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpContainerDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpContainerDataController"],
		beego.ControllerComments{
			Method: "RancherProjectList",
			Router: `/rancher/project`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpContainerDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpContainerDataController"],
		beego.ControllerComments{
			Method: "RancherStackList",
			Router: `/rancher/stacklist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpContainerDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpContainerDataController"],
		beego.ControllerComments{
			Method: "RancherServiceList",
			Router: `/rancher/servicelist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpContainerDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpContainerDataController"],
		beego.ControllerComments{
			Method: "CaasTeamList",
			Router: `/caas-new/teamlist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpContainerDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpContainerDataController"],
		beego.ControllerComments{
			Method: "CaasClustList",
			Router: `/caas-new/clustlist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpContainerDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpContainerDataController"],
		beego.ControllerComments{
			Method: "CaasStackList",
			Router: `/caas-new/stacklist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpContainerDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpContainerDataController"],
		beego.ControllerComments{
			Method: "CaasServiceList",
			Router: `/caas-new/servicelist`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpIstioDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpIstioDataController"],
		beego.ControllerComments{
			Method: "IstioNamespace",
			Router: `/istio/namespace`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:McpIstioDataController"] = append(beego.GlobalControllerRouter["controllers/info:McpIstioDataController"],
		beego.ControllerComments{
			Method: "IstioDeployment",
			Router: `/istio/deployment`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:UnitDetailController"] = append(beego.GlobalControllerRouter["controllers/info:UnitDetailController"],
		beego.ControllerComments{
			Method: "UnitSearch",
			Router: `/unit/search`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:UnitDetailController"] = append(beego.GlobalControllerRouter["controllers/info:UnitDetailController"],
		beego.ControllerComments{
			Method: "UnitDeployComp",
			Router: `/unit/dcomp`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:UnitTypeInfoController"] = append(beego.GlobalControllerRouter["controllers/info:UnitTypeInfoController"],
		beego.ControllerComments{
			Method: "GetType",
			Router: `/unit/type`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:UnitTypeInfoController"] = append(beego.GlobalControllerRouter["controllers/info:UnitTypeInfoController"],
		beego.ControllerComments{
			Method: "GetSubType",
			Router: `/unit/subtype`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:UserInfoController"] = append(beego.GlobalControllerRouter["controllers/info:UserInfoController"],
		beego.ControllerComments{
			Method: "SyncUserFromPms",
			Router: `/user/sync`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/info:UserInfoController"] = append(beego.GlobalControllerRouter["controllers/info:UserInfoController"],
		beego.ControllerComments{
			Method: "SearchUser",
			Router: `/user/search`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

}
