package operation

import (
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"library/common"
	"models"
	"strings"
	"time"
)

// @Title GetRecord
// @Description 获取容器平台的更新记录
// @Param	unit_id	query	string	false	"发布单元id"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} []models.OprCntrUpgrade
// @Failure 403
// @router /cntr/record [get]
func (c *CntrOprController) GetRecord() {
	unit_id := c.GetString("unit_id")
	page, rows := c.GetPageRows()
	cond := "1=1"
	if strings.Trim(unit_id, " ") != "" {
		cond = fmt.Sprintf("unit_id = %d", common.GetInt(unit_id))
	}
	var record []models.OprCntrUpgrade
	var cnt int
	err := initial.DB.Model(models.OprCntrUpgrade{}).Where(cond).Count(&cnt).Order("insert_time desc").Offset((page - 1)*rows).
		Limit(rows).Find(&record).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	var ret_data []RetData
	for _, v := range record {
		unit_info := cfunc.GetUnitInfoById(v.UnitId)
		per := RetData{
			OprCntrUpgrade: v,
			UnitEn: unit_info.Unit,
			UnitCn: unit_info.Info,
		}
		ret_data = append(ret_data, per)
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": ret_data,
	}
	c.SetJson(1, ret, "服务数据获取成功！")
}

// @Title GetRecordList
// @Description 获取容器平台的更新记录,可以根据开始结束时间/状态/发布单元ID/超时时间查询
// @Param	unit_id	query	string	false	"发布单元id"
// @Param	start	query	string	false	"查询开始日期 2019-12-01 格式"
// @Param	end		query	string  false	"查询结束日期 2019-12-01 格式"
// @Param	elaTime	query	string	false	"下拉选择<300,300~500,500~700,>700【分别传参300,500,700,700+】"
// @Param	status	query	string	false	"执行结果, s成功 f失败 为空全部"
// @Param	operator query	string	false	"执行用户, cpds / devops / others"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} []models.OprCntrUpgrade
// @Failure 403
// @router /cntr/record-list [get]
func (c *CntrOprController) GetRecordList() {
	var err error
	var ela string
	sStr := c.GetString("start")
	eStr := c.GetString("end")
	page, rows := c.GetPageRows()
	status := c.GetString("status")
	unitId := c.GetString("unit_id")
	ela = c.GetString("elaTime")
	ope := c.GetString("operator")
	
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	filtra := ""
	if sStr == "" && eStr == "" {
		filtra = "1 = 1"
	}else if sStr == "" || eStr == "" {
		c.SetJson(0,"","查询开始时间或截止时间未选择！")
		return
	}else {
		var (
			sTime time.Time
			eTime time.Time
		)

		if sTime, err = time.ParseInLocation("2006-01-02",sStr,time.Local) ; err != nil {
			c.SetJson(0,"",err.Error())
			return
		}
		if eTime, err = time.ParseInLocation("2006-01-02",eStr,time.Local) ; err != nil {
			c.SetJson(0,"",err.Error())
			return
		}
		if sTime.Sub(eTime).Seconds() > 0 {
			c.SetJson(0,"","开始时间大于截止时间，请重新选择！")
			return
		}

		sStr = sTime.Add(4 * time.Hour).Format("2006-01-02 15:04:05")
		eStr = eTime.Add(28 * time.Hour).Format("2006-01-02 15:04:05")

		filtra = "insert_time > " + "'" + sStr + "'" + " AND " + "insert_time <= " + "'" + eStr + "'"
	}

	if status == "s" {
		filtra = filtra + " AND result = 1"
	}else if status == "f" {
		filtra = filtra + " AND result != 1"
	}

	if unitId != "" {
		filtra = filtra + " AND unit_id = " + unitId
	}

	if ela != "" {
		switch ela {
		case "300":
			filtra = filtra + " AND cost_time < 300"
		case "500":
			filtra = filtra + " AND cost_time >= 300 AND cost_time <= 500"
		case "700":
			filtra = filtra + " AND cost_time > 500 AND cost_time <= 700"
		case "700+":
			filtra = filtra + " AND cost_time > 700"
		default:
			c.SetJson(0,"","耗时仅只可以指定 [<300 / 300~500 / 500~700 / >700] ！")
			return
		}
	}
	
	if ope != "" {
		switch ope {
		case "cpds":
			filtra = filtra + " AND operator = '" + "cpds" + "'"
		case "devops":
			filtra = filtra + " AND operator = '" +  "devops" + "'"
		case "others":
			filtra = filtra + " AND operator != " + "'" + "devops" + "'" + "AND operator != '" + "cpds" + "'"
		default:
			c.SetJson(0,"","操作人只可以选择 [cpds / devops / others] ！")
			return
		}	
	}

	count := 0
	records := make([]models.OprCntrUpgrade,0)
	err = initial.DB.Model(&models.OprCntrUpgrade{}).Where(filtra).Count(&count).Order("insert_time desc").Offset((page - 1)*rows).Limit(rows).Find(&records).Error
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	var ret_data []RetData
	for _, v := range records {
		unit_info := cfunc.GetUnitInfoById(v.UnitId)
		per := RetData{
			OprCntrUpgrade: v,
			UnitEn: unit_info.Unit,
			UnitCn: unit_info.Info,
		}
		ret_data = append(ret_data, per)
	}

	ret := map[string]interface{}{
		"cnt": count,
		"data": ret_data,
	}

	c.SetJson(1, ret, "服务数据获取成功！")
}

type RetData struct {
	models.OprCntrUpgrade
	UnitEn   string
	UnitCn   string
}