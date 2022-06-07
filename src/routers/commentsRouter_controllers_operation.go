package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/operation:CntrOprController"] = append(beego.GlobalControllerRouter["controllers/operation:CntrOprController"],
		beego.ControllerComments{
			Method: "GetRecord",
			Router: `/cntr/record`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:CntrOprController"] = append(beego.GlobalControllerRouter["controllers/operation:CntrOprController"],
		beego.ControllerComments{
			Method: "GetRecordList",
			Router: `/cntr/record-list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:CntrOprController"] = append(beego.GlobalControllerRouter["controllers/operation:CntrOprController"],
		beego.ControllerComments{
			Method: "SearchList",
			Router: `/cntr/search`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:CntrOprController"] = append(beego.GlobalControllerRouter["controllers/operation:CntrOprController"],
		beego.ControllerComments{
			Method: "UpgradeService",
			Router: `/cntr/upgrade`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ExecCmdController"] = append(beego.GlobalControllerRouter["controllers/operation:ExecCmdController"],
		beego.ControllerComments{
			Method: "GetCmdList",
			Router: `/host/cmd`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ExecCmdController"] = append(beego.GlobalControllerRouter["controllers/operation:ExecCmdController"],
		beego.ControllerComments{
			Method: "PostAndRun",
			Router: `/host/run`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ExecCmdController"] = append(beego.GlobalControllerRouter["controllers/operation:ExecCmdController"],
		beego.ControllerComments{
			Method: "AddExecTask",
			Router: `/host/cmd`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ExecCmdController"] = append(beego.GlobalControllerRouter["controllers/operation:ExecCmdController"],
		beego.ControllerComments{
			Method: "ExecTask",
			Router: `/host/cmd/:task_id`,
			AllowHTTPMethods: []string{"put"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ExecCmdController"] = append(beego.GlobalControllerRouter["controllers/operation:ExecCmdController"],
		beego.ControllerComments{
			Method: "DeleteCmdTask",
			Router: `/host/cmd/:task_id`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:HostInitController"] = append(beego.GlobalControllerRouter["controllers/operation:HostInitController"],
		beego.ControllerComments{
			Method: "GetInitList",
			Router: `/host/init/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:HostInitController"] = append(beego.GlobalControllerRouter["controllers/operation:HostInitController"],
		beego.ControllerComments{
			Method: "AddInitTask",
			Router: `/host/init`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:HostInitController"] = append(beego.GlobalControllerRouter["controllers/operation:HostInitController"],
		beego.ControllerComments{
			Method: "ExecInitTask",
			Router: `/host/init/:task_id`,
			AllowHTTPMethods: []string{"put"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:HostInitController"] = append(beego.GlobalControllerRouter["controllers/operation:HostInitController"],
		beego.ControllerComments{
			Method: "DeleteInitTask",
			Router: `/host/init/:task_id`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ManageDBInfoController"] = append(beego.GlobalControllerRouter["controllers/operation:ManageDBInfoController"],
		beego.ControllerComments{
			Method: "GetDBAccounts",
			Router: `/db/account`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ManageDBInfoController"] = append(beego.GlobalControllerRouter["controllers/operation:ManageDBInfoController"],
		beego.ControllerComments{
			Method: "NewDBAccount",
			Router: `/db/account`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ManageDBInfoController"] = append(beego.GlobalControllerRouter["controllers/operation:ManageDBInfoController"],
		beego.ControllerComments{
			Method: "UpdateDbPassword",
			Router: `/db/account`,
			AllowHTTPMethods: []string{"put"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:ManageDBInfoController"] = append(beego.GlobalControllerRouter["controllers/operation:ManageDBInfoController"],
		beego.ControllerComments{
			Method: "DeleteDBPassword",
			Router: `/db/account/:account_id`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:McpOprController"] = append(beego.GlobalControllerRouter["controllers/operation:McpOprController"],
		beego.ControllerComments{
			Method: "ServiceDetail",
			Router: `/mcp/service/detail`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:McpOprController"] = append(beego.GlobalControllerRouter["controllers/operation:McpOprController"],
		beego.ControllerComments{
			Method: "ServiceUpgrade",
			Router: `/mcp/service/upgrade`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:McpOprController"] = append(beego.GlobalControllerRouter["controllers/operation:McpOprController"],
		beego.ControllerComments{
			Method: "McpRecord",
			Router: `/mcp/record`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:McpOprController"] = append(beego.GlobalControllerRouter["controllers/operation:McpOprController"],
		beego.ControllerComments{
			Method: "McpRecordList",
			Router: `/mcp/record-list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:VMOprController"] = append(beego.GlobalControllerRouter["controllers/operation:VMOprController"],
		beego.ControllerComments{
			Method: "VMAppStatus",
			Router: `/vm/state`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:VMOprController"] = append(beego.GlobalControllerRouter["controllers/operation:VMOprController"],
		beego.ControllerComments{
			Method: "VMAppUpgrade",
			Router: `/vm/upgrade`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:VMOprController"] = append(beego.GlobalControllerRouter["controllers/operation:VMOprController"],
		beego.ControllerComments{
			Method: "VMAppUpgradeHistory",
			Router: `/vm/upgrades`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation:VMOprController"] = append(beego.GlobalControllerRouter["controllers/operation:VMOprController"],
		beego.ControllerComments{
			Method: "VMAppRestart",
			Router: `/vm/restart`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

}
