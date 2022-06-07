package online

import (
	"controllers"
	"initial"
	"github.com/astaxie/beego"
	"time"
	"github.com/astaxie/beego/httplib"
	"library/datasession"
	"library/common"
	"encoding/json"
	"library/git"
	"models"
	"strings"
	"errors"
	"fmt"
)

type ReleaseRecordController struct {
	controllers.BaseController
}

func (c *ReleaseRecordController) URLMapping() {
	c.Mapping("PmsReleaseRecord", c.PmsReleaseRecord)
	c.Mapping("QueryReleaseRecord", c.QueryReleaseRecord)
}

// @Title 从发布管理系统获取发布列表
// @Description 从发布管理系统获取发布列表
// @Param	online_date	  query	string	false	"上线日期，如20191001"
// @Success 200  true or false
// @Failure 403
// @router /record/pms [get]
func (c *ReleaseRecordController) PmsReleaseRecord() {
	last_time, flag := datasession.ReleaseRecordSyncTime()
	if time.Now().Add(- 100 * time.Second).Format(initial.DatetimeFormat) < common.GetString(last_time) && flag == 1 {
		c.SetJson(0, "", "发布单元100秒内只能同步一次，上次同步时间：" + common.GetString(last_time))
		return
	}
	if c.Role == "guest" || c.Role == "deploy-single" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	if beego.AppConfig.String("runmode") != "prd" {
		c.SetJson(0, "", "只有生产环境才能拉取！")
		return
	}

	online_date := c.GetString("online_date")
	// 如果不传参，拉当天的；如果0~6点，拉前一天的；
	if online_date == "" {
		online_date = time.Now().Format(initial.DateFormat)
	}
	now := time.Now().Format("15:04:05")
	if  now > "00:00:00" && now < "06:00:00" {
		d, err := time.Parse(initial.DateFormat, online_date)
		if err != nil {
			c.SetJson(0, "", "日期格式错误，"+err.Error())
			return
		}
		online_date = d.AddDate(0, 0, -1).Format(initial.DateFormat)
	}

	// 从pms拉取当天版本
	beego.Info("online_date is ", online_date)
	req := httplib.Get(beego.AppConfig.String("pms_baseurl") + "/mdp/release/record")
	req.Header("Authorization", "Basic mdeploy_d8c8680d046b1c60e63657deb3ce6d89")
	req.Header("Content-Type", "application/json")
	req.Param("online_date", online_date)
	info_byte, err := req.Bytes()
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	beego.Info(string(info_byte))
	type ReqData struct {
		Data []PmsRelData `json:"data"`
	}
	var ret ReqData
	err = json.Unmarshal(info_byte, &ret)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	for _, v := range ret.Data {
		utype, unit_id := GetUnitType(v.Unit)
		if utype == "nomatch" {
			continue
		}
		if utype == "cntr" {
			err = CntrRecordInsert(v, unit_id)
			if err != nil {
				c.SetJson(0, "", err.Error())
				return
			}
		}
		if utype == "vm" {
			err = VmRecordInsert(v, unit_id)
			if err != nil {
				c.SetJson(0, "", err.Error())
				return
			}
		}
		if utype == "db" {
			err = DbRecordInsert(v, unit_id)
			if err != nil {
				c.SetJson(0, "", err.Error())
				return
			}
		}
	}

	c.SetJson(1, "", online_date + "版本数据拉取成功！")
}

type PmsRelData struct {
	Unit        string   `json:"unit"`
	Branch      string   `json:"branch"`
	Sha         string   `json:"sha"`
	OnlineDate  string   `json:"online_date"`
	OnlineTime  string   `json:"online_time"`
	RecordId    string   `json:"record_id"`
}

