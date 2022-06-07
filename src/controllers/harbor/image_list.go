package harbor

import (
	"models"
	"strings"
	"fmt"
	"initial"
	"github.com/astaxie/beego"
	"encoding/json"
	"library/common"
	"github.com/astaxie/beego/httplib"
	"time"
)

// @Title ImageList
// @Description 获取镜像同步列表
// @Param	result	query	string	false	"结果，0/1/2/10代表不同结果"
// @Param	image	query	string	false	"镜像名，支持模糊搜索"
// @Param	user	query	string	false	"用户，cpds/其它用户/全部"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} models.HarborSync
// @Failure 403
// @router /image/list [get]
func (c *HarborOprController) ImageList() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	result := c.GetString("result")
	image := c.GetString("image")
	user := c.GetString("user")
	page, rows := c.GetPageRows()
	cond := " is_delete=0 "
	if strings.TrimSpace(result) != "" && strings.TrimSpace(result) != "100" {
		cond += fmt.Sprintf(" and result='%s' ", result)
	}
	if strings.TrimSpace(image) != "" {
		cond += fmt.Sprintf(" and image_url like '%%%s%%' ", image)
	}
	if strings.TrimSpace(user) != "" {
		if user == "cpds" {
			cond += fmt.Sprintf(" and operator = 'cpds' ")
		}
		if user == "other" {
			cond += fmt.Sprintf(" and operator != 'cpds' ")
		}
	}

	var cnt int
	var ulist []models.HarborSync
	err := initial.DB.Model(models.HarborSync{}).Where(cond).Count(&cnt).Order("id desc").Offset((page - 1)*rows).
		Limit(rows).Find(&ulist).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": ulist,
	}
	c.SetJson(1, ret, "harbor镜像同步列表获取成功！")
}

// @Title ImageAdd
// @Description 镜像同步新增信息
// @Param	image_url	query	string	true	"uat镜像地址"
// @Success 200 true or false
// @Failure 403
// @router /image/add [post]
func (c *HarborOprController) ImageAdd() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	image_url := c.GetString("image_url")

	// 校验测试环境harbor镜像是否存在
	req := httplib.Get("http://100.70.42.52/mdeploy/v1/ext/harbor/check")
	req.Header("Authorization", "mdeploy_IpFhvFjiQpV65PjIUywc3VHDjC0Wo9EM")
	req.Param("image", image_url)
	ret, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	img_check := make(map[string]interface{})
	err = json.Unmarshal([]byte(ret), &img_check)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if common.GetString(img_check["code"]) != "1" {
		c.SetJson(0, "", "harbor-uat不存在该镜像！")
		return
	}

	var is models.HarborSync
	is.ImageUrl = image_url
	is.Result = 10
	is.Message = ""
	is.CostTime = 0
	is.ApplyPerson = c.UserId
	is.Operator = c.UserId
	is.InsertTime = time.Now().Format(initial.DatetimeFormat)
	is.IsDelete = 0
	tx := initial.DB.Begin()
	err = tx.Create(&is).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "同步镜像添加成功！")
}

// @Title ImageDel
// @Description 删除镜像同步信息
// @Param	id	query	string	true	"镜像同步的id"
// @Success 200 true or false
// @Failure 403
// @router /image/del [post]
func (c *HarborOprController) ImageDel() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	id := c.GetString("id")
	var data models.HarborSync
	err := initial.DB.Model(models.HarborSync{}).Where("id=?", id).First(&data).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if data.Result != 10 {
		c.SetJson(0, "", "只有未开始的数据才能删除！")
		return
	}
	tx := initial.DB.Begin()
	err = tx.Model(models.HarborSync{}).Where("id=?", id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "同步镜像删除成功！")
}
