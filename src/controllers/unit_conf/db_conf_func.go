package unit_conf

// 主要是操作agent那一部分的功能
import (
	"strings"
	"models"
	"initial"
	"github.com/astaxie/beego"
	"controllers"
	"fmt"
	"library/cfunc"
	"library/common"
	"net/url"
	"time"
	"encoding/json"
	"github.com/astaxie/beego/httplib"
	"github.com/sethvargo/go-password/password"
	"library/database"
	"math/rand"
)

// @Title 数据库连通性测试
// @Description 数据库连通性测试，查看数据库是否正常连接
// @Param	id	query	string	true	"数据列的id"
// @Success 200 true or false
// @Failure 403
// @router /db/conn/check [get]
func (c *DBConfListController) ConnCheck() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	id, _ := c.GetInt("id")
	var info models.UnitConfDb
	err := initial.DB.Model(models.UnitConfDb{}).Where("id=? and is_delete=0", id).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && controllers.CheckUnitSingleAuth(info.UnitId, c.UserId) == false {
		c.SetJson(0, "", "您没有权限操作，只有发布单元相关人员有权限进行连通性测试！")
		return
	}

	// 访问agent测试连通性，在此略去不写，后续再补充
	agent, err := cfunc.GetAgentConfByComp(info.DeployComp)
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
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/%s/conn", agent.AgentIp, agent.AgentPort, info.Type)
	req := httplib.Post(url + "?" + query_str)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(1*time.Minute, 1*time.Minute)

	param := make(map[string]interface{})
	param["comp"] = info.DeployComp
	param["type"] = info.Type
	param["host"] = info.Host
	param["port"] = info.Port
	param["user"] = info.Username
	param["pwd"] = common.AgentEncrypt(common.AesDecrypt(info.EncryPwd))
	param["database"] = info.Dbname
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
	// 更新连通性测试结果
	tx := initial.DB.Begin()
	umap := map[string]interface{}{
		"conn_result": ret.Code,
		"conn_ctime": time.Now().Format(initial.DatetimeFormat),
	}
	err = tx.Model(models.UnitConfDb{}).Where("id=? and is_delete=0", id).Updates(umap).Error
	if err != nil {
		tx.Rollback()
		beego.Error(err.Error())
		c.SetJson(0, "", "数据库可连通，更新连通性结果出现问题，报错为："+err.Error())
		return
	}
	tx.Commit()
	if ret.Code != 1 {
		c.SetJson(0, "", ret.Msg)
		return
	}
	c.SetJson(1, ret.Data, "数据库可正常访问！")
}

// @Title 修改数据库密码，只有管理员和系统负责人可以修改
// @Description 修改数据库密码，只有管理员和系统负责人可以修改
// @Param	id	query	string	true	"数据列的id"
// @Success 200 true or false
// @Failure 403
// @router /db/change/pwd [post]
func (c *DBConfListController) ChangePwd() {
	show_auth := 0
	id, _ := c.GetInt("id")
	var info models.UnitConfDb
	err := initial.DB.Model(models.UnitConfDb{}).Where("id=? and is_delete=0", id).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if strings.Contains(c.Role, "admin") == true {
		show_auth = 1
	}
	if show_auth == 0 {
		c.SetJson(0, "", "您没有权限修改密码，第一阶段只有管理员才可以修改密码！")
		return
	}
	if info.PwdCtime > time.Now().AddDate(0, -1, 0).Format(initial.DatetimeFormat) {
		c.SetJson(0, "", "一个月内只能修改一次密码！！！")
		return
	}

	// 生成密码
	gen, _ := password.NewGenerator(&password.GeneratorInput{
		Symbols: "#^*",
	})
	password, err := gen.Generate(16, 4, 2, false, false)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	// 避免出现 #开头
	ls := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rands := rand.Intn(50)
	pwd := ls[rands: rands+1] + password
	beego.Info(pwd)

	// 访问agent测试连通性，在此略去不写，后续再补充
	agent, err := cfunc.GetAgentConfByComp(info.DeployComp)
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
	url := "http://" + fmt.Sprintf("%s:%s/agent/v1/db/%s/changepwd", agent.AgentIp, agent.AgentPort, info.Type)
	req := httplib.Post(url + "?" + query_str)
	req.Header("agent-auth", initial.AgentToken)
	req.SetTimeout(1*time.Minute, 1*time.Minute)

	param := make(map[string]interface{})
	param["comp"] = info.DeployComp
	param["type"] = info.Type
	param["host"] = info.Host
	param["port"] = info.Port
	param["user"] = info.Username
	param["pwd"] = common.AgentEncrypt(common.AesDecrypt(info.EncryPwd))
	param["database"] = info.Dbname
	param["new_pwd"] = pwd
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
	// 更新连通性测试结果
	if ret.Code == 1 {
		tx := initial.DB.Begin()
		umap := map[string]interface{}{
			"encry_pwd": common.AesEncrypt(pwd),
			"pwd_ctime": time.Now().Format(initial.DatetimeFormat),
		}
		// 多个发布单元共用一个库，需要一起改
		err = tx.Model(models.UnitConfDb{}).Where("host=? and dbname=? and username=? and is_delete=0", info.Host,
			info.Dbname, info.Username).Updates(umap).Error
		if err != nil {
			tx.Rollback()
			beego.Error(err.Error())
			c.SetJson(0, "", "数据库密码修改成功，部署平台密码保存失败，错误为："+err.Error())
			return
		}
		tx.Commit()
	} else {
		c.SetJson(0, "", "数据库密码修改失败，报错为："+ret.Msg)
		return
	}
	c.SetJson(1, common.WebPwdEncrypt(pwd), "密码修改成功！")
}