func GetUnitType(unit string) (string, int) {
	unit_list := strings.Split(unit, "(")
	unit_en := unit_list[len(unit_list)-1]
	unit_en = strings.Split(unit_en, ")")[0]
	// 获取unit_id
	var info models.UnitConfList
	err := initial.DB.Model(models.UnitConfList{}).Where("unit=? and is_offline=0", unit_en).First(&info).Error
	if err != nil {
		beego.Error(err.Error())
		return "nomatch", 0
	}

	// 判断标准容器
	var cnt int
	var cntr models.UnitConfCntr
	initial.DB.Model(models.UnitConfCntr{}).Where("unit_id=? and is_delete=0 and cpds_flag=0", info.Id).First(&cntr).Count(&cnt)
	if cnt > 0 {
		// 只更新镜像的不拉取
		if cntr.IsConfirm != 1 {
			return "nomatch", 0
		}
		return "cntr", info.Id
	}

	// 判断标准虚机
	var cnt_vm int
	var conf_vm models.UnitConfVM
	initial.DB.Model(models.UnitConfVM{}).Where("unit_id=? and is_delete=0", info.Id).First(&conf_vm).Count(&cnt_vm)
	if cnt_vm > 0 {
		if conf_vm.IsConfirm != 1 {
			return "nomatch", 0
		}
		return "vm", info.Id
	}

	// 判断数据库
	var cnt_db int
	var conf_db models.UnitConfDb
	initial.DB.Model(models.UnitConfDb{}).Where("unit_id=? and is_delete=0", info.Id).First(&conf_db).Count(&cnt_db)
	if cnt_db > 0 {
		if conf_db.ConnResult != 1 {
			return "nomatch", 0
		}
		return "db", info.Id
	}

	return "nomatch", 0
}

