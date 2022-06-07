package controllers

import (
	"github.com/astaxie/beego"
	"library/common"
	"strings"
	"initial"
	"encoding/json"
	"library/datalist"
)

// 数据级别的权限控制，只验证登录，权限在controller里面做
type BaseController struct {
	beego.Controller
	UserId   string
	Username string
	// super-admin, admin, deploy-global, deploy-single, guest
	Role     string
}

func (c *BaseController) Prepare() {
	header := c.Ctx.Input.Header("Authorization")
	cookie := c.Ctx.GetCookie("Authorization")
	if header == "" && cookie == "" {
		c.SetJson(401, "", "没有登录！")
		return
	}
	code, msg, data := BaseCheck(header, cookie)
	if code == 401 || code == 0 {
		c.SetJson(code, data, msg)
		return
	}
	if data == nil {
		c.SetJson(0, "", "登录解析错误！")
		return
	}
	c.UserId = data.Userid
	c.Username = data.UserName
	c.Role = data.Role
	//beego.Info(c.Role)
}

func BaseCheck(header, cookie string) (int, string, *datalist.UserInfo)  {
	xauth := header
	if xauth == "" {
		xauth = cookie
	}
	auth_md5 := common.Md5String(strings.Replace(xauth, "Basic ", "", -1))
	// 为空取数据
	if !initial.GetCache.IsExist(auth_md5) || common.GetString(initial.GetCache.Get(auth_md5)) == "" {
		code, msg, data := DBCheckToken(auth_md5, xauth)
		return code, msg, data
	} else {
		login_string := initial.GetCache.Get(auth_md5)
		var user_info datalist.UserInfo
		err := json.Unmarshal([]byte(common.GetString(login_string)), &user_info)
		if err != nil {
			return 0, err.Error(), nil
		}
		return 1, "校验成功!", &user_info
	}
}

func (c *BaseController) SetJson(code int, data interface{}, Msg string) {
	c.Data["json"] = map[string]interface{}{"code": code, "msg": Msg, "data": data}
	c.ServeJSON()
}

func (c *BaseController) GetPageRows() (int, int) {
	page, err := c.GetInt("page")
	if err != nil {
		page = 1
	}
	rows, err := c.GetInt("rows")
	if err != nil {
		rows = 15
	}
	return page, rows
}
