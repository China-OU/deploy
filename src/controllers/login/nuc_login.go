package login

import (
	"github.com/astaxie/beego"
	"encoding/json"
	"strings"
	"library/common"
	"github.com/astaxie/beego/httplib"
	"io/ioutil"
	"initial"
	"time"
)

type NucLoginController struct {
	beego.Controller
}

func (c *NucLoginController) URLMapping() {
	c.Mapping("DirectLogin", c.DirectLogin)          // 直接登录，用户名+密码+验证码
	c.Mapping("VerifyCode", c.VerifyCode)            // 获取图形验证码
	c.Mapping("TwoFactorLogin", c.TwoFactorLogin)    // 双因子登录，发送短信
	c.Mapping("RefreshSMS", c.RefreshSMS)            // 刷新短信
	c.Mapping("ValidateSMS", c.ValidateSMS)          // 双因子登录校验
	c.Mapping("LogOut", c.LogOut)                    // 登出
}

func (c *NucLoginController) SetJson(code int, data interface{}, Msg string) {
	c.Data["json"] = map[string]interface{}{"code": code, "msg": Msg, "data": data}
	c.ServeJSON()
}

// @Title Post
// @Description 调nuc直接登录接口，用户名+密码+验证码
// @Param	body	body	login.LoginInputData	true	"用户名和密码，密码是加密过的密码，不是原始密码。body形式的数据"
// @Success 200 {object} {}
// @Failure 403
// @router /direct/login [post]
func (c *NucLoginController) DirectLogin() {
	var input LoginInputData
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		c.SetJson(0, "", "参数输入有误！")
		return
	}
	// 测试默认：CmftNuc@#2019#         lxmAIIhsV6j+6M7/hd1imw==
	username := strings.Trim(input.UserName, " ")
	password := input.Password
	if username == "" || password == "" {
		c.SetJson(0, "", "用户名或密码为空！")
		return
	}
	pwd := strings.Trim(common.WebPwdDecrypt(password), " ")
	if pwd == "" {
		c.SetJson(0, "", "密码长度不对！")
		return
	}

	nuc_base_url := beego.AppConfig.String("nuc_base_url")
	req := httplib.Post(nuc_base_url + "LoginApp/v1/login")
	req.Header("Content-Type", "application/json")
	req.Header("CallerModule", "CMFT_UAD")
	login_option := LoginOption{
		EncodePasswd: "true",
		ModuleCode: "CMFT_UAD",
		ImageUUID: input.ImageUUID,
	}
	login_body := NucLoginData{
		AccountId: username,
		Password: pwd,
		VerifyCode: strings.TrimSpace(input.VerifyCode),
		Options: login_option,
	}
	data, _ := json.Marshal(login_body)
	req.Body(data)
	rs, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	var login_data NucBaseRet
	login_data.Data = LoginRet{}
	err = json.Unmarshal([]byte(rs), &login_data)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if login_data.State == false {
		c.SetJson(0, "", login_data.Message)
		return
	}
	c.SetJson(1, login_data, "接口访问成功")
}

// @Title Get
// @Description 调nuc获取图形验证码
// @Success 200 {object} {}
// @Failure 403
// @router /verify/code [get]
func (c *NucLoginController) VerifyCode() {
	nuc_base_url := beego.AppConfig.String("nuc_base_url")
	req := httplib.Get(nuc_base_url + "VerifyCodeImage.jpg")
	req.Header("Content-Type", "image/jpeg")
	req.Header("CallerModule", "CMFT_UAD")
	req.Param("type", "getCode")
	resp, err := req.DoRequest()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	type VerifyCode struct {
		UUID  string `json:"imageUUID"`
		Image []byte `json:"image"`
	}
	var verify VerifyCode
	defer resp.Body.Close()
	verify.Image, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "_imageUUID" {
			verify.UUID = cookie.Value
			break
		}
	}
	if verify.UUID == "" {
		c.SetJson(0, "", "image uuid not found")
		return
	}
	c.SetJson(1, verify, "接口访问成功")
}

