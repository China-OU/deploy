package login

import (
	"strings"
	"library/common"
	"initial"
	"models"
	"github.com/astaxie/beego"
)

type LogoutController struct {
	beego.Controller
}

// Post方法
// @Title Post
// @Description 登出，带header为：Authorization: Basic xxxxxx，支持cookie和header
// @Success 200 {string} 登出成功或失败
// @Failure 403
// @router /logout [post]
func (c *LogoutController) Post() {
	header := c.Ctx.Input.Header("Authorization")
	cookie := c.Ctx.GetCookie("Authorization")
	if header == "" && cookie == "" {
		c.SetJson(401, "", "没有登录！")
		return
	}
	xauth := header
	if xauth == "" {
		xauth = cookie
	}
	auth_md5 := common.Md5String(strings.Replace(xauth, "Basic ", "", -1))
	if initial.GetCache.IsExist(auth_md5) {
		err := initial.GetCache.Delete(auth_md5)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	}
	// 删除token
	tx := initial.DB.Begin()
	err := tx.Where("token_md5 = ?", auth_md5).Delete(models.UserToken{}).Error
	if err != nil {
		tx.Rollback()
		beego.Error(err.Error())
	}
	tx.Commit()
	c.SetJson(1, "", "登出成功！")
}

func (c *LogoutController) SetJson(code int, data interface{}, Msg string) {
	c.Data["json"] = map[string]interface{}{"code": code, "msg": Msg, "data": data}
	c.ServeJSON()
}
