package ext

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"initial"
	"library/common"
	"models"
	"net/url"
	"strings"
	"time"
)

type DCCInput struct {
	AgentIp     string  `json:"agent_ip"`
	AgentPort   string  `json:"agent_port"`
	DeployComp  string  `json:"deploy_comp"`
	Host        string  `json:"host"`
	Dbname      string  `json:"dbname"`
}

// @Title oracle连通性测试，给预发布激活数据库专用，不作其他用途; 用明文不用密文
// @Description oracle连通性测试，给预发布激活数据库专用，不作其他用途
// @Param	body	body	ext.DCCInput	true	"body形式的数据"
// @Success 200 true or false
// @Failure 403
// @router /db/conn/check [post]
func (c *MultiEnvConnController) DbConnCheck() {
	header := c.Ctx.Request.Header
	auth := ""
	if header["Authorization"] != nil && len(header["Authorization"]) > 0 {
		auth = header["Authorization"][0]
	} else {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "没有header!"}
		c.ServeJSON()
		return
	}
	if strings.Replace(auth, "Basic ", "", -1) != "mdeploy_IpFhvFjiQpVw4cDIUywc3VHDjC0Wo9EM" {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "header校验失败!"}
		c.ServeJSON()
		return
	}

	var input DCCInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	var info models.UnitConfDb
	err = initial.DB.Model(models.UnitConfDb{}).Where("host=? and dbname=? and is_delete=0", input.Host, input.Dbname).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	ip_list := strings.Join(common.GetLocalIp(), ",")
	args := url.Values{
		"ip_list": []string{ip_list},
	}
	query_str := args.Encode()
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/oracle/conn", input.AgentIp, input.AgentPort)
	req := httplib.Post(url + "?" + query_str)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(10*time.Second, 10*time.Second)

	param := make(map[string]interface{})
	param["comp"] = input.DeployComp
	param["type"] = "oracle"
	param["host"] = input.Host
	param["port"] = info.Port
	param["user"] = info.Username
	param["pwd"] = common.AgentEncrypt(common.AesDecrypt(info.EncryPwd))
	param["database"] = input.Dbname
	data, _ := json.Marshal(param)
	req.Body(data)
	ret_bytes, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	type StructRet struct {
		Code int `json:"code"`
		Msg  string `json:"msg"`
		Data interface{} `json:"data"`
	}
	var ret StructRet
	err = json.Unmarshal(ret_bytes, &ret)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if ret.Code != 1 {
		c.SetJson(0, "", ret.Msg)
		return
	}
	c.SetJson(1, ret.Data, "数据库可正常访问！")
}

func (c *MultiEnvConnController) SetJson(code int, data interface{}, Msg string) {
	c.Data["json"] = map[string]interface{}{"code": code, "msg": Msg, "data": data}
	c.ServeJSON()
}
