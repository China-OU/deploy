package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "InitRoute",
			Router: `/route/int`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "GetRoute",
			Router: `/route/:id/detail`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "ListRoute",
			Router: `/route/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "ListRouteLog",
			Router: `/route/log/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "FetchRouteService",
			Router: `/route/sync-caas/:comp`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "EditRoute",
			Router: `/route/edit`,
			AllowHTTPMethods: []string{"put"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "DeleteRule",
			Router: `/route/rule/:id`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "DeleteTarget",
			Router: `/route/target/:id`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"] = append(beego.GlobalControllerRouter["controllers/operation/caas_route:CaaSRoute"],
		beego.ControllerComments{
			Method: "DeleteRoute",
			Router: `/route/:id`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

}
