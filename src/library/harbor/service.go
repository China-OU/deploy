package harbor

import (
	"github.com/astaxie/beego"
	"strings"
	"errors"
	"library/common"
)

var harbor_user = "deployop"
var harbor_pwd = "ce81cb87a0092b4399dc9037bf8bd0d0"

func HarborCheckImage(image string) error {
	//镜像域名检查
	base_url := "https://harbor.uat.cmft.com"
	run_mode := beego.AppConfig.String("runmode")
	if run_mode == "prd" {
		base_url = "https://harbor.cmft.com"
		if strings.Contains(image, "harbor.cmft.com") == false {
			return errors.New("镜像域名有误，生产环境为harbor.uat.cmft.com！")
		}
	}
	if common.InList(run_mode, []string{"dev", "di", "st"}) {
		base_url = "https://harbor.uat.cmft.com"
		if strings.Contains(image, "harbor.uat.cmft.com") == false {
			return errors.New("镜像域名有误，测试环境为harbor.uat.cmft.com！")
		}
	}
	if run_mode == "dr" {
		base_url = "https://harbor-dr.cmft.com"
		if strings.Contains(image, "harbor-dr.cmft.com") == false {
			return errors.New("镜像域名有误，容灾环境为harbor-dr.cmft.com！")
		}
	}

	image = strings.Replace(image, "harbor.uat.cmft.com/", "", -1)
	image = strings.Replace(image, "harbor.cmft.com/", "", -1)
	image = strings.Replace(image, "harbor-dr.cmft.com/", "", -1)
	image_arr := strings.Split(image, ":")
	if len(image_arr) < 2 {
		return errors.New("镜像地址有误！")
	}
	// Get the tag of the repository.  /repositories/{repo_name}/tags/{tag}
	harbor_client := NewClient(nil, base_url,harbor_user,common.AesDecrypt(harbor_pwd))
	tag_resp, _, errs := harbor_client.Repositories.GetRepositoryTag(image_arr[0], image_arr[1])
	beego.Info(errs)
	if tag_resp.Size > 0 {
		return nil
	} else {
		return errors.New("镜像不存在")
	}
}
