package report

import (
	"fmt"
	"initial"
	"library/common"
	"sort"
	"strings"
	"time"
)

// @Title CntrDeploy
// @Description 获取容器发布统计信息;当开始/截止时间和快速查询都未指定时,默认查询周期为最近一个月
// @Param	start	query	string	false	"查询开始时间，2019-01-02格式"
// @Param	end		query	string	false	"查询截至时间，2019-01-02格式"
// @Param	quick	query	string	false	"快速查询 【最近一个月,最近三个月,最近六个月】，分别对应【1 3 6】"
// @Success 200 true or false
// @Failure 403
// @router /cntr-deploy [get]
func (c *ReportController) CntrDeploy() {
	var err error
	sStr := c.GetString("start")
	eStr := c.GetString("end")
	qStr := c.GetString("quick")

	sStr, eStr, err = common.FormatTime(qStr,sStr,eStr)
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	// 标准容器 - 获取统计周期内发布次数总TOP和错误次数TOP
	db := initial.DB
	var iseList []*issueSts
	if err = db.Table("online_std_cntr a").Joins("left join online_all_list b on a.online_id = b.id").
		Joins("left join unit_conf_list c on b.unit_id = c.id").
		Select("c.unit, c.info, b.is_success, c.id").
		Where("a.is_delete = 0 AND b.insert_time > ? AND b.insert_time < ?", sStr, eStr).Find(&iseList).Error ; err != nil {
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
	aCntArr := make([]*dayIssue,0)
	fCntArr := make([]*dayIssue,0)

	type issueCntr struct {
		Result int `json:"result" gorm:"column:is_success"`
		InsertTime string `json:"insertTime" gorm:"column:insert_time"`
		ID int `json:"id" gorm:"column:id"`
	}
	var upAll []*issueCntr

	if err = db.Table("online_std_cntr a").Joins("left join online_all_list b on a.online_id = b.id").
		Select("b.is_success, b.insert_time").
		Where("a.is_delete = 0 AND b.insert_time >= ? AND b.insert_time < ?",sStr,eStr).Find(&upAll).Error ; err != nil {
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
		aCntArr = append(aCntArr,&dayIssue{
			Date:strings.Split(s, " ")[0],
			Count:aCnt,
		})

		fCntArr = append(fCntArr,&dayIssue{
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
		"successTop":sArr,
		"faultTop":fArr,
		"allIssue":aCntArr,
		"faultIssue":fCntArr,
	}

	c.SetJson(1,res,"服务数据获取成功！")
}

// 排序发布总次数和发布错误次数
func sortIseDeploy(iseList []*issueSts) (aArr, fArr []*issueSts, err error){
	aMap := make(map[string]*issueSts)
	fMap := make(map[string]*issueSts)
	aNameL := ""
	fNameL := ""

	for _, v := range iseList {
		if _, ok := aMap[v.Name] ; ok {
			aMap[v.Name].Cnt = aMap[v.Name].Cnt + 1
		}else {
			aMap[v.Name] = &issueSts{
				Name:v.Name,
				Info:v.Info,
				Cnt:1,
				ID:v.ID,
			}
			aNameL = aNameL + "'" + v.Name + "'" + ","
		}

		if v.Result == 0{
			if _, ok := fMap[v.Name] ; ok {
				fMap[v.Name].Cnt = fMap[v.Name].Cnt + 1
			}else {
				fMap[v.Name] = &issueSts{
					Name:v.Name,
					Info:v.Info,
					Cnt:1,
					ID:v.ID,
				}
				fNameL = fNameL + "'" + v.Name + "'" + ","
			}
		}
	}

	for _,v := range aMap {
		aArr = append(aArr, &issueSts{
			ID:v.ID,
			Name:v.Name,
			Info:v.Info,
			Cnt:v.Cnt,
		})
	}
	for _,v := range fMap {
		fArr = append(fArr, &issueSts{
			ID:v.ID,
			Name:v.Name,
			Info:v.Info,
			Cnt:v.Cnt,
		})
	}

	// 排序(先按照数量,数量相同时按照发布单元ID)
	sort.Slice(aArr, func(i, j int) bool {
		var s  bool
		if  aArr[i].Cnt > aArr[j].Cnt {
			s = true
		}else if aArr[i].Cnt == aArr[j].Cnt {
			if aArr[i].ID > aArr[j].ID {
				s = true
			}
		}
		return s
	})
	sort.Slice(fArr, func(i, j int) bool {
		var s  bool
		if  fArr[i].Cnt > fArr[j].Cnt {
			s = true
		}else if fArr[i].Cnt == fArr[j].Cnt {
			if fArr[i].ID > fArr[j].ID {
				s =true
			}
		}
		return s
	})


	// 不足10追加,超过只取前10
	db := initial.DB
	data := make([]*issueSts,0)
	sql := `SELECT unit, info FROM unit_conf_list  WHERE (is_offline = 0 AND unit NOT IN  {exist}) LIMIT {limit} OFFSET 20`

	if len(aArr) > 10 {
		aArr = aArr[:10]
	}else if len(aArr) < 10 {
		l := 10 - len(aArr)
		aSql := strings.Replace(sql,`{exist}`,"(" + aNameL + "'" + "" + "'" + ")",-1)
		aSql = strings.Replace(aSql,`{limit}`,fmt.Sprintf("%d", l),-1)
		if err = db.Raw(aSql).Scan(&data).Error ; err != nil {
			return nil, nil, err
		}
		aArr = append(aArr,data...)
	}

	if len(fArr) > 10 {
		fArr = fArr[:10]
	}else if len(fArr) < 10 {
		l := 10 - len(fArr)
		fSql := strings.Replace(sql,`{exist}`,"(" + fNameL + "'" + "" + "'" + ")",-1)
		fSql = strings.Replace(fSql,`{limit}`,fmt.Sprintf("%d", l),-1)
		if err = db.Raw(fSql).Scan(&data).Error ; err != nil {
			return nil, nil, err
		}
		if err != nil {
			return nil, nil, err
		}

		fArr = append(fArr,data...)
	}

	return aArr, fArr, nil
}



