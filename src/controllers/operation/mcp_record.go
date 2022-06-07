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

// @Description 获取多容器平台的更新记录
// @Param	unit_id	query	string	false	"发布单元id"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} []models.McpUpgradeList
// @Failure 403
// @router /mcp/record [get]
func (c *McpOprController) McpRecord() {
	unit_id := c.GetString("unit_id")
	page, rows := c.GetPageRows()
	cond := "1=1"
	if strings.Trim(unit_id, " ") != "" {
		cond = fmt.Sprintf("unit_id = %d", common.GetInt(unit_id))
	}
	var record []models.McpUpgradeList
	var cnt int
	err := initial.DB.Model(models.McpUpgradeList{}).Where(cond).Count(&cnt).Order("insert_time desc").Offset((page - 1)*rows).
		Limit(rows).Find(&record).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	var ret_data []McpRetData
	for _, v := range record {
		unit_info := cfunc.GetUnitInfoById(v.UnitId)
		_, mconf := cfunc.GetContainerTypeByUnitId(v.UnitId)
		per := McpRetData{
			McpUpgradeList: v,
			UnitEn: unit_info.Unit,
			UnitCn: unit_info.Info,
			ContainerType: mconf.ContainerType,
		}
		ret_data = append(ret_data, per)
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": ret_data,
	}
	c.SetJson(1, ret, "多容器平台升级数据获取成功！")
}

// @Description 获取多容器平台的更新记录,可以根据开始结束时间/状态/发布单元ID/超时时间查询
// @Param	unit_id	query	string	false	    "发布单元id"
// @Param	start_time	query	string	false	"查询开始日期 2019-12-01 格式"
// @Param	end_time		query	string  false	"查询结束日期 2019-12-01 格式"
// @Param	timeout	query	string	false	"更新超时时间，参数为300/500/700/700+"
// @Param	status	query	string	false	"执行结果, 1成功 0失败 为空全部"
// @Param	operator query	string	false	"执行用户, cpds / devops / others"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} []models.McpUpgradeList
// @Failure 403
// @router /mcp/record-list [get]
func (c *McpOprController) McpRecordList() {
	unit_id := c.GetString("unit_id")
	start_time := c.GetString("start_time")
	end_time := c.GetString("end_time")
	timeout := c.GetString("timeout")
	status := c.GetString("status")
	operator := c.GetString("operator")
	page, rows := c.GetPageRows()

	cond := " 1=1 "
	if start_time != "" && end_time == "" {
		end_time = time.Now().Format("2006-01-02")
	}
	if start_time == "" && end_time != "" {
		start_time = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if start_time != "" && end_time != "" {
		st, err := time.ParseInLocation("2006-01-02", start_time, time.Local)
		if err != nil {
			c.SetJson(0,"",err.Error())
			return
		}
		et, err := time.ParseInLocation("2006-01-02", end_time, time.Local)
		if err != nil {
			c.SetJson(0,"",err.Error())
			return
		}
		if st.Sub(et).Seconds() > 0 {
			c.SetJson(0,"","开始时间大于截止时间，请重新选择！")
			return
		}
		st_time := st.Add(4 * time.Hour).Format(initial.DatetimeFormat)
		et_time := et.Add(28 * time.Hour).Format(initial.DatetimeFormat)
		cond += fmt.Sprintf(" and insert_time > '%s' AND insert_time < '%s' ", st_time, et_time)
	}
	if status != "" {
		cond += fmt.Sprintf(" and result = '%s' ", status)
	}
	if strings.Trim(unit_id, " ") != "" {
		cond += fmt.Sprintf(" and unit_id = %d", common.GetInt(unit_id))
	}

	if timeout != "" {
		if timeout == "300" {
			cond += fmt.Sprintf(" and cost_time<%d", 300)
		} else if timeout == "500" {
			cond += fmt.Sprintf(" and cost_time>%d and cost_time<%d", 300, 500)
		} else if timeout == "700" {
			cond += fmt.Sprintf(" and cost_time>%d and cost_time<%d", 500, 700)
		} else if timeout == "700+" {
			cond += fmt.Sprintf(" and cost_time>%d", 700)
		} else {
			c.SetJson(0,"","耗时仅只可以指定 [<300 / 300~500 / 500~700 / >700] ！")
			return
		}
	}

	if operator != "" {
		if operator == "others" {
			cond += fmt.Sprintf(" and operator != 'cpds' and operator != 'devops' ")
		} else {
			cond += fmt.Sprintf(" and operator = '%s' ", operator)
		}
	}

	var record []models.McpUpgradeList
	var cnt int
	err := initial.DB.Model(models.McpUpgradeList{}).Where(cond).Count(&cnt).Order("insert_time desc").Offset((page - 1)*rows).
		Limit(rows).Find(&record).Error
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	var ret_data []McpRetData
	for _, v := range record {
		unit_info := cfunc.GetUnitInfoById(v.UnitId)
		_, mconf := cfunc.GetContainerTypeByUnitId(v.UnitId)
		per := McpRetData{
			McpUpgradeList: v,
			UnitEn: unit_info.Unit,
			UnitCn: unit_info.Info,
			ContainerType: mconf.ContainerType,
		}
		ret_data = append(ret_data, per)
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": ret_data,
	}
	c.SetJson(1, ret, "多容器平台升级数据获取成功！")
}

type McpRetData struct {
	models.McpUpgradeList
	UnitEn   string   `json:"unit_en"`
	UnitCn   string   `json:"unit_cn"`
	ContainerType  string    `json:"container_type"`
}