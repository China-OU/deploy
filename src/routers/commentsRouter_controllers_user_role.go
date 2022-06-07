package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/user_role:RoleController"] = append(beego.GlobalControllerRouter["controllers/user_role:RoleController"],
		beego.ControllerComments{
			Method: "GetUserPrivilege",
			Router: `/role/my-privilege`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/user_role:RoleController"] = append(beego.GlobalControllerRouter["controllers/user_role:RoleController"],
		beego.ControllerComments{
			Method: "GetAll",
			Router: `/role/list`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/user_role:RoleController"] = append(beego.GlobalControllerRouter["controllers/user_role:RoleController"],
		beego.ControllerComments{
			Method: "AddRole",
			Router: `/role/add`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/user_role:RoleController"] = append(beego.GlobalControllerRouter["controllers/user_role:RoleController"],
		beego.ControllerComments{
			Method: "OperationRole",
			Router: `/role/operation`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

}
