package login

import (
	"github.com/astaxie/beego"
	"encoding/json"
	"library/datalist"
	"models"
	"time"
	"library/common"
	"initial"
	"strings"
	"controllers"
	"github.com/astaxie/beego/httplib"
)

type LoginController struct {
	beego.Controller
}

func (c *LoginController) URLMapping() {
	c.Mapping("Post", c.Post)
}

// Post方法
// @Title Post
// @Description 登录校验，返回结果：401未登录，1成功，0或2是失败，以后所有的接口返回都是这样，在此只写一次。
// @Param	body	body	datalist.LoginInputData	true	"用户名和密码，密码是加密过的密码，不是原始密码。body形式的数据"
// @Success 200 {object} datalist.UserInfo
// @Failure 403
// @router / [post]
func (c *LoginController) Post() {
	var input datalist.LoginInputData
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		c.Data["json"] = map[string]interface{}{"code": 0, "msg": "参数输入有误！"}
		c.ServeJSON()
		return
	}
	username := strings.Trim(input.UserName, " ")
	password := input.Password
	if username == "" || password == "" {
		c.Data["json"] = map[string]interface{}{"code": 0, "msg": "用户名或密码为空"}
		c.ServeJSON()
		return
	}
	pwd := strings.Trim(common.WebPwdDecrypt(password), " ")
	if pwd == "" {
		c.Data["json"] = map[string]interface{}{"code": 0, "msg": "密码长度不对"}
		c.ServeJSON()
		return
	}

	// 手动构建登录接口，只允许两个登录，后续关掉
	//if pwd != "mdeploy" {
	//	c.Data["json"] = map[string]interface{}{"code": 0, "msg": "密码不正确，请重新输入！"}
	//	c.ServeJSON()
	//	return
	//}
	//var info datalist.UserInfo
	//info.Userid = username
	//info.UserName = username
	//info.Phone = "18126350000"
	//info.Email = username + "@cmft.com"
	//info.Title = "员工"
	//info.Department = "系统运营部"
	//info.Token = "eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiJjbXRva2VuIiwiaWF0IjoxNTY1ODU3NDEyLCJzdWIiOiJ7XCJ1c2VyVHlwZVwiOlwiU1RBRkZcIixcInVzZXJJZFwiOlwibGl5MDAxQGNtcmguY29tXCIsXCJ0b2tlblwiOlwidENHMGZnL29CSFhwUTB1eFRvMHFXUT09XCJ9IiwiZXhwIjoxNTY1OTQzODEyfQ.wlWXf6iJhDXvKRx5ekYA1kRiFFQ3tEuCGDvJpx0p"
	//rand.Seed(time.Now().Unix())
	//info.Token = info.Token + common.GetString(rand.Intn(899) + 100) + username
	//info.Company = "招商金融科技"
	//info.Center = "运维中心"
	//info.Role = controllers.GetLoginUserRole(info.Userid)

	//封网期间，登录代码限制死，后续打开
	cmtoken := beego.AppConfig.String("cmtoken")
	req := httplib.Post(cmtoken + "login")
	req.Header("Content-Type", "application/json")
	param := map[string]string{}
	param["username"] = username
	param["password"] = pwd
	param["userType"] = "STAFF"
	param["systemId"] = beego.AppConfig.String("system_id")
	data, _ := json.Marshal(param)
	req.Body(data)
	rs, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		c.Data["json"] = map[string]interface{}{"code": 2, "msg": err.Error()}
		c.ServeJSON()
		return
	}
	var login_data datalist.LoginData
	err = json.Unmarshal([]byte(rs), &login_data)
	if err != nil {
		beego.Error(err.Error())
		c.Data["json"] = map[string]interface{}{"code": 2, "msg": err.Error()}
		c.ServeJSON()
		return
	}
	if login_data.Code != "Y" {
		c.Data["json"] = map[string]interface{}{"code": 0, "msg": login_data.Message}
		c.ServeJSON()
		return
	}
	var info datalist.UserInfo
	info.Userid = login_data.Data.UmDTO.UnId
	if info.Userid == "" {
		info.Userid = username
	}
	info.UserName = login_data.Data.UmDTO.UserName
	info.Email = login_data.Data.UmDTO.UserId
	info.Token = login_data.Data.Token
	if len(login_data.Data.UmDTO.Users) > 0 {
		info.Department = login_data.Data.UmDTO.Users[0]["department"]
		info.Title = login_data.Data.UmDTO.Users[0]["title"]
		info.Phone = login_data.Data.UmDTO.Users[0]["mobilePhone"]
		info.Company = login_data.Data.UmDTO.Users[0]["company"]
		info.Center = login_data.Data.UmDTO.Users[0]["center"]
	}
	info.Role = controllers.GetLoginUserRole(info.Userid)

	go InsertLoginUser(info)
	go CheckAndSession(info)
	c.Data["json"] = map[string]interface{}{"code": 1, "msg": "登录成功", "data": info}
	c.ServeJSON()
}

// 记录登录人员
func InsertLoginUser(info datalist.UserInfo)  {
	cnt := 0
	initial.DB.Model(models.UserLogin{}).Where("userid = ?", info.Userid).Count(&cnt)
	tx := initial.DB.Begin()
	if cnt == 0 {
		var m models.UserLogin
		m.Userid = info.Userid
		m.UserName = info.UserName
		m.Phone = info.Phone
		m.Email = info.Email
		m.Title = info.Title
		m.Company = info.Company
		m.Center = info.Center
		m.Department = info.Department
		m.InsertTime = time.Now().Format("2006-01-02 15:04:05")
		err := tx.Create(&m).Error
		if err != nil {
			tx.Rollback()
			beego.Error(err.Error())
		}
	}
	tx.Commit()
}

// 缓存用户信息，录入数据库
func CheckAndSession(info datalist.UserInfo) {
	token_md5 := common.Md5String(info.Token)
	if !initial.GetCache.IsExist(token_md5) || common.GetString(initial.GetCache.Get(token_md5)) == "" {
		login_info_byte, _ := json.Marshal(info)
		initial.GetCache.Put(token_md5, string(login_info_byte), 1 * time.Minute)
		// 录入数据库
		cnt := 0
		initial.DB.Model(models.UserToken{}).Where("token_md5 = ?", token_md5).Count(&cnt)
		tx := initial.DB.Begin()
		if cnt == 0 {
			var token models.UserToken
			token.UserId = info.Userid
			token.TokenMd5 = token_md5
			token.Email = info.Email
			token.Expire = time.Now().Add(4 * time.Hour).Format(initial.DatetimeFormat)
			info_json, _ := json.Marshal(info)
			token.Info = string(info_json)
			err := tx.Create(&token).Error
			if err != nil {
				tx.Rollback()
				beego.Error(err.Error())
			}
		}
		tx.Commit()
	}
}


