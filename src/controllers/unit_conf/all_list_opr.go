package unit_conf

import (
	"time"
	"library/common"
	"library/datasession"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"models"
	"encoding/json"
	"initial"
	"strings"
	"fmt"
	"regexp"
)

// Sync 同步
// @Title 发布单元同步接口
// @Description 从发布管理系统同步发布单元基础数据
// @Success 200 true or false
// @Failure 403
// @router /all/sync [post]
func (c *UnitConfListController) Sync() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	if common.InList(beego.AppConfig.String("runmode"), []string{"dev", "prd"}) == false {
		c.SetJson(0, "", "只有生产环境和开发环境才能拉取发布单元数据！")
		return
	}
	last_time, flag := datasession.PmstUnitSyncTime()
	if time.Now().Add(- 500 * time.Second).Format("2006-01-02 15:04:05") < common.GetString(last_time) && flag == 1 {
		c.SetJson(0, "", "发布单元500秒内只能同步一次，上次同步时间：" + common.GetString(last_time))
		return
	}
	// 同步发布单元基础信息
	req := httplib.Get(beego.AppConfig.String("pms_baseurl") + "/mdp/unit/sync")
	req.Header("Authorization", "Basic mdeploy_d8c8680d046b1c60e63657deb3ce6d89")
	req.Header("Content-Type", "application/json")
	info_byte, err := req.Bytes()
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	type ReqData struct {
		Data []models.UnitConfList `json:"data"`
	}
	var ret ReqData
	err = json.Unmarshal(info_byte, &ret)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	tx := initial.DB.Begin()
	for _, v := range ret.Data {
		cnt := 0
		v.Operator = "pms"
		tx.Model(models.UnitConfList{}).Where("unit = ?", v.Unit).Count(&cnt)
		if cnt > 0 {
			err = tx.Model(models.UnitConfList{}).Where("unit=?", v.Unit).Update(v).Error
			if err != nil {
				tx.Rollback()
				c.SetJson(0, "", err.Error())
				return
			}
		} else {
			err = tx.Create(&v).Error
			if err != nil {
				tx.Rollback()
				c.SetJson(0, "", err.Error())
				return
			}
		}
	}
	tx.Commit()
	c.SetJson(1, "", "数据增量同步成功！")
}

// @Title 管理员新增发布单元
// @Description 管理员新增发布单元
// @Param	body	body	unit_conf.UnitInput	true	"body形式的数据，发布单元信息"
// @Success 200  true or false
// @Failure 403
// @router /allunit/add [post]
func (c *UnitConfListController) Add()  {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作，只有管理员才能新增发布单元！")
		return
	}

	var unit UnitInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &unit)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 数据校验
	unit.EnName = strings.Trim(unit.EnName, " ")
	ok, _ := regexp.MatchString("^[a-z0-9-]{3,50}$", unit.EnName)
	if !ok {
		c.SetJson(0, "", "发布单元必须是英文名！")
		return
	}
	if strings.Contains(unit.EnName, " ") {
		c.SetJson(0, "", "发布单元必须是英文名！")
		return
	}
	var cnt int
	err = initial.DB.Model(models.UnitConfList{}).Where("unit=? and is_offline=0", unit.EnName).Count(&cnt).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if cnt > 0 {
		c.SetJson(0, "", "不能重复添加发布单元！")
		return
	}
	var ind models.UnitConfList
	ind.Unit = unit.EnName
	ind.Name = unit.CnName
	ind.Info = fmt.Sprintf("%s(%s)", unit.CnName, unit.EnName)
	ind.Leader = unit.Leader
	ind.Developer = unit.Developer + ","
	ind.Test = unit.Testopr + ","
	ind.InsertTime = time.Now().Format(initial.DatetimeFormat)
	ind.Operator = c.UserId

	tx := initial.DB.Begin()
	err = tx.Create(&ind).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "发布单元添加成功！")
}

type UnitInput struct {
	EnName   string    `json:"en_name"`
	CnName   string    `json:"cn_name"`
	Leader   string    `json:"leader"`
	Developer    string   `json:"developer"`
	Testopr      string   `json:"testopr"`
}

// @Title 负责人复制发布单元
// @Description 管理员复制发布单元
// @Param	body	body	unit_conf.UnitCopyInput	true	"body形式的数据，发布单元复制信息"
// @Success 200  true or false
// @Failure 403
// @router /allunit/copy [post]
func (c *UnitConfListController) Copy() {
	var unit UnitCopyInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &unit)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	unit.EnName = strings.Trim(unit.EnName, " ")
	ok, _ := regexp.MatchString("^[a-z0-9-]{3,50}$", unit.EnName)
	if !ok {
		c.SetJson(0, "", "发布单元必须是英文名！")
		return
	}
	if strings.Contains(unit.EnName, " ") {
		c.SetJson(0, "", "发布单元必须是英文名！")
		return
	}
	var info models.UnitConfList
	err = initial.DB.Model(models.UnitConfList{}).Where("id=? and is_offline=0", unit.CopyId).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	var cnt int
	err = initial.DB.Model(models.UnitConfList{}).Where("unit=? and is_offline=0", unit.EnName).Count(&cnt).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if cnt > 0 {
		c.SetJson(0, "", "不能重复添加发布单元！")
		return
	}
	if info.Leader != c.UserId {
		c.SetJson(0, "", "只有系统负责人才有权限复制！")
		return
	}
	tx := initial.DB.Begin()
	info.Unit = unit.EnName
	info.Name = unit.CnName
	info.Info = fmt.Sprintf("%s(%s)", unit.CnName, unit.EnName)
	info.Operator = c.UserId
	info.Id = 0
	err = tx.Create(&info).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "发布单元复制成功！")
}

type UnitCopyInput struct {
	EnName   string    `json:"en_name"`
	CnName   string    `json:"cn_name"`
	CopyId   string    `json:"copy_id"`
}