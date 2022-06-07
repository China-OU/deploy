package itest

import (
	"controllers"
	"models"
	"initial"
	"fmt"
)

type ClearLogUserController struct {
	controllers.BaseController
}

func (c *ClearLogUserController) URLMapping() {
	c.Mapping("ClearUser", c.ClearUser)
}

// @Title 清除无效用户
// @Description 清除无效用户
// @Param	opr_type	query	string	false	"操作类型"
// @Success 200 {object} {}
// @Failure 403
// @router /clear/user [get]
func (c *ClearLogUserController) ClearUser() {
	if c.Role != "super-admin" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	opr_type := c.GetString("opr_type")
	clear_cnt := 0
	clear_name := ""
	var user []models.UserLogin
	initial.DB.Model(models.UserLogin{}).Find(&user)
	for _, v := range user {
		var cnt int
		err := initial.DB.Table("user_nuc").Where("username=?", v.Userid).Count(&cnt).Error
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		if cnt == 0 {
			clear_cnt += 1
			clear_name += fmt.Sprintf("%s(%s), ", v.UserName, v.Userid)
			if opr_type == "clear" {
				DeleteInvalidUser(v.Userid)
			}
		}
	}

	ret := map[string]interface{}{
		"cnt": clear_cnt,
		"data": clear_name,
	}
	if opr_type == "clear" {
		c.SetJson(1, ret, "无效用户已清除！")
		return
	}
	c.SetJson(1, ret, "无效用户查询成功！")
}

func DeleteInvalidUser(userid string)  {
	tx := initial.DB.Begin()
	err := tx.Where("userid=?", userid).Delete(&models.UserLogin{}).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}
