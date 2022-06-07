package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/report:ReportController"] = append(beego.GlobalControllerRouter["controllers/report:ReportController"],
		beego.ControllerComments{
			Method: "VmDeploy",
			Router: `/vm-deploy`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/report:ReportController"] = append(beego.GlobalControllerRouter["controllers/report:ReportController"],
		beego.ControllerComments{
			Method: "CntrDeploy",
			Router: `/cntr-deploy`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/report:ReportController"] = append(beego.GlobalControllerRouter["controllers/report:ReportController"],
		beego.ControllerComments{
			Method: "CntrUp",
			Router: `/cntr-up`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/report:ReportController"] = append(beego.GlobalControllerRouter["controllers/report:ReportController"],
		beego.ControllerComments{
			Method: "GetIndex",
			Router: `/index`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

}
