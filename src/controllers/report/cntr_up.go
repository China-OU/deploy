package report

import (
	"errors"
	"fmt"
	"initial"
	"library/common"
	"models"
	"sort"
	"strings"
	"time"
)

// @Title CntrUp
// @Description 获取容器更新统计信息;当开始/截止时间和快速查询都未指定时,默认查询周期为最近一个月;
// @Param	start	query	string	false	"查询开始时间，2019-01-02格式"
// @Param	end		query	string	false	"查询截至时间，2019-01-02格式"
// @Param	fail	query	string	false	"失败统计标准定义:更新失败，更新耗时大于300s，更新耗时大于500s，更新耗时大于700s；分别传参【fail，300，500，700】"
// @Param	quick	query	string	false	"快速查询 【最近一个月,最近三个月,最近六个月】，分别对应【1 3 6】"
// @Success 200 true or false
// @Failure 403
// @router /cntr-up [get]
func (c *ReportController) CntrUp() {
	var err error
	sStr := c.GetString("start")
	eStr := c.GetString("end")
	qStr := c.GetString("quick")
	fail := c.GetString("fail")

	sStr, eStr, err = common.FormatTime(qStr,sStr,eStr)
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	sArr, fArr, err := sortIse(sStr, eStr, fail)
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	aCntArr, fCntArr, err := getDayCnt(sStr, eStr, fail)
	if err != nil {
		c.SetJson(0,"",err.Error())
		return
	}

	var upList []*models.OprCntrUpgrade
	if err = initial.DB.Model(&models.OprCntrUpgrade{}).Where("insert_time >= ? AND insert_time < ?", sStr, eStr).Find(&upList).Error ; err != nil {
		c.SetJson(0,"",err.Error())
		return
	}
	srcTop:= sourceTop(upList)
	costTop:= costTimeTop(upList)

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
		"srcTop":srcTop,
		"costTimeTop":costTop,
	}

	c.SetJson(1, res, "服务数据获取成功！")
}

