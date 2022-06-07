package user_role

import (
	"controllers"
	"fmt"
	"github.com/jinzhu/gorm"
	"initial"
	"models"
	"strings"
	"time"
)

type RoleController struct {
	controllers.BaseController
}

func (c *RoleController) URMMapping() {
	c.Mapping("RoleList",c.GetAll)
	c.Mapping("OperationRole",c.OperationRole)
	c.Mapping("AddRole",c.AddRole)
	c.Mapping("GetUserPrivilege",c.GetUserPrivilege)
}

// @Title RoleList
// @Description 获取当前角色为admin/deploy-global的人员列表
// @Param	username	query	string	false	"用户名 按照用户名查询"
// @Param	role		query	string	false	"角色 按照角色查询"
// @Param	page	query	string	true	"第几页"
// @Param	rows	query	string	true	"每页有几行"
// @Success 200 {object} []user_role.URole
// @Failure 403
// @router /role/list [get]
func (c *RoleController) GetAll() {
	username := c.GetString("username")
	role := c.GetString("role")
	page, rows := c.GetPageRows()

	if !(strings.Contains(c.Role,"super-admin")  || strings.Contains(c.Role,"admin")) {
		c.SetJson(0,"","您没有权限！")
		return
	}

	var (
		err error
		urList []*URole
		cnt int
	)
	filter := "is_delete = 0 "
	if username != "" {
		filter = fmt.Sprintf("%s AND ( username like '%%%s%%' OR realname like '%%%s%%' )",filter,username,username)
	}
	if role == "admin" {
		filter = fmt.Sprintf("%s AND role = %s",filter,"'admin'")
	}else if role == "deploy-global" {
		filter = fmt.Sprintf("%s AND role = %s",filter,"'deploy-global'")
	}else if role == "" {
		filter = fmt.Sprintf("%s AND (role = 'admin' OR role = 'deploy-global' OR role = 'super-admin')",filter)
	}else {
		c.SetJson(0,"","角色类型只准为admin或deploy-global")
	}

	sql := `SELECT *, 
	CASE
		WHEN role = 'super-admin' THEN 'A'
		WHEN role = 'admin' THEN 'B'
		WHEN role = 'deploy-global' THEN 'C'
	END role_level
FROM user_role 
WHERE {filter}
ORDER BY role_level
LIMIT {rows} OFFSET {page}`
	sql = strings.Replace(sql, "{filter}", filter, -1)
	sql = strings.Replace(sql, "{rows}", fmt.Sprintf("%d", rows), -1)
	sql = strings.Replace(sql, "{page}", fmt.Sprintf("%d", (page - 1) * rows), -1)
	if err = initial.DB.Raw(sql).Find(&urList).Error ; err != nil {
		c.SetJson(0,"",err.Error())
		return
	}
	if err = initial.DB.Table("user_role").Where(filter).Count(&cnt).Error ; err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	res := make(map[string]interface{})
	res["cnt"] = cnt
	res["data"] = urList

	c.SetJson(1, res,"数据获取成功！")
}

type URole struct {
	Id int	`json:"id" gorm:"column:id"`
	Username string `json:"username" gorm:"column:realname"`
	Userid string `json:"userid" gorm:"column:username"`
	Role string `json:"role" gorm:"column:role"`
	InsertTime string `gorm:"column:insert_time" json:"insert_time"`
	Value int `json:"-"`
}

