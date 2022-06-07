package report

import (
	"github.com/astaxie/beego"
	"initial"
	"library/common"
	"math"
	"models"
	"strings"
	"time"
)

// 日期-发布数量
type DayIssue struct {
	Date string `json:"date"`
	Count int	`json:"count"`
}

// 发布时间/发布结果
type dayIssueIndex struct {
	Result int `json:"result" gorm:"column:is_success"`
	InsertTime string `json:"insertTime" gorm:"column:insert_time"`
}

// 成功率
type successRate struct {
	Value float64 `json:"value"`
	Molecule int`json:"molecule"`
	Denominator int `json:"denominator"`
}

// @Title GetIndex
// @Description 获取首页信息。
// @Success 200 {object} []index.DayIssue
// @Failure 403
// @router /index [get]
func (c *ReportController) GetIndex() {
	var (
		sTime string
		eTime string
		err error
	)

	nowDay := time.Now().Day()
	nowMonth := time.Now().Month()
	nowYear := time.Now().Year()
	sTime = time.Date(nowYear,nowMonth,nowDay - 39, 4, 0, 0, 0, time.Local).Format("2006-01-02 15:04:05")
	eTime = time.Date(nowYear,nowMonth,nowDay + 1,4,0,0,0, time.Local).Format("2006-01-02 15:04:05")


	issueAll := make([]*dayIssueIndex,0)
	db := initial.DB
	if err = db.Table("online_std_cntr a").Joins("left join online_all_list b on a.online_id = b.id").Select("b.is_success, b.insert_time").
		Where("a.is_delete = 0 AND b.is_delete = 0 AND b.insert_time >= ? AND b.insert_time < ?",sTime,eTime).Find(&issueAll).Error ; err != nil {
		beego.Error(err)
		c.SetJson(0,"", err.Error())
		return
	}

	dIseArr := make([]*DayIssue,0)

	dayList, err := common.GetDay(sTime, eTime)
	if err != nil {
		beego.Error(err)
		c.SetJson(0,"", err.Error())
		return
	}

	for _, day := range dayList {
		var count int
		eTime := day.Add(24 * time.Hour)

		for i := 0 ; i < len(issueAll) ; i++ {
			var insTime time.Time
			if insTime, err = time.ParseInLocation("2006-01-02 15:04:05", issueAll[i].InsertTime, time.Local) ; err != nil {
				c.SetJson(0,"", err.Error())
				return
			}
			if insTime.Sub(day).Seconds() >= 0 && eTime.Sub(insTime).Seconds() > 0{
				count+=1
				issueAll = append(issueAll[:i], issueAll[i+1:]...)
				i --
			}
		}
		dIseArr = append(dIseArr,&DayIssue{
			Date:strings.Split(day.Format("2006-01-02 15:04:05"), " ")[0],
			Count:count,
		})
	}

	// 标准容器总发布成功率
	var aIseCnt int
	if err = db.Table("online_std_cntr a").Joins("left join online_all_list b on a.online_id = b.id").Where("a.is_delete = 0 AND b.is_delete = 0").
		Count(&aIseCnt).Error ; err != nil {
		c.SetJson(0,"", err.Error())
		return
	}
	if aIseCnt == 0 {
		aIseCnt = -1
	}
	var sIseCnt int
	if err = db.Table("online_std_cntr a").Joins("left join online_all_list b on a.online_id = b.id").
		Where("a.is_delete = 0 AND b.is_delete = 0 AND b.is_success = 1").Count(&sIseCnt).Error ; err != nil {
		c.SetJson(0,"", err.Error())
		return
	}
	isuSR := successRate{
		Value:GetNumFormat(float64(sIseCnt) / float64(aIseCnt) * 100),
		Molecule:sIseCnt,
		Denominator:aIseCnt,
	}

	// 标准容器构建成功率
	var stdCnt int
	if err = db.Table("online_std_cntr a").Joins("left join online_all_list b on a.online_id = b.id").
		Where("a.is_delete = 0 AND b.is_delete = 0 AND a.jenkins_success = 1").Count(&stdCnt).Error ; err != nil {
		c.SetJson(0,"", err.Error())
		return
	}
	bSR := successRate{
		Value:GetNumFormat(float64(stdCnt) / float64(aIseCnt) * 100),
		Molecule:stdCnt,
		Denominator:aIseCnt,
	}

	// 总更新成功率
	var aUpCnt int
	if err = db.Table("opr_cntr_upgrade").Where("unit_id != 0").Count(&aUpCnt).Error ; err != nil {
		c.SetJson(0,"", err.Error())
		return
	}
	if aUpCnt == 0 {
		aUpCnt = -1
	}
	var sUpCnt int
	if err = db.Table("opr_cntr_upgrade").Where("result = 1").Count(&sUpCnt).Error ; err != nil {
		c.SetJson(0,"", err.Error())
		return
	}
	upSR := successRate{
		Value:GetNumFormat(float64(sUpCnt) / float64(aUpCnt) * 100),
		Molecule:sUpCnt,
		Denominator:aUpCnt,
	}

	res := map[string]interface{}{
		"dayIssueCnt" : dIseArr,
		"issueSuccess" : isuSR,
		"buildSuccess" :bSR,
		"upSuccess" : upSR,
	}


	// #################################### 标准虚机 ####################################
	issueAllVM := make([]*dayIssueIndex,0)
	if err = db.Table("online_std_vm a").Joins("inner join online_all_list b on a.online_id = b.id").Select("b.is_success, b.insert_time").
		Where("a.is_delete = 0 AND b.is_delete = 0 AND b.insert_time >= ? AND b.insert_time < ?",sTime,eTime).Find(&issueAllVM).Error ; err != nil {
		beego.Error(err)
		c.SetJson(0,"", err.Error())
		return
	}

	vm, dayVM, err := VmIndexData(dayList, issueAllVM)
	if err != nil {
		c.SetJson(0,"", err.Error())
		return
	}

	for k, v := range vm {
		res[k] = v
	}

	/////// 折线图数据
	type lineChart struct {
		Date string `json:"date"`
		VM int	`json:"虚机应用"`
		Cntr int `json:"标准容器"`
	}

	chartArr := make([]*lineChart, 0)
	for k, v := range dIseArr{
		v2 := dayVM[k]
		if v.Date == v2.Date {
			chartArr = append(chartArr, &lineChart{
				Date: v2.Date,
				VM: v2.Count,
				Cntr: v.Count,
			})
		}
	}

	res["lineChart"] = chartArr

	c.SetJson(1, res,"服务数据获取成功！")
}

