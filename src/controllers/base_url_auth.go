package controllers

import (
	"github.com/astaxie/beego"
	"strings"
	"errors"
	"strconv"
	"time"
	"net/url"
	"sort"
	"fmt"
	"initial"
	"models"
	"library/common"
	"encoding/json"
)

// url签名认证
type BaseUrlAuthController struct {
	beego.Controller
	Username string
	Role     string
}

func (c *BaseUrlAuthController) Prepare() {
	data, err := c.verifySign()
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	if data != nil {
		c.SetJson(0, data, "")
	}
}

// 验证签名
func (c *BaseUrlAuthController) verifySign() (map[string]string, error) {
	_ = c.Ctx.Request.ParseForm()
	req := c.Ctx.Request.Form
	debug := strings.Join(req["debug"], "")
	ak := strings.Join(req["ak"], "")
	sn := strings.Join(req["sn"], "")
	ts := strings.Join(req["ts"], "")
	if ak == "" {
		return nil, errors.New("用户名不能为空！")
	}

	// 验证来源
	var url_role models.UrlRole
	err := initial.DB.Model(models.UrlRole{}).Where("app_key=?", ak).First(&url_role).Error
	if err != nil {
		return nil, err
	}
	c.Username = url_role.AppKey
	c.Role = url_role.Role

	// 将body里面的数据添加进来
	if string(c.Ctx.Input.RequestBody) != "" {
		var body_data map[string]interface{}
		err = json.Unmarshal(c.Ctx.Input.RequestBody, &body_data)
		if err != nil {
			return nil, err
		}
		for k, v := range body_data {
			req.Set(k, common.GetString(v))
		}
	}

	if debug == "1" && beego.AppConfig.String("runmode") == "dev" {
		current_unix := time.Now().Unix()
		req.Set("ts", strconv.FormatInt(current_unix, 10))
		res := map[string]string{
			"ts": strconv.FormatInt(current_unix, 10),
			"sn": createSign(req, url_role.SecretKey),
		}
		return res, nil
	}

	// 验证过期时间
	timestamp := time.Now().Unix()
	ts_int, _ := strconv.ParseInt(ts, 10, 64)
	if ts_int > timestamp || timestamp - ts_int > 60 {
		return nil, errors.New("ts expired!")
	}

	// 验证签名
	if sn == "" || sn != createSign(req, url_role.SecretKey) {
		return nil, errors.New("sn error!")
	}
	return nil, nil
}

// 创建签名
func createSign(params url.Values, sk string) string {
	return common.Md5String(fmt.Sprintf("%s-%s-%s", sk, createEncryptStr(params), sk))
}

func createEncryptStr(params url.Values) string {
	var key []string
	var str = ""
	for k := range params {
		if k != "sn" && k != "debug" {
			key = append(key, k)
		}
	}
	sort.Strings(key)
	for i := 0; i < len(key); i++ {
		if i == 0 {
			str = fmt.Sprintf("%v=%v", key[i], params.Get(key[i]))
		} else {
			str = str + fmt.Sprintf("&%v=%v", key[i], params.Get(key[i]))
		}
	}
	beego.Info(str)
	return str
}

func (c *BaseUrlAuthController) SetJson(code int, data interface{}, Msg string) {
	c.Data["json"] = map[string]interface{}{"code": code, "msg": Msg, "data": data}
	c.ServeJSON()
}
