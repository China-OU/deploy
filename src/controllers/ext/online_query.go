package ext

import (
	"strings"
	"encoding/json"
	"models"
	"initial"
	"library/cfunc"
	"github.com/astaxie/beego"
	"time"
)

// @Title 检查当前环境该发布单元是否发布
// @Description 检查当前环境该发布单元是否发布
// @Param	body	body	ext.OnlineInput	true	"body形式的数据"
// @Success 200 true or false
// @Failure 403
// @router /online/query [post]
func (c *MultiEnvConnController) UnitOnlineQuery() {
	header := c.Ctx.Request.Header
	auth := ""
	var br []EnvRet
	if header["Authorization"] != nil && len(header["Authorization"]) > 0 {
		auth = header["Authorization"][0]
	} else {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "没有header!", "data": br}
		c.ServeJSON()
		return
	}
	if strings.Replace(auth, "Basic ", "", -1) != "mdeploy_IpFhvFjiQpV65PjIUywc3VHDjC0Wo9EM" {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "header校验失败!", "data": br}
		c.ServeJSON()
		return
	}

	var input OnlineInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		c.Data["json"] = map[string]interface{}{"code": 0, "msg": err.Error(), "data": br}
		c.ServeJSON()
		return
	}

	if input.QType == "date" {
		sha_arr := ShaQuery(input)
		if len(sha_arr) > 0 {
			c.Data["json"] = map[string]interface{}{"code": 1, "msg": "获取成功", "data": sha_arr}
			c.ServeJSON()
			return
		}
		unit_arr :=  UnitQuery(input)
		if len(unit_arr) > 0 {
			c.Data["json"] = map[string]interface{}{"code": 1, "msg": "获取成功", "data": unit_arr}
			c.ServeJSON()
			return
		}
	}

	if input.QType == "sha" {
		sha_arr := ShaQuery(input)
		if len(sha_arr) > 0 {
			c.Data["json"] = map[string]interface{}{"code": 1, "msg": "获取成功", "data": sha_arr}
			c.ServeJSON()
			return
		}
	}

	if input.QType == "unit" {
		unit_arr :=  UnitQuery(input)
		if len(unit_arr) > 0 {
			c.Data["json"] = map[string]interface{}{"code": 1, "msg": "获取成功", "data": unit_arr}
			c.ServeJSON()
			return
		}
	}

	if input.QType == "db_filename" {
		file_arr := FileNameQuery(input)
		if len(file_arr) > 0 {
			c.Data["json"] = map[string]interface{}{"code": 1, "msg": "获取成功", "data": file_arr}
			c.ServeJSON()
			return
		}
	}

	c.Data["json"] = map[string]interface{}{"code": 0, "msg": "没有找到数据", "data": br}
	c.ServeJSON()
}

type OnlineInput struct {
	QType   string   `json:"q_type"`
	UnitEn  string   `json:"unit_en"`
	Sha     string   `json:"sha"`
	Filename     string `json:"filename"`
	// 以下字段不需要传给子函数
	Branch       string  `json:"branch"`
	ReleaseDate  string  `json:"release_date"`
}

type EnvRet struct {
	// 有的数据没有 英文名和中文名，可以从di或st获取
	UnitEn       string  `json:"unit_en"`
	UnitCn       string  `json:"unit_cn"`
	Branch     string  `json:"branch"`
	Sha        string  `json:"sha"`
	Rd         string  `json:"rd"`
	Flag       int     `json:"flag"`
}

// 不能返回标签，sha匹配为1，unit匹配为2
func ShaQuery(input OnlineInput) []EnvRet {
	var online []models.OnlineAllList
	var data []EnvRet
	err := initial.DB.Model(models.OnlineAllList{}).Where("is_delete=0 and commit_id like ? and is_success in (0,1,10)",
		"%" + input.Sha + "%").Order("id desc").Limit(3).Find(&online).Error
	if err != nil {
		beego.Error(err.Error())
		return data
	}
	for _, v := range online {
		info := cfunc.GetUnitInfoById(v.UnitId)
		var per EnvRet
		per.UnitEn = info.Unit
		per.UnitCn = info.Name
		per.Branch = v.Branch
		per.Sha = v.CommitId
		per.Rd = v.OnlineDate
		per.Flag = 1
		data = append(data, per)
	}
	return data
}

func UnitQuery(input OnlineInput) []EnvRet {
	var online []models.OnlineAllList
	var data []EnvRet
	// 取最近15天
	detime := time.Now().AddDate(0, 0, -15).Format(initial.DatetimeFormat)
	err := initial.DB.Table("online_all_list a").Select("a.*").Joins("left join unit_conf_list b" +
		" ON a.unit_id = b.id").Where("b.unit = ? and a.is_delete = 0 and a.is_success in (0,1,10) and a.insert_time > ?",
			input.UnitEn, detime).Order("a.id desc").Limit(3).Find(&online).Error
	if err != nil {
		beego.Error(err.Error())
		return data
	}
	for _, v := range online {
		info := cfunc.GetUnitInfoById(v.UnitId)
		var per EnvRet
		per.UnitEn = info.Unit
		per.UnitCn = info.Name
		per.Branch = v.Branch
		per.Sha = v.CommitId
		per.Rd = v.OnlineDate
		per.Flag = 2
		data = append(data, per)
	}
	return data
}

func FileNameQuery(input OnlineInput) []EnvRet {
	var log []models.OnlineDbLog
	var data []EnvRet
	// 取最近15天
	detime := time.Now().AddDate(0, 0, -15).Format(initial.DatetimeFormat)
	err := initial.DB.Model(models.OnlineDbLog{}).Where("file_name=? and is_success in (0,1) and insert_time>?",
		input.Filename, detime).Order("id desc").Limit(3).Find(&log).Error
	if err != nil {
		beego.Error(err.Error())
		return data
	}
	for _, v := range log {
		ol := getOnlineList(v.OnlineId)
		info := cfunc.GetUnitInfoById(ol.UnitId)
		var per EnvRet
		per.UnitEn = info.Unit
		per.UnitCn = info.Name
		per.Branch = ol.Branch
		per.Sha = ol.CommitId
		per.Rd = ol.OnlineDate
		per.Flag = 2
		data = append(data, per)
	}
	return data
}

func getOnlineList(id int) models.OnlineAllList {
	var data models.OnlineAllList
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=?", id).First(&data).Error
	if err != nil {
		return models.OnlineAllList{}
	}
	return data
}