// @Title 获取数据库失效对象
// @Description 获取数据库失效对象
// @Param	id	query	string	true	"数据列的id"
// @Success 200 true or false
// @Failure 403
// @router /db/invalid/pkg [post]
func (c *DBConfListController) InvalidPkg() {
	if c.Role == "guest" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	id, _ := c.GetInt("id")
	var info models.UnitConfDb
	err := initial.DB.Model(models.UnitConfDb{}).Where("id=? and is_delete=0", id).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && controllers.CheckUnitSingleAuth(info.UnitId, c.UserId) == false {
		c.SetJson(0, "", "您没有权限操作，只有发布单元相关人员有权限编译失效对象！")
		return
	}
	if info.Type != "oracle" {
		c.SetJson(0, "", "失效对象检测目前只支持oracle！")
		return
	}
	sql_name := "pkg_invalid.sql"
	data, err := database.OracleCommonQuery(info, sql_name)
	if err != nil {
		c.SetJson(0, data, err.Error())
		return
	}
	c.SetJson(1, data, "失效对象查询成功！")
}

// @Title 编译数据库失效对象
// @Description 编译数据库失效对象
// @Param	id	query	string	true	"数据列的id"
// @Success 200 true or false
// @Failure 403
// @router /db/compile/pkg [post]
func (c *DBConfListController) CompilePkg() {
	if c.Role == "guest" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	id, _ := c.GetInt("id")
	var info models.UnitConfDb
	err := initial.DB.Model(models.UnitConfDb{}).Where("id=? and is_delete=0", id).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && controllers.CheckUnitSingleAuth(info.UnitId, c.UserId) == false {
		c.SetJson(0, "", "您没有权限操作，只有发布单元相关人员有权限编译失效对象！")
		return
	}
	if info.Type != "oracle" {
		c.SetJson(0, "", "失效对象检测目前只支持oracle！")
		return
	}
	sql_name := "pkg_compile.sql"
	data, err := database.OracleCommonQuery(info, sql_name)
	if err != nil {
		c.SetJson(0, data, err.Error())
		return
	}
	c.SetJson(1, data, "失效对象编译通过！")
}

// @Title 锁状态检测
// @Description 锁状态检测
// @Param	id	query	string	true	"数据列的id"
// @Success 200 true or false
// @Failure 403
// @router /db/lock/check [get]
func (c *DBConfListController) LockCheck() {
	if c.Role == "guest" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	id, _ := c.GetInt("id")
	var info models.UnitConfDb
	err := initial.DB.Model(models.UnitConfDb{}).Where("id=? and is_delete=0", id).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && controllers.CheckUnitSingleAuth(info.UnitId, c.UserId) == false {
		c.SetJson(0, "", "您没有权限操作，只有发布单元相关人员有权限进行锁检测！")
		return
	}
	if info.Type != "mysql" {
		c.SetJson(0, "", "锁状态检测目前只支持mysql！")
		return
	}

	sql_str := "SHOW PROCESSLIST;"
	data, err := database.MysqlCommonQuery(info, sql_str)
	if err != nil {
		c.SetJson(0, data, err.Error())
		return
	}
	c.SetJson(1, data, "数据库活跃进程查询成功！")
}