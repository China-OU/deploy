package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/harbor:HarborOprController"] = append(beego.GlobalControllerRouter["controllers/harbor:HarborOprController"],
		beego.ControllerComments{
			Method: "ImageSync",
			Router: `/image/sync`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/harbor:HarborOprController"] = append(beego.GlobalControllerRouter["controllers/harbor:HarborOprController"],
		beego.ControllerComments{
			Method: "ImageList",
			Router: `/image/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/harbor:HarborOprController"] = append(beego.GlobalControllerRouter["controllers/harbor:HarborOprController"],
		beego.ControllerComments{
			Method: "ImageAdd",
			Router: `/image/add`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/harbor:HarborOprController"] = append(beego.GlobalControllerRouter["controllers/harbor:HarborOprController"],
		beego.ControllerComments{
			Method: "ImageDel",
			Router: `/image/del`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

}
