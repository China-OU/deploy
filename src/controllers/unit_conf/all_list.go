package unit_conf

import (
	"controllers"
	"strings"
	"fmt"
	"initial"
	"models"
	"github.com/astaxie/beego"
	"library/cfunc"
)

type UnitConfListController struct {
	controllers.BaseController
}

func (c *UnitConfListController) URLMapping() {
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Sync", c.Sync)
	c.Mapping("Add", c.Add)
	c.Mapping("Del", c.Del)
	c.Mapping("Copy", c.Copy)
}

// GetAll 方法
// @Title Get All
// @Description 获取所有发布单元列表
// @Param	en_name	query	string	false	"发布单元英文名，支持模糊搜索"
// @Param	zh_name	query	string	false	"发布单元中文名，支持模糊搜索"
// @Param	leader	query	string	false	"系统负责人，支持模糊搜索"
// @Param	dumd_en_name	query	string	false	"dumd子系统英文名，支持模糊搜索"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} models.UnitConfList
// @Failure 403
// @router /all [get]
func (c *UnitConfListController) GetAll() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	en_name := c.GetString("en_name")
	zh_name := c.GetString("zh_name")
	leader := c.GetString("leader")
	dumd_en_name := c.GetString("dumd_en_name")
	page, rows := c.GetPageRows()
	cond := " is_offline=0 "
	if strings.TrimSpace(en_name) != "" {
		cond += fmt.Sprintf(" and unit like '%%%s%%' ", en_name)
	}
	if strings.TrimSpace(zh_name) != "" {
		cond += fmt.Sprintf(" and name like '%%%s%%' ", zh_name)
	}
	if leader != "" {
		cond += fmt.Sprintf(" and leader = '%s' ", leader)
	}
	if strings.TrimSpace(dumd_en_name) != "" {
		cond += fmt.Sprintf(" and dumd_sub_sysname like '%%%s%%' ", dumd_en_name)
	}
	var cnt int
	var ulist []models.UnitConfList
	err := initial.DB.Model(models.UnitConfList{}).Where(cond).Count(&cnt).Order("id desc").Offset((page - 1)*rows).Limit(rows).Find(&ulist).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	// 组装数据
	type UnitConfDetail struct {
		models.UnitConfList
		LeaderName   string    `json:"leader_name"`
		TestName     string    `json:"test_name"`
		DeveloperName    string    `json:"developer_name"`
		CompName     string    `json:"comp_name"`
		TypeName     string    `json:"type_name"`
	}
	var data_ret []UnitConfDetail
	for _, v := range ulist {
		per := UnitConfDetail{
			v,
			cfunc.GetUserCnName(v.Leader),
			cfunc.GetUserCnName(v.Test),
			cfunc.GetUserCnName(v.Developer),
			cfunc.GetCompCnName(v.DumdCompEn),
			cfunc.GetTypeCnName(v.AppType),
		}
		data_ret = append(data_ret, per)
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": data_ret,
	}
	c.SetJson(1, ret, "数据获取成功！")
}

// @Description 删除发布单元接口
// @Param	id	query	string	true	"发布单元的id"
// @Success 200  true or false
// @Failure 403
// @router /allunit/del [post]
func (c *UnitConfListController) Del() {
	id := c.GetString("id")
	var data models.UnitConfList
	err := initial.DB.Model(models.UnitConfList{}).Where("id=?", id).First(&data).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if c.UserId != data.Operator {
		c.SetJson(0, "", "此发布单元不是您创建的，您没有删除权限！")
		return
	}
	tx := initial.DB.Begin()
	err = tx.Model(models.UnitConfList{}).Where("id=?", id).Update("is_offline", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "发布单元删除成功！")
}