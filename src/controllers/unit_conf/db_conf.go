package unit_conf

import (
	"controllers"
	"strings"
	"fmt"
	"initial"
	"models"
	"github.com/astaxie/beego"
	"library/cfunc"
	"library/common"
)

type DBConfListController struct {
	controllers.BaseController
}

func (c *DBConfListController) URLMapping() {
	c.Mapping("GetDBList", c.GetDBList)
	c.Mapping("DBSave", c.DBSave)
	c.Mapping("DBDel", c.DBDel)
	c.Mapping("ConnCheck", c.ConnCheck)        // 连通性测试
	c.Mapping("ChangePwd", c.ChangePwd)        // 修改deployop密码
	c.Mapping("InvalidPkg", c.InvalidPkg)      // 编译失效对象
	c.Mapping("CompilePkg", c.CompilePkg)      // 编译失效对象
	c.Mapping("LockCheck", c.LockCheck)        // 锁状态检测
	c.Mapping("GetDBPwd", c.GetDBPwd)          // 获取数据库密码
}

// GetAll 方法
// @Title Get All
// @Description 获取所有发布单元列表，前端不显示密码，密码更改有记录可以查询，历史记录只能管理员查询
// @Param	db_type	query	string	false	"数据库类型，oracle/mysql/pgsql"
// @Param	en_name	query	string	false	"发布单元英文名，支持模糊搜索"
// @Param	host	query	string	false	"db的host，支持模糊搜索"
// @Param	conn_result	query	string	false	"连接正常，0/1/10"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} models.UnitConfList
// @Failure 403
// @router /db/list [get]
func (c *DBConfListController) GetDBList() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	db_type := c.GetString("db_type")
	en_name := c.GetString("en_name")
	host := c.GetString("host")
	connect := c .GetString("conn_result")
	page, rows := c.GetPageRows()
	cond := " a.is_delete=0 "
	if c.Role == "deploy-single" {
		cond += fmt.Sprintf(" and concat(b.leader, ',', b.developer, b.test) like '%%%s%%' ", c.UserId)
	}
	if db_type != "" {
		cond += fmt.Sprintf(" and a.type = '%s' ", db_type)
	}
	if strings.TrimSpace(en_name) != "" {
		cond += fmt.Sprintf(" and b.unit like '%%%s%%' ", en_name)
	}
	if strings.TrimSpace(host) != "" {
		cond += fmt.Sprintf(" and a.host like '%%%s%%' ", host)
	}
	if strings.TrimSpace(connect) != "" {
		cond += fmt.Sprintf(" and a.conn_result = '%s' ", connect)
	}

	type DbInfo struct {
		models.UnitConfDb
		Unit string `json:"unit"`
		Name string `json:"name"`
		Leader string `json:"leader"`
		LeaderName   string    `json:"leader_name"`
		CompName     string    `json:"comp_name"`
	}

	var cnt int
	var dblist []DbInfo
	err := initial.DB.Table("unit_conf_db a").Select("a.*, b.unit, b.name, b.leader").
		Joins("left join unit_conf_list b on a.unit_id = b.id").
		Where(cond).Count(&cnt).Order("a.id desc").Offset((page - 1)*rows).Limit(rows).Find(&dblist).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	for i:=0; i<len(dblist); i++ {
		dblist[i].LeaderName = cfunc.GetUserCnName(dblist[i].Leader)
		dblist[i].CompName = cfunc.GetCompCnName(dblist[i].DeployComp)
		dblist[i].EncryPwd = "*****"
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": dblist,
	}
	c.SetJson(1, ret, "数据库列表获取成功！")
}

// @Title DBDel
// @Description 删除数据库的配置
// @Param	id	query	string	true	"数据列的id"
// @Success 200 true or false
// @Failure 403
// @router /db/del [post]
func (c *DBConfListController) DBDel() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作，请联系管理员进行删除！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	tx := initial.DB.Begin()
	err = tx.Model(models.UnitConfDb{}).Where("id=?", id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "数据库的配置删除成功！")
}


// @Title 获取db的密码，可以明文显示到前端
// @Description 获取db的密码，可以明文显示到前端
// @Param	id	query	string	true	"数据列的id"
// @Success 200 true or false
// @Failure 403
// @router /db/password [get]
func (c *DBConfListController) GetDBPwd() {
	show_auth := 0
	id, _ := c.GetInt("id")
	var info models.UnitConfDb
	err := initial.DB.Model(models.UnitConfDb{}).Where("id=? and is_delete=0", id).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if strings.Contains(c.Role, "admin") == true || info.Operator == c.UserId {
		show_auth = 1
	}
	if show_auth == 0 {
		c.SetJson(0, "", "您没有权限查询密码，管理员和录入人员才有权限查看，录入人员为："+info.Operator)
		return
	}
	pwd:= common.WebPwdEncrypt(common.AesDecrypt(info.EncryPwd))
	ret := map[string]interface{}{
		"pwd": pwd,
		"modify_time": info.PwdCtime,
	}
	c.SetJson(1, ret, "密码获取成功！")
}

