package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/master:CaasBaseConfController"] = append(beego.GlobalControllerRouter["controllers/master:CaasBaseConfController"],
		beego.ControllerComments{
			Method: "CaasConfList",
			Router: `/caas/conf`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/master:CaasBaseConfController"] = append(beego.GlobalControllerRouter["controllers/master:CaasBaseConfController"],
		beego.ControllerComments{
			Method: "DBConfList",
			Router: `/database/conf`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/master:CaasBaseConfController"] = append(beego.GlobalControllerRouter["controllers/master:CaasBaseConfController"],
		beego.ControllerComments{
			Method: "McpConfList",
			Router: `/mcp/conf`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

}
