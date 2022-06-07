package online

import (
	"github.com/astaxie/beego"
	"library/common"
	"models"
	"initial"
	"library/cfunc"
	"regexp"
	"strings"
	"encoding/json"
	"github.com/astaxie/beego/httplib"
	"fmt"
	"time"
)

// @Title 查询di/st部署平台的发布结果
// @Description 查询di/st部署平台的发布结果
// @Param	qry_type	  query	string	false	"查询类型，如 date/sha/unit/db_filename"
// @Param	qry_word	  query	string	false	"关键字，输入框"
// @Success 200  true or false
// @Failure 403
// @router /record/query [get]
func (c *ReleaseRecordController) QueryReleaseRecord() {
	if c.Role == "guest" || c.Role == "deploy-single" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	if beego.AppConfig.String("runmode") != "prd" {
		c.SetJson(0, "", "只有生产环境才能拉取！")
		return
	}

	qtype := c.GetString("qry_type")
	qword := c.GetString("qry_word")
	if common.InList(qtype, []string{"date", "sha", "unit", "db_filename"}) == false {
		c.SetJson(0, "", "请选择正确的查询类型！")
		return
	}
	if qtype == "date" {
		flag, _ := regexp.MatchString("^[0-9]{8}$", qword)
		if !flag {
			c.SetJson(0, "", "日期格式不正确！")
			return
		}
		input := GetDateList(qword)
		output := CheckIfRelease(input)
		c.SetJson(1, output, "结果查询成功！")
		return
	}

	if qtype == "sha" {
		if len(qword) < 8 {
			c.SetJson(0, "", "sha值最少要有8位！")
			return
		}
		var input []QueryInput
		input = append(input, QueryInput{
			QType: "sha",
			UnitEn: "",
			Sha: qword,
			Filename: "",
		})
		output := CheckIfRelease(input)
		c.SetJson(1, output, "结果查询成功！")
		return
	}

	if qtype == "unit" {
		// 待开发，判断unit正不正确
		var input []QueryInput
		input = append(input, QueryInput{
			QType: "unit",
			UnitEn: qword,
			Sha: "",
			Filename: "",
		})
		output := CheckIfRelease(input)
		c.SetJson(1, output, "结果查询成功！")
		return
	}

	if qtype == "db_filename" {
		if strings.Contains(qword, ".sql") == false {
			c.SetJson(0, "", "请输入完整的sql语句名，例如：xxxx.sql")
			return
		}
		var input []QueryInput
		input = append(input, QueryInput{
			QType: "db_filename",
			UnitEn: "",
			Sha: "",
			Filename: qword,
		})
		output := CheckIfRelease(input)
		c.SetJson(1, output, "结果查询成功！")
		return
	}
	c.SetJson(0, "", "查询有误！")
}

type QueryInput struct {
	QType   string   `json:"q_type"`
	UnitEn  string   `json:"unit_en"`
	Sha     string   `json:"sha"`
	Filename     string `json:"filename"`
	// 以下字段不需要传给子函数
	Branch       string  `json:"branch"`
	ReleaseDate  string  `json:"release_date"`
}

