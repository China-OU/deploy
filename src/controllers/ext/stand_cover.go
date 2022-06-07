package ext

import (
	"github.com/astaxie/beego"
	"initial"
	"strings"
)

// 标准化覆盖率统计
type StandCoverController struct {
	beego.Controller
}

func (c *StandCoverController) URLMapping() {
	c.Mapping("MdpStandUnit", c.MdpStandUnit)
}

func (c *StandCoverController) SetJson(code int, data interface{}, Msg string) {
	c.Data["json"] = map[string]interface{}{"code": code, "msg": Msg, "data": data}
	c.ServeJSON()
}

// @Title 获取标准化覆盖率
// @Description 获取标准化覆盖率
// @Success 200 true or false
// @Failure 403
// @router /mdp/stand [get]
func (c *StandCoverController) MdpStandUnit() {
	header := c.Ctx.Request.Header
	auth := ""
	if header["Authorization"] != nil && len(header["Authorization"]) > 0 {
		auth = header["Authorization"][0]
	} else {
		c.SetJson(0, "", "没有header!")
		return
	}
	if strings.Replace(auth, "Basic ", "", -1) != "mdeploy_IpFhvFjiQpV65SdeUywc3VHDjCAtDe9EM" {
		c.SetJson(0, "", "header校验失败!")
		return
	}

	var cv CoverData
	var mcp, vm, db []UnitNameList
	initial.DB.Table("unit_conf_mcp a").Select("b.unit").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").
		Where("a.is_delete = 0").Find(&mcp)
	initial.DB.Table("unit_conf_vm a").Select("b.unit").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").
		Where("a.is_delete = 0").Find(&vm)
	initial.DB.Table("unit_conf_db a").Select("b.unit").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").
		Where("a.is_delete = 0").Find(&db)
	for _, v := range mcp {
		cv.StdCntr = append(cv.StdCntr, v.Unit)
	}
	for _, v := range vm {
		cv.StdVm = append(cv.StdVm, v.Unit)
	}
	for _, v := range db {
		cv.StdDB = append(cv.StdDB, v.Unit)
	}
	c.SetJson(1, cv, "标准化进度数据获取成功!")
}

type CoverData struct {
	StdCntr    []string     `json:"std_cntr"`
	StdVm      []string     `json:"std_vm"`
	StdDB      []string     `json:"std_db"`
}

type UnitNameList struct {
	Unit string `gorm:"column:unit" json:"unit"`
}