func VmIndexData(dayList []time.Time, all []*dayIssueIndex) (data map[string]interface{}, dayIssueCntVM []*DayIssue,  err error){
	dayArr := make([]*DayIssue, 0)
	data = make(map[string]interface{})
	for _, day := range dayList {
		var count int
		eTime := day.Add(24 * time.Hour)

		for i := 0 ; i < len(all) ; i++ {
			var insTime time.Time
			if insTime, err = time.ParseInLocation("2006-01-02 15:04:05", all[i].InsertTime, time.Local) ; err != nil {
				return nil, nil, err
			}
			if insTime.Sub(day).Seconds() >= 0 && eTime.Sub(insTime).Seconds() > 0{
				count+=1
				all = append(all[:i],all[i+1:]...)
				i --
			}
		}

		dayArr = append(dayArr, &DayIssue{
			Date:strings.Split(day.Format("2006-01-02 15:04:05"), " ")[0],
			Count:count,
		})
	}
	//data["dayIssueCntVM"] = dayArr

	// 部署成功率
	var aIssCnt int
	if err = initial.DB.Model(&models.OnlineStdVM{}).Where(" is_delete = 0").Count(&aIssCnt).Error ; err != nil {
		return nil, nil, err
	}

	denominator := aIssCnt
	if aIssCnt == 0 {
		aIssCnt = -1
	}
	var sIssCnt int
	if err = initial.DB.Model(&models.OnlineStdVM{}).Where("is_delete = 0 AND upgrade_status = 1").Count(&sIssCnt).Error ; err != nil {
		return nil, nil , err
	}
	deploySR := successRate{
		Value: GetNumFormat(float64(sIssCnt) / float64(aIssCnt) *100),
		Molecule: sIssCnt,
		Denominator: denominator,
	}
	data["issueSuccessVM"] = deploySR

	// 构建成功率
	var buildS int
	if err = initial.DB.Model(&models.OnlineStdVM{}).Where("is_delete = 0 AND build_status = 1").Count(&buildS).Error ; err != nil {
		return nil, nil, err
	}
	buildSR := successRate{
		Value: GetNumFormat(float64(buildS) / float64(aIssCnt) *100),
		Molecule: buildS,
		Denominator: denominator,
	}
	data["buildSuccessVM"] = buildSR

	// 升级成功率(当前升级和构建未分离，现有构建后有升级)
	data["upSuccessVM"] = deploySR

	return data, dayArr, nil
}

// 四舍五入成功率,保留2位小数
func GetNumFormat(n float64) float64{
	res := math.Floor(n*100+0.05) / 100
	if res >= 100.0 {
		return math.Floor(n*100) / 100
	} else {
		return math.Floor(res*100) / 100
	}
}