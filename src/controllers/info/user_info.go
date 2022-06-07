package info

import (
	"controllers"
	"initial"
	"models"
	"github.com/astaxie/beego"
	"library/datasession"
	"time"
	"library/common"
	"encoding/json"
	"github.com/astaxie/beego/httplib"
	"github.com/jinzhu/gorm"
)

type UserInfoController struct {
	controllers.BaseController
}

func (c *UserInfoController) URLMapping() {
	c.Mapping("SyncUserFromPms", c.SyncUserFromPms)
	c.Mapping("SearchUser", c.SearchUser)
}

// @Title SyncUserFromPms
// @Description 从发布管理系统同步用户登录信息，保证系统初始时人员下拉可用。
// @Success 200  true or false
// @Failure 403
// @router /user/sync [post]
func (c *UserInfoController) SyncUserFromPms() {
	last_time, flag := datasession.PmstUserSyncTime()
	if time.Now().Add(- 600 * time.Second).Format(initial.DatetimeFormat) < common.GetString(last_time) && flag == 1 {
		c.SetJson(0, "", "发布单元600秒内只能同步一次，上次同步时间：" + common.GetString(last_time))
		return
	}

	// 同步发布单元基础信息
	req := httplib.Get(beego.AppConfig.String("pms_baseurl") + "/staff/login")
	req.Header("Content-Type", "application/json")
	info_byte, err := req.Bytes()
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	type StaffLogin struct {
		Userid     string    `json:"userid"`
		UserName   string    `json:"zh_name"`
		Phone      string    `json:"telephone"`
		Email      string    `json:"email"`
		Title      string    `json:"title"`
		Department     string    `json:"depart"`
	}
	type ReqData struct {
		Data []StaffLogin `json:"data"`
	}
	var ret ReqData
	err = json.Unmarshal(info_byte, &ret)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	beego.Info(ret.Data[0].Userid)
	tx := initial.DB.Begin()
	for _, v := range ret.Data {
		cnt := 0
		tx.Model(models.UserLogin{}).Where("userid = ?", v.Userid).Count(&cnt)
		if cnt > 0 {
			continue
		}
		var ul models.UserLogin
		ul.Userid = v.Userid
		ul.UserName = v.UserName
		ul.Email = v.Email
		ul.Department = common.CheckNil(v.Department)
		ul.Title = common.CheckNil(v.Title)
		ul.Phone = common.CheckNil(v.Phone)
		ul.InsertTime = time.Now().Format(initial.DatetimeFormat)
		err = tx.Create(&ul).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	}
	tx.Commit()
	c.SetJson(1, "", "用户数据同步成功！")
}

// @Title SearchUser
// @Description 根据花名或中文名进行模糊搜索，获取用户的下拉列表
// @Param	search	query	string	true	"搜索内容"
// @Success 200 {object} []models.CaasCluster
// @Failure 403
// @router /user/search [get]
func (c *UserInfoController) SearchUser() {
	search := c.GetString("search")
	var user_list []models.UserLogin
	err := initial.DB.Model(models.UserLogin{}).Where("userid like ? or username like ?", "%" + search + "%",
		"%" + search + "%").Limit(10).Find(&user_list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		c.SetJson(0, "", err.Error())
		return
	}
	c.SetJson(1, user_list, "用户搜索成功！")
}

