package itest

import (
	"controllers"
	"strings"
	"library/harbor"
	"github.com/astaxie/beego"
	"encoding/json"
)

type FuncTestController struct {
	controllers.BaseController
}

func (c *FuncTestController) URLMapping() {
	c.Mapping("Get", c.Get)
}

// @Title 函数测试接口
// @Description 测试函数的专用接口
// @Param	input	query	string	true	"输入参数"
// @Success 200 {object} {}
// @Failure 403
// @router /func [get]
func (c *FuncTestController) Get() {
	// image := "harbor.uat.cmft.com/cmrh-library/cmrh-di-glcs-web-20191120:latest"
	image := c.GetString("input")
	image = strings.Replace(image, "harbor.uat.cmft.com/", "", -1)
	image = strings.Replace(image, "harbor.cmft.com/", "", -1)
	image_arr := strings.Split(image, ":")
	if len(image_arr) < 2 {
		c.SetJson(1, "", "镜像地址有误！")
	}
	// Get the tag of the repository.  /repositories/{repo_name}/tags/{tag}
	harbor_client := harbor.NewClient(nil, "https://harbor.uat.cmft.com",
		"username","password")
	tag_resp, _, errs := harbor_client.Repositories.GetRepositoryTag(image_arr[0], image_arr[1])
	ret_body, _ := json.Marshal(tag_resp)
	beego.Info(string(ret_body))
	beego.Info(errs)
	c.SetJson(1, tag_resp, "发布单元获取成功！")
}
