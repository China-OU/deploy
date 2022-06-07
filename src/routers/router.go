// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"controllers/login"
	"controllers/operation"
	"controllers/operation/caas_cntr"
	"controllers/operation/caas_route"

	"controllers/caas_conf"
	"controllers/ext"
	"controllers/harbor"
	"controllers/info"
	"controllers/itest"
	"controllers/master"
	"controllers/online"
	"controllers/report"
	"controllers/unit_conf"
	"controllers/user_role"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"time"
)

func init() {
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins: true,
		AllowCredentials: true,
		AllowOrigins:    []string{"*"},
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:   []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		MaxAge:          5 * time.Minute,
	}))

	ns := beego.NewNamespace("/mdeploy/v1",
		// cmtoken登录
		beego.NSNamespace("/login",
			beego.NSInclude(
				&login.LoginController{},
				&login.LoginCheckController{},
				&login.LogoutController{},
			),
		),
		// nuc 双因子登录
		beego.NSNamespace("/nuc-login",
			beego.NSInclude(
				&login.NucLoginController{},
			),
		),
		// 配置
		beego.NSNamespace("/conf",
			beego.NSInclude(
				&unit_conf.UnitConfListController{},
				&unit_conf.DBConfListController{},
				&unit_conf.StdCntrConfController{},
				&unit_conf.StdVmConfController{},
				&unit_conf.NoStdVmConfController{},
				&caas_conf.CaasDataListController{},
				&unit_conf.MultiContainerConfController{},
			),
		),
		// 生产发布
		beego.NSNamespace("/online",
			beego.NSInclude(
				&online.DBOnlineController{},
				&online.StdCntrOnlineController{},
				&online.StdVmOnlineController{},
				&online.NvmOnlineController{},
				&online.ReleaseRecordController{},
			),
		),
		// 操作，包括初始化，重启，镜像更新等
		beego.NSNamespace("/opr",
			beego.NSInclude(
				&operation.CntrOprController{},
				&operation.McpOprController{},
				&operation.HostInitController{},
				&caas_route.CaaSRoute{},
				&caas_cntr.CntrController{},
				&harbor.HarborOprController{},
				&operation.VMOprController{},
			),
		),
		// 用于agent调用，从master节点获取配置、推送数据给master保存等功能
		beego.NSNamespace("/master",
			beego.NSInclude(
				&master.CaasBaseConfController{},
			),
		),
		// 管理操作，包括agent配置、caas信息同步，agent分发；基础配置管理；用户权限管理；
		beego.NSNamespace("/manage",
			beego.NSInclude(
				&caas_conf.ManageCaasAgentController{},
				&caas_conf.McpAgentConfController{},
				&operation.ManageDBInfoController{},
				&user_role.RoleController{},
			),
		),

		// 报表(含首页)
		beego.NSNamespace("/report",
			beego.NSInclude(
				&report.ReportController{},
			),
		),

		// 通用api，比如人员信息拉取、git信息拉取等
		beego.NSNamespace("/common",
			beego.NSInclude(
				&info.UnitTypeInfoController{},
				&info.UnitDetailController{},
				&info.CaasDetailController{},
				&info.UserInfoController{},
				&info.McpIstioDataController{},
				&info.McpContainerDataController{},
				//&ws.WsController{},
			),
		),
		beego.NSNamespace("/git",
			beego.NSInclude(
				&info.GitInfoController{},
			),
		),
		// 外部调用接口，用于部署平台之间的调用，devops的调用和cpds的调用
		beego.NSNamespace("/ext",
			beego.NSInclude(
				&ext.MultiEnvConnController{},
				&ext.ExtCntrDeployController{},
				&ext.ExtCpdsFuncController{},
				&ext.ExtPmsFuncController{},
				&ext.McpInfoController{},
				&ext.DeployInfoController{},
				&ext.StandCoverController{},
			),
		),









		beego.NSNamespace("/test",
			beego.NSInclude(
				&itest.FuncTestController{},
				&itest.ClearLogUserController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
