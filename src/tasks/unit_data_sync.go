package tasks

import (
	"github.com/astaxie/beego"
	"runtime"
	"time"
	"models"
	"encoding/json"
	"initial"
	"github.com/astaxie/beego/httplib"
	"fmt"
	"library/common"
)

// 同步发布单元信息
func UnitDataSync() error {
	//获取panic
	defer func() {
		if panic_err := recover(); panic_err != nil {
			var buf []byte = make([]byte, 1024)
			c := runtime.Stack(buf, false)
			beego.Error("UnitDataSync 同步发布单元信息错误:", panic_err)
			beego.Error("UnitDataSync 同步发布单元信息详细信息:", string(buf[0:c]))
		}
	}()

	if beego.BConfig.RunMode != "prd" {
		beego.Info("只有生产环境才能同步发布单元!")
		return nil
	}

	beego.Info("同步发布单元信息 Start." + time.Now().Format("2006-01-02 15:04:05"))
	unit_sync()
	beego.Info("同步发布单元信息 End." + time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

func unit_sync() error {
	// 从pms拉取发布单元信息
	req := httplib.Get(beego.AppConfig.String("pms_baseurl") + "/mdp/unit/sync")
	req.Header("Authorization", "Basic mdeploy_d8c8680d046b1c60e63657deb3ce6d89")
	req.Header("Content-Type", "application/json")
	info_byte, err := req.Bytes()
	if err != nil {
		beego.Error(err.Error())
		return nil
	}

	type ReqData struct {
		Data []models.UnitConfList `json:"data"`
	}
	var ret ReqData
	err = json.Unmarshal(info_byte, &ret)
	if err != nil {
		return nil
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
				return nil
			}
		} else {
			err = tx.Create(&v).Error
			if err != nil {
				tx.Rollback()
				return nil
			}
		}
	}
	tx.Commit()

	// 将prd的发布单元信息同步回di和st
	cs_host := []string{"100.69.170.14", "100.70.42.52", "100.83.34.71"}
	//cs_host := []string{"127.0.0.1"}
	flag := 1
	FF:
	for {
		for _, h := range cs_host {
			var unit_conf []models.UnitConfList
			err := initial.DB.Model(models.UnitConfList{}).Order("id desc").Offset((flag - 1)*10).
				Limit(10).Find(&unit_conf).Error
			if err != nil {
				beego.Error(err.Error())
				return nil
			}
			if len(unit_conf) == 0 {
				break FF
			}

			url := fmt.Sprintf("http://%s/mdeploy/v1/ext/unit/sync", h)
			req := httplib.Post(url)
			req.Header("Authorization", "Basic mdeploy_IpFhvFjiQpV65PjIUywc3VHDjC0Wo9EM")
			req.Header("Content-Type", "application/json")
			req_data, _ := json.Marshal(unit_conf)
			req.Body(req_data)
			info_byte, err := req.Bytes()
			if err != nil {
				beego.Error(string(info_byte))
				beego.Error(err.Error())
				flag += 1
				continue
			}
			var ret map[string]interface{}
			err = json.Unmarshal(info_byte, &ret)
			if err != nil {
				beego.Error(string(info_byte))
				beego.Error(err.Error())
				return nil
			}
			if common.GetInt(ret["code"]) == 0 {
				beego.Error(common.GetString(ret["msg"]))
				flag += 1
				continue
			}
		}
		flag += 1
	}
	return nil
}