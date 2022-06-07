package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "RowSqlExec",
			Router: `/db/row/exec`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "RowSqlInfo",
			Router: `/db/row/info`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "RowSqlDel",
			Router: `/db/row/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBPullDir",
			Router: `/db/pulldir`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBDeploy",
			Router: `/db/deploy`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBFreshSha",
			Router: `/db/fresh/sha`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBFreshResult",
			Router: `/db/fresh/result`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBList",
			Router: `/db/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBSave",
			Router: `/db/save`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBDel",
			Router: `/db/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBDetail",
			Router: `/db/detail/:id`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:DBOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:DBOnlineController"],
		beego.ControllerComments{
			Method: "DBResultQuery",
			Router: `/db/result/query`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:NvmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:NvmOnlineController"],
		beego.ControllerComments{
			Method: "NvmDeploy",
			Router: `/nvm/deploy`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:NvmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:NvmOnlineController"],
		beego.ControllerComments{
			Method: "NvmShellLog",
			Router: `/nvm/shell-log`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:NvmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:NvmOnlineController"],
		beego.ControllerComments{
			Method: "NvmOnlineList",
			Router: `/nvm/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:NvmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:NvmOnlineController"],
		beego.ControllerComments{
			Method: "NvmOnlineSave",
			Router: `/nvm/save`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:NvmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:NvmOnlineController"],
		beego.ControllerComments{
			Method: "NvmOnlineDelete",
			Router: `/nvm/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:NvmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:NvmOnlineController"],
		beego.ControllerComments{
			Method: "NvmResultQuery",
			Router: `/nvm/result/query`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:ReleaseRecordController"] = append(beego.GlobalControllerRouter["controllers/online:ReleaseRecordController"],
		beego.ControllerComments{
			Method: "QueryReleaseRecord",
			Router: `/record/query`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:ReleaseRecordController"] = append(beego.GlobalControllerRouter["controllers/online:ReleaseRecordController"],
		beego.ControllerComments{
			Method: "PmsReleaseRecord",
			Router: `/record/pms`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"],
		beego.ControllerComments{
			Method: "CntrList",
			Router: `/cntr/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"],
		beego.ControllerComments{
			Method: "CntrSave",
			Router: `/cntr/save`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"],
		beego.ControllerComments{
			Method: "CntrDel",
			Router: `/cntr/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"],
		beego.ControllerComments{
			Method: "CntrResultQuery",
			Router: `/cntr/result/query`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"],
		beego.ControllerComments{
			Method: "CntrBuild",
			Router: `/cntr/build`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"],
		beego.ControllerComments{
			Method: "CntrJenkLog",
			Router: `/cntr/jenk-log`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdCntrOnlineController"],
		beego.ControllerComments{
			Method: "CntrUpgrade",
			Router: `/cntr/upgrade`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"],
		beego.ControllerComments{
			Method: "GetVMOnlineList",
			Router: `/vm/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"],
		beego.ControllerComments{
			Method: "AddVMOnlineTask",
			Router: `/vm/save`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"],
		beego.ControllerComments{
			Method: "DelVMOnlineTask",
			Router: `/vm/del/:task_id`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"],
		beego.ControllerComments{
			Method: "VmResultQuery",
			Router: `/vm/result/query`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"],
		beego.ControllerComments{
			Method: "VmBuild",
			Router: `/vm/build`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"],
		beego.ControllerComments{
			Method: "VmJenkinsLog",
			Router: `/vm/jenk-log`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"],
		beego.ControllerComments{
			Method: "VmUpgrade",
			Router: `/vm/upgrade`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"] = append(beego.GlobalControllerRouter["controllers/online:StdVmOnlineController"],
		beego.ControllerComments{
			Method: "VMLog",
			Router: `/vm/log/:task_id`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

}
