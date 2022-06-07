package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/index:IndexController"] = append(beego.GlobalControllerRouter["controllers/index:IndexController"],
		beego.ControllerComments{
			Method: "Get",
			Router: `/get`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

}