// @Title Post
// @Description 双因子登录，发送短信
// @Param	body	body	login.LoginInputData	true	"用户名和密码，密码是加密过的密码，不是原始密码。body形式的数据"
// @Success 200 {object} {}
// @Failure 403
// @router /twofactor/login [post]
func (c *NucLoginController) TwoFactorLogin() {
	var input LoginInputData
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		c.SetJson(0, "", "参数输入有误！")
		return
	}
	// 测试默认：CmftNuc@#2019#         lxmAIIhsV6j+6M7/hd1imw==
	username := strings.Trim(input.UserName, " ")
	password := input.Password
	if username == "" || password == "" {
		c.SetJson(0, "", "用户名或密码为空！")
		return
	}
	pwd := strings.Trim(common.WebPwdDecrypt(password), " ")
	if pwd == "" {
		c.SetJson(0, "", "密码长度不对！")
		return
	}
	if initial.GetCache.IsExist(username) && common.GetString(initial.GetCache.Get(username)) != "" {
		c.SetJson(0, "", "NUC双因子短信获取接口访问间隔小于一分钟！")
		return
	}

	nuc_base_url := beego.AppConfig.String("nuc_base_url")
	req := httplib.Post(nuc_base_url + "LoginApp/v1/loginWithTwoFactorsAuth")
	req.Header("Content-Type", "application/json")
	req.Header("CallerModule", "CMFT_UAD")
	req.Header("X-FORWARDED-FOR", "113.116.91.23")
	login_option := LoginOption{
		EncodePasswd: "true",
		ModuleCode: "CMFT_UAD",
		ImageUUID: input.ImageUUID,
		AuthType: "passwdMobile",
	}
	login_body := NucLoginData{
		AccountId: username,
		Password: pwd,
		VerifyCode: strings.TrimSpace(input.VerifyCode),
		Options: login_option,
	}
	data, _ := json.Marshal(login_body)
	req.Body(data)
	rs, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 设60秒缓存
	type MsgRet struct {
		State bool `json:"state"`
	}
	var msg_ret MsgRet
	err = json.Unmarshal([]byte(rs), &msg_ret)
	if err != nil {
		beego.Info(rs)
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if msg_ret.State == true {
		initial.GetCache.Put(username, time.Now().Format(initial.DatetimeFormat), 60*time.Second)
	}
	c.SetJson(1, rs, "接口访问成功")
}

// @Title Post
// @Description 刷新短信
// @Param	body	body	login.RefreshOption	true	"messageId从获取短信的接口获取，在此不新增"
// @Success 200 {object} {}
// @Failure 403
// @router /refresh/sms [post]
func (c *NucLoginController) RefreshSMS() {
	var input RefreshOption
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		c.SetJson(0, "", "参数输入有误！")
		return
	}
	if input.Username == "" {
		c.SetJson(0, "", "需要输入用户名！")
		return
	}
	if initial.GetCache.IsExist(input.Username) && common.GetString(initial.GetCache.Get(input.Username)) != "" {
		c.SetJson(0, "", "NUC双因子短信获取接口访问间隔小于一分钟！")
		return
	}

	nuc_base_url := beego.AppConfig.String("nuc_base_url")
	req := httplib.Post(nuc_base_url + "LoginApp/v1/refreshMessageCode")
	req.Header("Content-Type", "application/json")
	req.Header("CallerModule", "CMFT_UAD")
	req.Header("X-FORWARDED-FOR", "113.116.91.23")
	data, _ := json.Marshal(input)
	req.Body(data)
	rs, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, rs, "接口访问成功")
}

// @Title Post
// @Description 双因子登录校验，最后的登录接口，返回token等数据
// @Param	body	body	login.SMSValidateOption	true	"验证码和messageId来登录"
// @Success 200 {object} {}
// @Failure 403
// @router /validate/sms [post]
func (c *NucLoginController) ValidateSMS() {
	var input SMSValidateOption
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		c.SetJson(0, "", "参数输入有误！")
		return
	}

	nuc_base_url := beego.AppConfig.String("nuc_base_url")
	req := httplib.Post(nuc_base_url + "LoginApp/v1/validateMessageCode")
	req.Header("Content-Type", "application/json")
	req.Header("CallerModule", "CMFT_UAD")
	req.Header("X-FORWARDED-FOR", "113.116.91.23")
	data, _ := json.Marshal(input)
	req.Body(data)
	rs, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, rs, "接口访问成功")
}

// @Title Post
// @Description 双因子登录校验
// @Param	body	body	datalist.LoginInputData	true	"用户名和密码，密码是加密过的密码，不是原始密码。body形式的数据"
// @Success 200 {object} {}
// @Failure 403
// @router /logout [post]
func (c *NucLoginController) LogOut() {
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

	nuc_base_url := beego.AppConfig.String("nuc_base_url")
	req := httplib.Post(nuc_base_url + "LoginApp/v1/logout")
	req.Header("Content-Type", "application/json")
	req.Header("CallerModule", "CMFT_UAD")
	req.Header("Authorization", header)
	rs, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, rs, "接口访问成功")
}