// @Title AddRole
// @Description 新增用户角色
// @Param	userid	query	string	true	"用户id 例如:ex-yanggq001"
// @Param	role	query	string	true	"用户角色: admin/deploy-global;"
// @Success 200 true or false
// @Failure 403
// @router /role/add [get]
func (c *RoleController) AddRole() {
	if !(strings.Contains(c.Role,"super-admin") || strings.Contains(c.Role,"admin")) {
		c.SetJson(0, "", "您没有权限！")
		return
	}

	userid := c.GetString("userid")
	role := c.GetString("role")

	if role == "" {
		c.SetJson(0, "", "角色类型为空！")
		return
	}
	if role != "admin" && role != "deploy-global" {
		c.SetJson(0, "", "新增角色类型只能为管理员或全局部署人员！")
		return
	}
	if role == "admin" {
		if strings.Contains(c.Role,"super-admin") == false {
			c.SetJson(0, "", "您没有权限！")
			return
		}
	}

	var uRole models.UserRole
	var err error

	err = initial.DB.Model(&models.UserRole{}).Where("username = ?",userid).First(&uRole).Error

	if err == nil {
		if uRole.IsDelete == 0 {
			c.SetJson(0, "", "用户角色已存在！")
			return
		}

		if uRole.IsDelete == 1 {
			inster := time.Now().Format("2006-01-02 15:04:05")
			tx := initial.DB.Begin()
			if err = tx.Model(&uRole).Updates(map[string]interface{}{"is_delete": 0, "insert_time": inster, "role": role}).Error ; err != nil {
				tx.Rollback()
				c.SetJson(0,"", err.Error())
				return
			}
			if err = tx.Commit().Error ; err != nil {
				tx.Rollback()
				c.SetJson(0,"",err.Error())
				return
			}
		}
	}else if err == gorm.ErrRecordNotFound {
		var uLogin models.UserLogin
		if err = initial.DB.Model(&models.UserLogin{}).Where("userid = ?",userid).First(&uLogin).Error ; err != nil {
			c.SetJson(0,"",err.Error())
			return
		}
		uRole.Username = userid
		uRole.Role = role
		uRole.Realname = uLogin.UserName
		uRole.Email = uLogin.Email
		uRole.InsertTime = time.Now().Format("2006-01-02 15:04:05")

		tx := initial.DB.Begin()
		if err = tx.Create(&uRole).Error ; err != nil {
			tx.Rollback()
			c.SetJson(0,"",err.Error())
			return
		}
		if err = tx.Commit().Error ; err != nil {
			tx.Rollback()
			c.SetJson(0,"",err.Error())
			return
		}
	}else {
		c.SetJson(0, "", err.Error())
		return
	}

	c.SetJson(1,"","角色添加成功！")
}

// @Title OperationRole
// @Description 修改或删除某用户的角色
// @Param	userid	query	string	true	"用户id"
// @Param	type	query	string	true	"修改或删除: alter,del"
// @Param	role	query	string	true	"用户角色: admin/deploy-global;"
// @Success 200 true or false
// @Failure 403
// @router /role/operation [get]
func (c *RoleController) OperationRole() {
	if !(strings.Contains(c.Role,"admin") || strings.Contains(c.Role,"super-admin")){
		c.SetJson(0,"","您没有权限！")
		return
	}

	userid := c.GetString("userid")
	role := c.GetString("role")
	op := c.GetString("type")

	if role == "" {
		c.SetJson(0,"","角色类型为空！")
		return
	}
	if role != "admin" && role != "deploy-global" {
		c.SetJson(0,"","操作对象只能为管理员和全局部署人员！")
		return
	}

	if role == "admin" {
		if c.Role != "super-admin" {
			c.SetJson(0,"","您没有权限！")
			return
		}
	}
	var uRole models.UserRole
	var err error
	if err = initial.DB.Model(&models.UserRole{}).Where("is_delete = 0 AND username = ?",userid).First(&uRole).Error ; err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	if uRole.Role == "admin" {
		if c.Role != "super-admin" {
			c.SetJson(0,"","您没有权限！")
			return
		}
	}

	msg := ""
	switch op {
	case "alter":
		uRole.Role = role
		msg = "角色修改成功！"
	case "del":
		uRole.IsDelete = 1
		msg = "角色删除成功！"
	}

	tx := initial.DB.Begin()
	if err = tx.Save(&uRole).Error ; err != nil {
		tx.Rollback()
		c.SetJson(0,"",err.Error())
		return
	}
	if err = tx.Commit().Error ; err != nil {
		tx.Rollback()
		c.SetJson(0,"",err.Error())
		return
	}

	c.SetJson(1,"", msg)
}
