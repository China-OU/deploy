package cfunc

import "github.com/astaxie/beego"

func GetJenkBaseUrl() string {
	env := beego.AppConfig.String("runmode")
	if env == "prd" || env == "dr" {
		return "http://100.65.169.39/jenkins/"
	} else {
		// "http://jenkins-di1.sit.cmrh.com:8080/jenkins/" 测试环境jenkins
		return "http://100.69.218.115:8080/jenkins/"
	}
}

func GetHarborUrl() string {
	env := beego.AppConfig.String("runmode")
	if env == "prd" {
		return "harbor.cmft.com"
	} else if env == "dr" {
		return "harbor-dr.cmft.com"
	} else {
		return "harbor.uat.cmft.com"
	}
}
