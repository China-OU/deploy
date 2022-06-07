package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["controllers/login:LoginCheckController"] = append(beego.GlobalControllerRouter["controllers/login:LoginCheckController"],
		beego.ControllerComments{
			Method: "Post",
			Router: `/check`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/login:LoginController"] = append(beego.GlobalControllerRouter["controllers/login:LoginController"],
		beego.ControllerComments{
			Method: "Post",
			Router: `/`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/login:LogoutController"] = append(beego.GlobalControllerRouter["controllers/login:LogoutController"],
		beego.ControllerComments{
			Method: "Post",
			Router: `/logout`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/login:NucLoginController"] = append(beego.GlobalControllerRouter["controllers/login:NucLoginController"],
		beego.ControllerComments{
			Method: "DirectLogin",
			Router: `/direct/login`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/login:NucLoginController"] = append(beego.GlobalControllerRouter["controllers/login:NucLoginController"],
		beego.ControllerComments{
			Method: "VerifyCode",
			Router: `/verify/code`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/login:NucLoginController"] = append(beego.GlobalControllerRouter["controllers/login:NucLoginController"],
		beego.ControllerComments{
			Method: "TwoFactorLogin",
			Router: `/twofactor/login`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/login:NucLoginController"] = append(beego.GlobalControllerRouter["controllers/login:NucLoginController"],
		beego.ControllerComments{
			Method: "RefreshSMS",
			Router: `/refresh/sms`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/login:NucLoginController"] = append(beego.GlobalControllerRouter["controllers/login:NucLoginController"],
		beego.ControllerComments{
			Method: "ValidateSMS",
			Router: `/validate/sms`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["controllers/login:NucLoginController"] = append(beego.GlobalControllerRouter["controllers/login:NucLoginController"],
		beego.ControllerComments{
			Method: "LogOut",
			Router: `/logout`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

}
