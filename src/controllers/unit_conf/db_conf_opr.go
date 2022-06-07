package unit_conf

import (
	"models"
	"encoding/json"
	"controllers"
	"initial"
	"library/git"
	"fmt"
	"library/common"
	"regexp"
	"time"
	"github.com/astaxie/beego"
	"strings"
)

// DB发布单元录入
// @Title DB发布单元录入
// @Description 从发布单元列表选取数据，同时作相关信息确认和维护
// @Param	body	body	models.UnitConfDb	true	"body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /db/save [post]
func (c *DBConfListController) DBSave() {
	if c.Role == "guest" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	if beego.AppConfig.String("runmode") == "prd" && strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "生产环境权限收缩，您没有权限操作！")
		return
	}
	var db_info models.UnitConfDb
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &db_info)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(db_info.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的编辑权限，请联系发布单元负责人进行操作！")
		return
	}

	// 校验
	db_info.Host = strings.TrimSpace(db_info.Host)
	db_info.Dbname = strings.TrimSpace(db_info.Dbname)
	db_info.Username = strings.TrimSpace(db_info.Username)
	flag, msg := DBInfoCheck(db_info)
	if flag == false {
		c.SetJson(0, "", msg)
		return
	}
	var cnt int
	initial.DB.Model(models.UnitConfDb{}).Where("id != ? and unit_id = ? and is_delete = 0", db_info.Id, db_info.UnitId).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "此发布单元已经创建，不能重复创建。如果需要创建备份发布单元，请在主列表页录入新发布单元！")
		return
	}

	// 录入或者更新
	tx := initial.DB.Begin()
	input_id := db_info.Id
	if db_info.Id > 0 {
		// 只更新五个字段
		update_map := map[string]interface{}{
			"type": db_info.Type,
			"git_id": db_info.GitId,
			"git_unit": db_info.GitUnit,
			"git_url": db_info.GitUrl,
			"deploy_comp": db_info.DeployComp,
			"username": db_info.Username,
			"host": db_info.Host,
			"port": db_info.Port,
			"dbname": db_info.Dbname,
			"operator": c.UserId,
		}
		if db_info.EncryPwd != "" && common.WebPwdDecrypt(db_info.EncryPwd) != "" {
			update_map["encry_pwd"] = common.AesEncrypt(common.WebPwdDecrypt(db_info.EncryPwd))
			RecordDatabasePwd(db_info.Id)
		}
		err = tx.Model(models.UnitConfDb{}).Where("id=?", db_info.Id).Updates(update_map).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	} else {
		if db_info.EncryPwd == "" || common.WebPwdDecrypt(db_info.EncryPwd) == "" {
			c.SetJson(0, "", "数据库密码不能为空！")
			return
		}
		db_info.InsertTime = time.Now().Format(initial.DatetimeFormat)
		db_info.EncryPwd = common.AesEncrypt(common.WebPwdDecrypt(db_info.EncryPwd))
		db_info.IsDelete = 0
		db_info.Operator = c.UserId
		db_info.ConnResult = 10
		err = tx.Create(&db_info).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	}
	tx.Commit()
	ret_msg := "数据库信息新增成功！"
	if input_id > 0 {
		ret_msg = "数据库信息维护成功！"
	}
	c.SetJson(1, "", ret_msg)
}

func DBInfoCheck(info models.UnitConfDb) (bool, string) {
	// git信息确认
	git_info := git.SearchByGitId(info.GitId)
	if info.GitUnit != git_info.PathWithNamespace {
		return false, "发布单元路径不对！"
	}
	if info.GitUrl != git_info.HTTPURLToRepo {
		return false, "发布单元url不对！"
	}
	if common.InList(info.Type, []string{"oracle", "mysql", "pgsql"})  == false {
		return false, "数据库目前只支持oracle、mysql和pgsql ！"
	}
	// 检查发布单元是否正确
	var ul models.UnitConfList
	err := initial.DB.Model(models.UnitConfList{}).Where("is_offline=0 and id = ?", info.UnitId).First(&ul).Error
	if err != nil {
		return false, err.Error()
	}
	//if ul.GitUnit != "" && ul.GitUnit != "<nil>" && ul.GitUnit != info.GitUnit {
	//	return false, "发布单元信息不匹配，请重新填写！"
	//}

	// 检验数据库ip和端口合法性，ip可能有域名暂不校验
	if common.CheckIp(info.Host) == false {
		mStr :="^[a-zA-Z.]+\\.cmftdc.cn"
		if m, _ := regexp.Match(mStr, []byte(info.Host)); !m {
			return false, "数据库ip或者域名不正确！"
		}
	}
	if info.Port < 1024 || info.Port > 65536 {
		return false, "端口值错误！"
	}
	// 检查网络区域是否正确
	net_flag := CheckBasicInfo(fmt.Sprintf("dumd_comp_en = '%s'", info.DeployComp))
	if !net_flag {
		return false, "部署租户不正确！"
	}
	return true, ""
}

// 记录当前数据库信息，以免密码修改错误
func RecordDatabasePwd(id int)  {
	var data models.UnitConfDb
	err := initial.DB.Model(models.UnitConfDb{}).Where("id = ? and is_delete = 0", id).First(&data).Error
	if err != nil {
		beego.Error(err.Error())
		return
	}

	var cnt int
	initial.DB.Model(models.RcDbPwd{}).Where("db_conf_id=? and username=? and encry_pwd=? and host=? and " +
		"port=? and dbname=?", data.Id, data.Username, data.EncryPwd, data.Host, data.Port, data.Dbname).Count(&cnt)
	if cnt > 0 {
		return
	}

	var rc models.RcDbPwd
	rc.DbConfId = data.Id
	rc.Username = data.Username
	rc.EncryPwd = data.EncryPwd
	rc.Host = data.Host
	rc.Port = data.Port
	rc.Dbname = data.Dbname
	rc.Operator = data.Operator
	rc.InsertTime = data.InsertTime
	tx := initial.DB.Begin()
	err = tx.Create(&rc).Error
	if err != nil {
		beego.Error(err.Error())
		return
	}
	tx.Commit()
}
