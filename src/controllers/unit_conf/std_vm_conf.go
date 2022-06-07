package unit_conf

import (
	"controllers"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"models"
	"strings"
)

type StdVmConfController struct {
	controllers.BaseController
}

func (c *StdVmConfController) URLMapping() {
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("New", c.New)
	c.Mapping("Update", c.Update)
	c.Mapping("Delete", c.Delete)
	c.Mapping("JenkinsXmlConfirm", c.JenkinsXmlConfirm)
	c.Mapping("JenkinsXmlEdit", c.JenkinsXmlEdit)
}

// GetAll 方法
// @Title Get All
// @Description 获取所有发布单元列表
// @Param	search	query string false "搜索关键词，支持按发布单元中英文名、主机IP搜索，支持模糊搜索"
// @Param	corp	query	string	false	"按租户筛选"
// @Param	leader	query	string	false	"按开发负责人筛选"
// @Param	buildType	query	string	false	"按构建类型筛选"
// @Param	deployType	query	string	false	"按部署类型筛选"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} models.UnitConfVm
// @Failure 403
// @router /vm/list [get]
func (c *StdVmConfController) GetAll() {
	searchKey := c.GetString("search")
	corp := c.GetString("corp")
	leader := c.GetString("leader")
	buildType := c.GetString("buildType")
	deployType := c.GetString("deployType")
	page, rows := c.GetPageRows()
	queryStr := " a.is_delete = 0"
	if strings.TrimSpace(searchKey) != "" {
		queryStr += fmt.Sprintf(" and concat(b.unit, b.name, a.hosts) like '%%%s%%'", searchKey)
	}
	if strings.TrimSpace(corp) != "" {
		queryStr += fmt.Sprintf(" AND a.deploy_comp = '%s'", corp)
	}
	if strings.TrimSpace(leader) != "" {
		queryStr += fmt.Sprintf(" AND b.leader = '%s'", leader)
	}
	if strings.TrimSpace(buildType) != "" {
		queryStr += fmt.Sprintf(" AND a.app_sub_type = '%s'", buildType)
	}
	if strings.TrimSpace(deployType) != "" {
		queryStr += fmt.Sprintf(" AND a.deploy_type like '%%%s%%'", deployType)
	}

	type data struct {
		models.UnitConfVM
		UnitEnName	string	`json:"unit_en_name"`
		UnitCnName	string	`json:"unit_cn_name"`
		Leader		string	`json:"leader"`
		LeaderName	string	`json:"leader_name"`
	}

	var count int
	var unitList []data
	selectStr := fmt.Sprintf("a.*, b.unit as unit_en_name, b.name as unit_cn_name, b.leader")
	err := initial.DB.Table("unit_conf_vm a").Select(selectStr).
		Joins("left join unit_conf_list b on a.unit_id = b.id").Where(queryStr).Order("a.id desc").
		Count(&count).Offset((page - 1)*rows).Limit(rows).Find(&unitList).Error
	if err != nil {
		beego.Error(err)
		c.SetJson(0, "", "获取数据出错！")
		return
	}

	for i := 0; i < len(unitList); i ++ {
		unitList[i].LeaderName = cfunc.GetUserCnName(unitList[i].Leader)
	}

	resp := map[string]interface{}{
		"count": count,
		"data": unitList,
	}
	c.SetJson(1, resp, "获取数据成功！")
	return
}