// 排序更新总次数和更新失败次数
func sortIse(sStr, eStr, fail string) (aArr, fArr []*issueSts, err error){
	if fail == "" {
		fail = "fail"
	}
	aMap := make(map[string]*issueSts)
	fMap := make(map[string]*issueSts)
	aNameL := ""
	fNameL := ""

	type cntrUpSts struct {
		ID int `json:"id" gorm:"column:id"`
		Name string `json:"name" gorm:"column:unit"`
		Info string `json:"info" gorm:"column:info"`
		Result int `json:"result" gorm:"column:result"`
		Cnt int `json:"count"`
		CostTime int `json:"costTime" gorm:"column:cost_time"`
	}
	var iseList []*cntrUpSts

	if err = initial.DB.Table("opr_cntr_upgrade a").Joins("left join unit_conf_list b on a.unit_id = b.id").
		Select("b.unit, b.info, a.result, b.id, a.cost_time").
		Where("b.is_offline = 0 AND a.insert_time > ? AND a.insert_time < ?", sStr, eStr).Find(&iseList).Error ; err != nil {
		return nil, nil, err
	}
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
		switch fail {
		case "fail":
			if v.Result == 0 {
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
		case "300":
			if v.CostTime > 300 || v.Result == 0 {
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
		case "500":
			if v.CostTime > 500 || v.Result == 0 {
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
		case "700":
			if v.CostTime > 700 || v.Result == 0 {
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
		default:
			return nil, nil, errors.New("失败标准只允许为：更新失败，更新耗时大于300s，更新耗时大于500s，更新耗时大于700s")
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

// 容器更新每天失败次数和总次数
func getDayCnt(sStr, eStr, fail string) ([]*dayIssue, []*dayIssue, error){
	var (
		dayList []time.Time
		err error
	)

	if dayList, err = common.GetDay(sStr,eStr) ; err != nil {
		return nil, nil, err
	}
	// 遍历获取每天的更新次数
	aCntArr := make([]*dayIssue,0)
	fCntArr := make([]*dayIssue,0)

	type cntrUp struct {
		Result int `json:"result" gorm:"column:result"`
		InsertTime string `json:"insertTime" gorm:"column:insert_time"`
		ID int `json:"id" gorm:"column:id"`
		CostTime int `json:"costTime" gorm:"column:cost_time"`
	}
	var upAll []*cntrUp

	if err = initial.DB.Table("opr_cntr_upgrade").Where("insert_time >= ? AND insert_time < ?", sStr, eStr).
		Find(&upAll).Error ; err != nil {
			return nil, nil, err
	}
	if fail == "" {
		fail = "fail"
	}
	for _, v := range dayList {
		eTime := v.Add(24 * time.Hour)
		aCnt := 0
		fCnt := 0

		for i := 0 ; i < len(upAll); i++ {
			var insTime time.Time
			if insTime, err = time.ParseInLocation("2006-01-02 15:04:05", upAll[i].InsertTime, time.Local) ; err != nil {
				return nil, nil, err
			}
			if insTime.Sub(v).Seconds() >= 0 && eTime.Sub(insTime).Seconds() > 0{
				aCnt = aCnt + 1
				switch fail {
				case "fail":
					if upAll[i].Result == 0 {
						fCnt = fCnt + 1
					}
				case "300":
					if upAll[i].CostTime > 300 || upAll[i].Result == 0 {
						fCnt = fCnt + 1
					}
				case "500":
					if upAll[i].CostTime > 500 || upAll[i].Result == 0  {
						fCnt = fCnt + 1
					}
				case "700":
					if upAll[i].CostTime > 700  || upAll[i].Result == 0 {
						fCnt = fCnt + 1
					}
				default:
					return nil, nil, errors.New("失败标准只允许为：发布失败，耗时大于300s，耗时大于500s，耗时大于700s")
				}
				// 删除已匹配到的记录
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

	return aCntArr, fCntArr, nil
}

// 来源(更新总表)
func sourceTop(upList []*models.OprCntrUpgrade) ([]*source){
	var (
		cpds int
		devops int
		other int
	)

	for _, v := range upList {
		if v.Operator == "devops"{
			devops += 1
		}else if v.Operator == "cpds" {
			cpds += 1
		}else {
			other += 1
		}
	}
	top := []*source{
		{
			Name:"devops",
			Count:devops,
		},
		{
			Name:"cpds",
			Count:cpds,
		},
		{
			Name:"other",
			Count:other,
		},
	}
	sort.Slice(top, func(i, j int) bool {
		s := false
		if top[i].Count > top[j].Count {
			s = true
		}else if top[i].Count == top[j].Count {
			if top[i].Name[0] > top[j].Name[0] {
				s = true
			}
		}
		return s
	})

	return top
}

// 耗时(更新总表)
func costTimeTop(upList []*models.OprCntrUpgrade) ([]*source) {
	var (
		gt int	// >300
		tf int	// 300-500
		fs int	// 500-700
		gs int // > 700
	)

	for _, v := range upList {
		if v.CostTime < 300 && v.CostTime > 0{
			gt += 1
		}else if v.CostTime > 700 {
			gs += 1
		}else if v.CostTime > 500 && v.CostTime <= 700 {
			fs += 1
		}else if v.CostTime >= 300 && v.CostTime <= 500 {
			tf += 1
		}
	}
	top := []*source{
		{
			Name:"小于300s",
			Count:gt,
		},
		{
			Name:"300s~500s",
			Count:tf,
		},
		{
			Name:"500s~700s",
			Count:fs,
		},
		{
			Name:"大于700s",
			Count:gs,
		},
	}
	sort.Slice(top, func(i, j int) bool {
		return top[i].Count > top[j].Count
	})

	return top
}

type issueSts struct {
	ID int `json:"id" gorm:"column:id"`
	Name string `json:"name" gorm:"column:unit"`
	Info string `json:"info" gorm:"column:info"`
	Result int `json:"result" gorm:"column:is_success"`
	Cnt int `json:"count"`
	CostTime int `json:"costTime" gorm:"column:cost_time"`
}

type dayIssue struct {
	Date string `json:"date"`
	Count int	`json:"count"`
}

type source struct {
	Name string  `json:"name"`
	Count int	`json:"count"`
}