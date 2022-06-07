package login

import (
	"controllers"
	"github.com/astaxie/beego"
	"math/rand"
	"time"
	"initial"
	"models"
	"library/datalist"
	"library/common"
)

type LoginCheckController struct {
	beego.Controller
}

func (c *LoginCheckController) URLMapping() {
	c.Mapping("Post", c.Post)
}

// Post方法
// @Title Post
// @Description 登录校验，支持cookie和header，五种角色三种视图：，super-admin, admin, deploy-global, deploy-single, guest
// @Success 200 {object} datalist.UserInfo
// @Failure 403
// @router /check [post]
func (c *LoginCheckController) Post() {
	header := c.Ctx.Input.Header("Authorization")
	cookie := c.Ctx.GetCookie("Authorization")
	if header == "" && cookie == "" {
		c.SetJson(401, "", "没有登录！")
		return
	}
	code, msg, data := controllers.BaseCheck(header, cookie)
	if code == 1 {
		sec := time.Now().Second()
		if sec%60 == 1 {
			// 删除过期token，一分钟一次
			go DelExpiredToken()
		}
		if sec%5 == 1 {
			// 滚动更新过期时间，更新上就后滚，更新不上就四小时过期
			go UpdateExpireTime(*data)
		}
	}
	c.SetJson(code, data, msg)
}

func DelExpiredToken() {
	rand.Seed(time.Now().Unix())
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	tx := initial.DB.Begin()
	err := tx.Where("expire < NOW()").Delete(models.UserToken{}).Error
	if err != nil {
		tx.Rollback()
		beego.Error(err.Error())
		return
	}
	tx.Commit()
}

func UpdateExpireTime(data datalist.UserInfo) {
	rand.Seed(time.Now().Unix())
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	token_md5 := common.Md5String(data.Token)
	new_expire := time.Now().Add(4 * time.Hour).Format(initial.DatetimeFormat)
	// 安全加固30分钟，后续注释
	now_date := time.Now().Format(initial.DateFormat)
	if beego.AppConfig.String("runmode") == "prd" && now_date > "20200626" && now_date < "20200720" {
		new_expire = time.Now().Add(30 * time.Minute).Format(initial.DatetimeFormat)
	}
	// 安全加固30分钟，后续注释
	tx := initial.DB.Begin()
	err := tx.Model(models.UserToken{}).Where("token_md5=?", token_md5).Update("expire", new_expire).Error
	if err != nil {
		tx.Rollback()
		beego.Error(err.Error())
		return
	}
	tx.Commit()
}

func (c *LoginCheckController) SetJson(code int, data interface{}, Msg string) {
	c.Data["json"] = map[string]interface{}{"code": code, "msg": Msg, "data": data}
	c.ServeJSON()
}