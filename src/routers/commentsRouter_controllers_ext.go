package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/ext:DeployInfoController"] = append(beego.GlobalControllerRouter["controllers/ext:DeployInfoController"],
		beego.ControllerComments{
			Method: "GetUnitDeployConf",
			Router: `/unit/deployconf/search`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCntrDeployController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCntrDeployController"],
		beego.ControllerComments{
			Method: "CpdsUpgrade",
			Router: `/cntr/cpds/upgrade`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCntrDeployController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCntrDeployController"],
		beego.ControllerComments{
			Method: "Search",
			Router: `/cntr/service/search`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCntrDeployController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCntrDeployController"],
		beego.ControllerComments{
			Method: "DevopsUpgrade",
			Router: `/cntr/devops/upgrade`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCntrDeployController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCntrDeployController"],
		beego.ControllerComments{
			Method: "Poll",
			Router: `/cntr/service/poll`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"],
		beego.ControllerComments{
			Method: "CntrServiceUpgrade",
			Router: `/cpds/cntr-upgrade`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"],
		beego.ControllerComments{
			Method: "CntrServicePoll",
			Router: `/cpds/cntr-poll`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"],
		beego.ControllerComments{
			Method: "CntrServiceDetail",
			Router: `/cpds/cntr-detail`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"],
		beego.ControllerComments{
			Method: "HarborImageSync",
			Router: `/cpds/image/sync`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtCpdsFuncController"],
		beego.ControllerComments{
			Method: "HarborImageSyncPoll",
			Router: `/cpds/image-sync/poll`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:ExtPmsFuncController"] = append(beego.GlobalControllerRouter["controllers/ext:ExtPmsFuncController"],
		beego.ControllerComments{
			Method: "GetMcpConf",
			Router: `/pms/mcp/query`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:McpInfoController"] = append(beego.GlobalControllerRouter["controllers/ext:McpInfoController"],
		beego.ControllerComments{
			Method: "MultiContainerList",
			Router: `/cntr/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:MultiEnvConnController"] = append(beego.GlobalControllerRouter["controllers/ext:MultiEnvConnController"],
		beego.ControllerComments{
			Method: "UnitOnlineQuery",
			Router: `/online/query`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:MultiEnvConnController"] = append(beego.GlobalControllerRouter["controllers/ext:MultiEnvConnController"],
		beego.ControllerComments{
			Method: "DbConnCheck",
			Router: `/db/conn/check`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:MultiEnvConnController"] = append(beego.GlobalControllerRouter["controllers/ext:MultiEnvConnController"],
		beego.ControllerComments{
			Method: "Recieve",
			Router: `/unit/sync`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:MultiEnvConnController"] = append(beego.GlobalControllerRouter["controllers/ext:MultiEnvConnController"],
		beego.ControllerComments{
			Method: "HarborImgCheck",
			Router: `/harbor/check`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/ext:StandCoverController"] = append(beego.GlobalControllerRouter["controllers/ext:StandCoverController"],
		beego.ControllerComments{
			Method: "MdpStandUnit",
			Router: `/mdp/stand`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

}
