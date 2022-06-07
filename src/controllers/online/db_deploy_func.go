package online

import (
	"models"
	"errors"
	"time"
	"github.com/astaxie/beego"
	"fmt"
	"initial"
	"library/common"
	"library/database"
	"strings"
	"github.com/astaxie/beego/httplib"
)

type DBDeploy struct {
	OnlineList    models.OnlineAllList
	Conf          models.UnitConfDb
	Operator      string
}

func (c *DBDeploy) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("DB Deploy Panic error:", err)
		}
	}()
	timeout := time.After(60 * time.Minute)
	if beego.AppConfig.String("runmode") == "di" || beego.AppConfig.String("runmode") == "st" {
		timeout = time.After(15 * time.Minute)
	}
	result_ch := make(chan bool, 1)
	go func() {
		result := c.GetDBData()
		// 更新发布管理系统结果
		go c.UpdatePmsResult(result)
		result_ch <- result
	}()
	select {
	case <-result_ch:
		beego.Info(fmt.Sprintf("数据库 %s 部署完成", c.Conf.Dbname))
	case <-timeout:
		beego.Info(fmt.Sprintf("数据库 %s 部署超时", c.Conf.Dbname))
		c.UpdateSumRecord(map[string]interface{}{"is_success": 0, "error_log": "数据库脚本执行超时！"})
	}
}

func (c *DBDeploy) GetDBData() bool {
	// 更新当前数据发布状态为发布中
	c.UpdateSumRecord(map[string]interface{}{"is_success": 2})

	// 执行顺序：ddl > trig > pkg_type > pkg_head > pkg_body > dml。所有数据库都是这个顺序
	err := c.ExecuteSql(c.ToDeployRecord("ddl"), "ddl", false)
	if err != nil {
		c.UpdateSumRecord(map[string]interface{}{"is_success": 0, "error_log": err.Error()})
		return false
	}

	err = c.ExecuteSql(c.ToDeployRecord("trig"), "trig", false)
	if err != nil {
		c.UpdateSumRecord(map[string]interface{}{"is_success": 0, "error_log": err.Error()})
		return false
	}

	err = c.ExecuteSql(c.ToDeployRecordById("pkg_type"), "pkg_type", false)
	if err != nil {
		c.UpdateSumRecord(map[string]interface{}{"is_success": 0, "error_log": err.Error()})
		return false
	}

	err = c.ExecuteSql(c.ToDeployRecordById("pkg_head"), "pkg_head", false)
	if err != nil {
		c.UpdateSumRecord(map[string]interface{}{"is_success": 0, "error_log": err.Error()})
		return false
	}

	err = c.ExecuteSql(c.ToDeployRecordById("pkg_body"), "pkg_body", false)
	if err != nil {
		c.UpdateSumRecord(map[string]interface{}{"is_success": 0, "error_log": err.Error()})
		return false
	}

	err = c.ExecuteSql(c.ToDeployRecord("dml"), "dml", false)
	if err != nil {
		c.UpdateSumRecord(map[string]interface{}{"is_success": 0, "error_log": err.Error()})
		return false
	}

	c.UpdateSumRecord(map[string]interface{}{"is_success": 1})
	return true
}

// 语句执行逻辑
func (c *DBDeploy) ExecuteSql(data []models.OnlineDbLog, db_type string, repeat bool) error {
	if len(data) == 0 {
		return nil
	}
	for _, v := range data {
		if v.IsSuccess == 1 && !repeat {
			continue
		}
		if v.IsSuccess == 2 {
			return errors.New(v.FileName + "正在运行中，禁止再次执行！")
		}
		if v.IsSuccess == 0 && db_type == "ddl" {
			return errors.New(v.FileName + "，执行失败，请进入详情单独执行！")
		}

		var err error
		var msg string
		start_time := time.Now()
		// 更改状态为执行中
		c.UpdateLogRecord(v.Id, map[string]interface{}{"is_success": 2})
		if c.Conf.Type == "mysql" {
			msg, err = database.MysqlExecSql(c.Conf, v)
		} else if c.Conf.Type == "oracle" {
			msg, err = database.OracleExecSql(c.Conf, v)
		} else if c.Conf.Type == "pgsql" {
			msg, err = database.PgsqlExecSQL(c.Conf, v)
		} else {
			c.UpdateLogRecord(v.Id, map[string]interface{}{"is_success": 10})
			return errors.New("暂不支持此类型的发布！")
		}
		umap := make(map[string]interface{})
		tmsg := msg
		if msg == "" {
			if err != nil {
				tmsg = err.Error()
			} else {
				tmsg = "语句执行成功"
			}
		}
		umap["message"] = tmsg
		umap["execute_time"] = common.GetInt(time.Now().Sub(start_time).Seconds())
		umap["start_time"] = start_time.Format(initial.DatetimeFormat)
		umap["operator"] = c.Operator

		if err == nil {
			umap["is_success"] = 1
			c.UpdateLogRecord(v.Id, umap)
		} else {
			umap["is_success"] = 0
			c.UpdateLogRecord(v.Id, umap)
			return err
		}
	}
	return nil
}

func (c *DBDeploy) ToDeployRecord(stype string) []models.OnlineDbLog {
	var data []models.OnlineDbLog
	initial.DB.Model(models.OnlineDbLog{}).Where("online_id=? and is_delete=0 and sql_type=? and file_sha=?",
		c.OnlineList.Id, stype, c.OnlineList.CommitId).Order("file_name asc").Find(&data)
	return data
}

func (c *DBDeploy) ToDeployRecordById(stype string) []models.OnlineDbLog {
	var data []models.OnlineDbLog
	initial.DB.Model(models.OnlineDbLog{}).Where("online_id=? and is_delete=0 and sql_type=? and file_sha=?",
		c.OnlineList.Id, stype, c.OnlineList.CommitId).Order("id asc").Find(&data)
	return data
}

func (c *DBDeploy) UpdateSumRecord(umap map[string]interface{}) {
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineAllList{}).Where("id=?", c.OnlineList.Id).Updates(umap).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (c *DBDeploy) UpdateLogRecord(id int, umap map[string]interface{}) {
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineDbLog{}).Where("id=?", id).Updates(umap).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (c *DBDeploy) UpdatePmsResult(result bool) {
	sid := strings.TrimSpace(c.OnlineList.SourceId)
	res := "0"
	if result {
		res = "1"
	}
	if sid != "" && sid != "0" {
		req := httplib.Get(beego.AppConfig.String("pms_baseurl") + "/mdp/release/result")
		req.Header("Authorization", "Basic mdeploy_d8c8680d046b1c60e63657deb3ce6d89")
		req.Header("Content-Type", "application/json")
		req.Param("record_id", sid)
		req.Param("result", res)
		_, err := req.String()
		if err != nil {
			beego.Error(err.Error())
		}
	}
}