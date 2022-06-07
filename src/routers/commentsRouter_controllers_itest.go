package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/itest:ClearLogUserController"] = append(beego.GlobalControllerRouter["controllers/itest:ClearLogUserController"],
		beego.ControllerComments{
			Method: "ClearUser",
			Router: `/clear/user`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/itest:FuncTestController"] = append(beego.GlobalControllerRouter["controllers/itest:FuncTestController"],
		beego.ControllerComments{
			Method: "Get",
			Router: `/func`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

}
