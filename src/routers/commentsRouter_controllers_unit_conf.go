package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "ConnCheck",
			Router: `/db/conn/check`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "ChangePwd",
			Router: `/db/change/pwd`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "InvalidPkg",
			Router: `/db/invalid/pkg`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "CompilePkg",
			Router: `/db/compile/pkg`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "LockCheck",
			Router: `/db/lock/check`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "DBSave",
			Router: `/db/save`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "GetDBList",
			Router: `/db/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "DBDel",
			Router: `/db/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:DBConfListController"],
		beego.ControllerComments{
			Method: "GetDBPwd",
			Router: `/db/password`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:MultiContainerConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:MultiContainerConfController"],
		beego.ControllerComments{
			Method: "McpConfDel",
			Router: `/mcp/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:MultiContainerConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:MultiContainerConfController"],
		beego.ControllerComments{
			Method: "McpConfDetail",
			Router: `/mcp/detail`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:MultiContainerConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:MultiContainerConfController"],
		beego.ControllerComments{
			Method: "McpConfList",
			Router: `/mcp/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:MultiContainerConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:MultiContainerConfController"],
		beego.ControllerComments{
			Method: "McpConfEdit",
			Router: `/mcp/edit`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:NoStdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:NoStdVmConfController"],
		beego.ControllerComments{
			Method: "NvmConfList",
			Router: `/nvm/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:NoStdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:NoStdVmConfController"],
		beego.ControllerComments{
			Method: "NvmConfEdit",
			Router: `/nvm/edit`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:NoStdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:NoStdVmConfController"],
		beego.ControllerComments{
			Method: "NvmConfDel",
			Router: `/nvm/delete`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"],
		beego.ControllerComments{
			Method: "JenkXmlList",
			Router: `/cntr/jenk-xml`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"],
		beego.ControllerComments{
			Method: "JenkXmlEdit",
			Router: `/cntr/jenk-xml`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"],
		beego.ControllerComments{
			Method: "JenkXmlConfirm",
			Router: `/cntr/confirm`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"],
		beego.ControllerComments{
			Method: "CntrList",
			Router: `/cntr/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"],
		beego.ControllerComments{
			Method: "CntrDel",
			Router: `/cntr/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"],
		beego.ControllerComments{
			Method: "CpdsConfig",
			Router: `/cntr/cpds/config`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdCntrConfController"],
		beego.ControllerComments{
			Method: "CntrEdit",
			Router: `/cntr/edit`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"],
		beego.ControllerComments{
			Method: "New",
			Router: `/vm/new`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"],
		beego.ControllerComments{
			Method: "Update",
			Router: `/vm/update`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"],
		beego.ControllerComments{
			Method: "Delete",
			Router: `/vm/delete`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"],
		beego.ControllerComments{
			Method: "JenkinsXmlCheck",
			Router: `/vm/jenk/xml`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"],
		beego.ControllerComments{
			Method: "JenkinsXmlConfirm",
			Router: `/vm/jenk/confirm`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"],
		beego.ControllerComments{
			Method: "JenkinsXmlEdit",
			Router: `/vm/jenk/update`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:StdVmConfController"],
		beego.ControllerComments{
			Method: "GetAll",
			Router: `/vm/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"],
		beego.ControllerComments{
			Method: "Sync",
			Router: `/all/sync`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"],
		beego.ControllerComments{
			Method: "Add",
			Router: `/allunit/add`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"],
		beego.ControllerComments{
			Method: "Copy",
			Router: `/allunit/copy`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"],
		beego.ControllerComments{
			Method: "GetAll",
			Router: `/all`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"] = append(beego.GlobalControllerRouter["controllers/unit_conf:UnitConfListController"],
		beego.ControllerComments{
			Method: "Del",
			Router: `/allunit/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

}