func CntrRecordInsert(v PmsRelData, unit_id int) error {
	now := time.Now().Format(initial.DatetimeFormat)
	// 赋初始值
	var online_main models.OnlineAllList
	var online_cntr models.OnlineStdCntr
	online_main.UnitId = unit_id
	online_main.Branch = v.Branch
	//online_main.CommitId = ""
	//online_main.ShortCommitId = ""
	online_main.OnlineDate = v.OnlineDate
	online_main.OnlineTime = v.OnlineTime
	online_main.Version = v.OnlineDate
	online_main.IsProcessing = 0
	online_main.IsSuccess = 10   // 10表示未开始
	online_main.IsDelete = 0
	online_main.Operator = "pms"
	online_main.ExcuteTime = ""
	online_main.InsertTime = now
	online_main.ErrorLog = ""
	online_main.SourceId = v.RecordId

	//online_cntr.OnlineId = 0
	online_cntr.JenkinsSuccess = 10
	online_cntr.IsDelete = 0
	online_cntr.InsertTime = now

	var cntr_conf models.UnitConfCntr
	initial.DB.Model(models.UnitConfCntr{}).Where("is_delete=0 and unit_id=?", unit_id).First(&cntr_conf)
	commit_detail := git.GetCommitDetail(cntr_conf.GitId, v.Sha)
	if commit_detail == nil {
		return errors.New(fmt.Sprintf("发布单元%s错误，或者sha值错误", cntr_conf.GitUnit))
	}
	online_main.CommitId = commit_detail.ID
	online_main.ShortCommitId = commit_detail.ShortID

	// 同一发布单元同一sha值，不允许重复录入
	var cnt int
	initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and commit_id=? and  online_date=? and is_delete=0",
		unit_id, v.Branch, commit_detail.ID, v.OnlineDate).Count(&cnt)
	if cnt > 0 {
		beego.Info("已录入！")
		return nil
	}
	// 如果sha值更新，未发布的发布单元需要更新sha值，已发布或者发布中的，不能更新sha值
	cnt = 0
	initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and online_date=? and is_delete=0 and is_success=10",
		unit_id, v.Branch, v.OnlineDate).Count(&cnt)
	if cnt > 0 {
		tx := initial.DB.Begin()
		update_map := map[string]interface{}{
			"commit_id": commit_detail.ID,
			"short_commit_id": commit_detail.ShortID,
		}
		err := initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and online_date=? and is_delete=0 and is_success=10",
			unit_id, v.Branch, v.OnlineDate).Updates(update_map).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	}

	// 录入
	tx := initial.DB.Begin()
	err := tx.Create(&online_main).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	online_cntr.OnlineId = online_main.Id
	err = tx.Create(&online_cntr).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func VmRecordInsert(v PmsRelData, unit_id int) error {
	now := time.Now().Format(initial.DatetimeFormat)
	// 赋初始值
	var online_main models.OnlineAllList
	var online_vm models.OnlineStdVM
	online_main.UnitId = unit_id
	online_main.Branch = v.Branch
	//online_main.CommitId = ""
	//online_main.ShortCommitId = ""
	online_main.OnlineDate = v.OnlineDate
	online_main.OnlineTime = v.OnlineTime
	online_main.Version = v.OnlineDate
	online_main.IsProcessing = 0
	online_main.IsSuccess = 10   // 10表示未开始
	online_main.IsDelete = 0
	online_main.Operator = "pms"
	online_main.ExcuteTime = ""
	online_main.InsertTime = now
	online_main.ErrorLog = ""
	online_main.SourceId = v.RecordId

	//online_cntr.OnlineId = 0
	online_vm.CreateTime = now
	online_vm.BuildStatus = 10
	online_vm.UpgradeStatus = 10
	online_vm.UpgradeDuration = 0
	online_vm.IsDelete = 0

	var vm_conf models.UnitConfVM
	initial.DB.Model(models.UnitConfVM{}).Where("is_delete=0 and unit_id=?", unit_id).First(&vm_conf)
	commit_detail := git.GetCommitDetail(vm_conf.GitID, v.Sha)
	if commit_detail == nil {
		return errors.New(fmt.Sprintf("发布单元%s错误，或者sha值错误", vm_conf.GitUnit))
	}
	online_main.CommitId = commit_detail.ID
	online_main.ShortCommitId = commit_detail.ShortID

	// 同一发布单元同一sha值，不允许重复录入
	var cnt int
	initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and commit_id=? and  online_date=? and is_delete=0",
		unit_id, v.Branch, commit_detail.ID, v.OnlineDate).Count(&cnt)
	if cnt > 0 {
		beego.Info("已录入！")
		return nil
	}
	// 如果sha值更新，未发布的发布单元需要更新sha值，已发布或者发布中的，不能更新sha值
	cnt = 0
	initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and online_date=? and is_delete=0 and is_success=10",
		unit_id, v.Branch, v.OnlineDate).Count(&cnt)
	if cnt > 0 {
		tx := initial.DB.Begin()
		update_map := map[string]interface{}{
			"commit_id": commit_detail.ID,
			"short_commit_id": commit_detail.ShortID,
		}
		err := initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and online_date=? and is_delete=0 and is_success=10",
			unit_id, v.Branch, v.OnlineDate).Updates(update_map).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	}

	// 录入
	tx := initial.DB.Begin()
	err := tx.Create(&online_main).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	online_vm.OnlineID = online_main.Id
	err = tx.Create(&online_vm).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func DbRecordInsert(v PmsRelData, unit_id int) error {
	now := time.Now().Format(initial.DatetimeFormat)
	// 赋初始值
	var online_main models.OnlineAllList
	var online_db models.OnlineDbList
	online_main.UnitId = unit_id
	online_main.Branch = v.Branch
	//online_main.CommitId = ""
	//online_main.ShortCommitId = ""
	online_main.OnlineDate = v.OnlineDate
	online_main.OnlineTime = v.OnlineTime
	online_main.Version = v.OnlineDate
	online_main.IsProcessing = 0
	online_main.IsSuccess = 10   // 10表示未开始
	online_main.IsDelete = 0
	online_main.Operator = "pms"
	online_main.ExcuteTime = ""
	online_main.InsertTime = now
	online_main.ErrorLog = ""
	online_main.SourceId = v.RecordId

	//online_cntr.OnlineId = 0
	online_db.IsDelete = 0
	online_db.DirName = ""
	online_db.IsDirClear = 0
	online_db.IsPullDir = 10

	var db_conf models.UnitConfDb
	initial.DB.Model(models.UnitConfDb{}).Where("is_delete=0 and unit_id=?", unit_id).First(&db_conf)
	commit_detail := git.GetCommitDetail(db_conf.GitId, v.Sha)
	if commit_detail == nil {
		return errors.New(fmt.Sprintf("发布单元%s错误，或者sha值错误", db_conf.GitUnit))
	}
	online_main.CommitId = commit_detail.ID
	online_main.ShortCommitId = commit_detail.ShortID

	// 同一发布单元同一sha值，不允许重复录入
	var cnt int
	initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and commit_id=? and  online_date=? and is_delete=0",
		unit_id, v.Branch, commit_detail.ID, v.OnlineDate).Count(&cnt)
	if cnt > 0 {
		beego.Info("已录入！")
		return nil
	}
	// 如果sha值更新，未发布的发布单元需要更新sha值，已发布或者发布中的，不能更新sha值
	cnt = 0
	initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and online_date=? and is_delete=0 and is_success=10",
		unit_id, v.Branch, v.OnlineDate).Count(&cnt)
	if cnt > 0 {
		tx := initial.DB.Begin()
		update_map := map[string]interface{}{
			"commit_id": commit_detail.ID,
			"short_commit_id": commit_detail.ShortID,
		}
		err := initial.DB.Model(models.OnlineAllList{}).Where("unit_id=? and branch=? and online_date=? and is_delete=0 and is_success=10",
			unit_id, v.Branch, v.OnlineDate).Updates(update_map).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	}

	// 录入
	tx := initial.DB.Begin()
	err := tx.Create(&online_main).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	online_db.OnlineId = online_main.Id
	err = tx.Create(&online_db).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}