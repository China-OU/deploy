package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "CntrConfigUpdate",
			Router: `/cntr/config`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "CntrConfigGet",
			Router: `/cntr/config`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "CntrConfigSync",
			Router: `/cntr/config/sync`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "ReInitService",
			Router: `/cntr/init`,
			AllowHTTPMethods: []string{"put"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "InitService",
			Router: `/cntr/init`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "ListCntrInit",
			Router: `/cntr/init/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "GetCntrInit",
			Router: `/cntr/init`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "ListAppUnit",
			Router: `/cntr/init/unit-list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "DeleteInitRecord",
			Router: `/cntr/init`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "ListOprLog",
			Router: `/cntr/init/opr-log`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "SyncCntrInitStatus",
			Router: `/cntr/init/:id/sync`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"] = append(beego.GlobalControllerRouter["controllers/operation/caas_cntr:CntrController"],
		beego.ControllerComments{
			Method: "CntrConfigScan",
			Router: `/cntr/config/scan`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

}