type OutputData struct {
	UnitEn       string  `json:"unit_en"`
	UnitCn       string  `json:"unit_cn"`
	Branch       string  `json:"branch"`
	ShortSha     string  `json:"short_sha"`
	Filename     string  `json:"filename"`
	ReleaseDate  string  `json:"release_date"`
	// rd: release_date, flag是版本标签：1表示同sha值发布，2表示最近三天有发布过相关发布单元，10表示不需要关注
	DiBranch     string  `json:"di_branch"`
	DiSha        string  `json:"di_sha"`
	DiRd         string  `json:"di_rd"`
	DiFlag       int     `json:"di_flag"`
	StBranch     string  `json:"st_branch"`
	StSha        string  `json:"st_sha"`
	StRd         string  `json:"st_rd"`
	StFlag       int     `json:"st_flag"`
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

type DataRet struct {
	Code   int       `json:"code"`
	Data   []EnvRet  `json:"data"`
	Msg    string    `json:"msg"`
}

func GetDateList(online_date string) []QueryInput {
	var input []QueryInput
	var date_list []models.OnlineAllList
	// 查发布成功的和未发布的，后续再看要不要扩大范围
	initial.DB.Model(models.OnlineAllList{}).Where("online_date=? and is_delete=0 and is_success in (1, 10)", online_date).Find(&date_list)
	for _, v := range date_list {
		vinfo := cfunc.GetUnitInfoById(v.UnitId)
		input = append(input, QueryInput{
			QType: "date",
			UnitEn: vinfo.Unit,
			Sha: v.CommitId,
			Filename: "",
			Branch: v.Branch,
			ReleaseDate: v.OnlineDate,
		})
	}
	return input
}

func CheckIfRelease(input []QueryInput) []OutputData {
	var ret []OutputData
	for _, v := range input {
		if v.QType == "date" {
			// 返回一条数据
			info := cfunc.GetUnitInfoByName(v.UnitEn)
			var per OutputData
			per.UnitEn = v.UnitEn
			per.UnitCn = info.Name
			per.Branch = v.Branch
			per.ShortSha = v.Sha
			per.Filename = v.Filename
			per.ReleaseDate = v.ReleaseDate

			di_rel := CheckEnvRelInfo(v, "100.69.170.14")
			if len(di_rel) > 0 {
				per.DiBranch = di_rel[0].Branch
				per.DiSha = di_rel[0].Sha
				per.DiRd = di_rel[0].Rd
				per.DiFlag = di_rel[0].Flag
			}
			st_rel := CheckEnvRelInfo(v, "100.70.42.52")
			if len(st_rel) > 0 {
				per.StBranch = st_rel[0].Branch
				per.StSha = st_rel[0].Sha
				per.StRd = st_rel[0].Rd
				per.StFlag = st_rel[0].Flag
			}
			ret = append(ret, per)
		} else {
			// 返回一条或多条数据，如果有多条单环境最多显示三条
			di_rel := CheckEnvRelInfo(v, "100.69.170.14")
			di_arr := GetRelArr(v, di_rel, nil)
			ret = append(ret, di_arr...)

			st_rel := CheckEnvRelInfo(v, "100.70.42.52")
			st_arr := GetRelArr(v, nil, st_rel)
			ret = append(ret, st_arr...)
		}
	}
	return ret
}

func CheckEnvRelInfo(v QueryInput, ip string) []EnvRet {
	req := httplib.Post(fmt.Sprintf("http://%s/mdeploy/v1/ext/online/query", ip) )
	req.Header("Authorization", "Basic mdeploy_IpFhvFjiQpV65PjIUywc3VHDjC0Wo9EM")
	req.Header("Content-Type", "application/json")
	req.SetTimeout(10*time.Second, 10*time.Second)
	req_data, _ := json.Marshal(v)
	req.Body(req_data)
	info_byte, err := req.Bytes()
	if err != nil {
		beego.Error(string(info_byte))
		beego.Error(err.Error())
		return nil
	}

	var ret DataRet
	err = json.Unmarshal(info_byte, &ret)
	if err != nil {
		beego.Error(string(info_byte))
		beego.Error(err.Error())
		return nil
	}
	if ret.Code == 0 {
		beego.Error(ret.Msg)
		return nil
	}
	return ret.Data
}

func GetRelArr(base QueryInput, di_rel, st_rel []EnvRet) []OutputData {
	var data []OutputData
	if di_rel != nil {
		for _, v := range di_rel {
			var per OutputData
			per.UnitEn = base.UnitEn
			per.UnitCn = ""
			per.Branch = base.Branch
			per.ShortSha = base.Sha
			per.Filename = base.Filename
			per.ReleaseDate = base.ReleaseDate
			per.DiBranch = v.Branch
			per.DiSha = v.Sha
			per.DiRd = v.Rd
			per.DiFlag = v.Flag
			if per.UnitEn == "" {
				per.UnitEn = v.UnitEn
				per.UnitCn = v.UnitCn
			} else {
				info := cfunc.GetUnitInfoByName(v.UnitEn)
				per.UnitCn = info.Name
			}
			data = append(data, per)
		}
	}

	if st_rel != nil {
		for _, v := range st_rel {
			var per OutputData
			per.UnitEn = base.UnitEn
			per.UnitCn = ""
			per.Branch = base.Branch
			per.ShortSha = base.Sha
			per.Filename = base.Filename
			per.ReleaseDate = base.ReleaseDate
			per.StBranch = v.Branch
			per.StSha = v.Sha
			per.StRd = v.Rd
			per.StFlag = v.Flag
			if per.UnitEn == "" {
				per.UnitEn = v.UnitEn
				per.UnitCn = v.UnitCn
			} else {
				info := cfunc.GetUnitInfoByName(v.UnitEn)
				per.UnitCn = info.Name
			}
			data = append(data, per)
		}
	}
	return data
}