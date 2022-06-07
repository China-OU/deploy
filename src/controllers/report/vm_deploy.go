package report

import (
	"initial"
	"library/common"
	"strings"
	"time"
)

// @Title VmDeploy
// @Description 获取容器发布统计信息;当开始/截止时间和快速查询都未指定时,默认查询周期为最近一个月
// @Param	start	query	string	false	"查询开始时间，2019-01-02格式"
// @Param	end		query	string	false	"查询截至时间，2019-01-02格式"
// @Param	quick	query	string	false	"快速查询 【最近一个月,最近三个月,最近六个月】，分别对应【1 3 6】"
// @Success 200 true or false
// @Failure 403
// @router /vm-deploy [get]
func (c *ReportController) VmDeploy() {
	var err error
	sStr := c.GetString("start")
	eStr := c.GetString("end")
	qStr := c.GetString("quick")

	// 获取查询开始/结束时间
	sStr, eStr, err = common.FormatTime(qStr,sStr,eStr)
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	// 获取发布总次数和错误总次数
	db := initial.DB
	var iseList []*issueSts
	if err = db.Table("online_std_vm a").Joins("left join online_all_list b on a.online_id = b.id").
		Joins("left join unit_conf_list c on b.unit_id = c.id").
		Select("c.unit, c.info, b.is_success, c.id").
		Where("a.is_delete = 0 AND b.is_delete = 0 AND b.insert_time > ? AND b.insert_time < ?", sStr, eStr).Find(&iseList).Error ; err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	sArr, fArr, err := sortIseDeploy(iseList)
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	var dayList []time.Time
	type dayIssue struct {
		Date string `json:"date"`
		Count int	`json:"count"`
	}

	if dayList, err = common.GetDay(sStr,eStr) ; err != nil {
		c.SetJson(0,"",err.Error())
		return
	}
	// 遍历获取每天的发布次数
	aVmArr := make([]*dayIssue,0)
	fVmArr := make([]*dayIssue,0)

	type issueVm struct {
		Result int `json:"result" gorm:"column:is_success"`
		InsertTime string `json:"insertTime" gorm:"column:insert_time"`
		ID int `json:"id" gorm:"column:id"`
	}
	var upAll []*issueVm

	if err = db.Table("online_std_vm a").Joins("left join online_all_list b on a.online_id = b.id").
		Select("b.is_success, b.insert_time").
		Where("a.is_delete = 0 AND b.is_delete = 0 AND b.insert_time >= ? AND b.insert_time < ?",sStr,eStr).Find(&upAll).Error ; err != nil {
		c.SetJson(0,"", err.Error())
		return
	}

	for _, v := range dayList {
		eTime := v.Add(24 * time.Hour)
		aCnt := 0
		fCnt := 0

		for i := 0 ; i < len(upAll); i++ {
			var insTime time.Time
			if insTime, err = time.ParseInLocation("2006-01-02 15:04:05", upAll[i].InsertTime, time.Local) ; err != nil {
				c.SetJson(0,"", err.Error())
				return
			}
			if insTime.Sub(v).Seconds() >= 0 && eTime.Sub(insTime).Seconds() > 0{
				aCnt = aCnt + 1
				if upAll[i].Result == 0{
					fCnt = fCnt + 1
				}
				upAll = append(upAll[:i],upAll[i+1:]...)
				i --
			}
		}

		s := v.Format("2006-01-02 15:03:04")
		aVmArr  = append(aVmArr ,&dayIssue{
			Date:strings.Split(s, " ")[0],
			Count:aCnt,
		})

		fVmArr = append(fVmArr ,&dayIssue{
			Date:strings.Split(s, " ")[0],
			Count:fCnt,
		})
	}

	tmp, err := time.ParseInLocation("2006-01-02 15:04:05", eStr, time.Local)
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}
	eStr = tmp.Add(-24 * time.Hour).Format("2006-01-02")

	res := map[string]interface{}{
		"start":strings.Split(sStr," ")[0],
		"end":eStr,
		"allTop":sArr,
		"faultTop":fArr,
		"allIssue":aVmArr,
		"faultIssue":fVmArr,
	}

	c.SetJson(1,res,"服务数据获取成功！")
}
