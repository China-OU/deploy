package ext

import (
	"initial"
	"models"
	"github.com/astaxie/beego"
	"fmt"
	"github.com/jinzhu/gorm"
	"library/cfunc"
	"strings"
	"encoding/json"
	"controllers/operation"
	"high-conc"
	"library/caas"
	"controllers"
	"library/harbor"
)

type ExtCntrDeployController struct {
	controllers.BaseUrlAuthController
}

func (c *ExtCntrDeployController) URLMapping() {
	c.Mapping("Search", c.Search)
	c.Mapping("DevopsUpgrade", c.DevopsUpgrade)
	c.Mapping("CpdsUpgrade", c.CpdsUpgrade)
	c.Mapping("Poll", c.Poll)
}

// @Title 根据service_name获取发布单元位置信息
// @Description 根据service_name获取发布单元位置信息
// @Param	service_name	query	string	true	"service_name，可模糊搜索"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	true	"调试模式"
// @Success 200 true or false
// @Failure 403
// @router /cntr/service/search [get]
func (c *ExtCntrDeployController) Search() {
	service := c.GetString("service_name")
	cond := fmt.Sprintf(" is_delete=0 and service_name like '%%%s%%' ", service)
	var cnt int
	var ulist []models.CaasConfDetail
	err := initial.DB.Model(models.CaasConfDetail{}).Where(cond).Count(&cnt).Limit(10).Find(&ulist).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	var ret_data []models.CaasDetailRet
	for _, v := range ulist {
		cc := cfunc.GetCompNetworkById(v.CaasId)
		ret_data = append(ret_data, models.CaasDetailRet{
			cc.DeployComp,
			cc.DeployNetwork,
			v,
		})
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": ret_data,
	}
	c.SetJson(1, ret, "信息获取成功！")
}

// @Title UpgradeService，需要传caas基础信息
// @Description 更新容器平台镜像，需要传caas基础信息
// @Param	body	body	ext.DevopsInput 	true	"body形式的数据，发布单元id名和镜像"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	true	"调试模式"
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/devops/upgrade [post]
func (c *ExtCntrDeployController) DevopsUpgrade() {
	if beego.AppConfig.String("runmode") == "prd" {
		c.SetJson(0, "", "生产暂不开放此权限!")
		return
	}
	if c.Role != "admin" {
		// 如果是受限用户，需要做权限认定
	}

	var input DevopsInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if input.ServiceId == 0 || strings.TrimSpace(input.Image) == "" || strings.TrimSpace(input.RecordId) == "" {
		c.SetJson(0, "", "输入参数不能为空！")
		return
	}

	unit_id := 0
	var detail models.CaasConfDetail
	err = initial.DB.Model(models.CaasConfDetail{}).Where("id=?", input.ServiceId).First(&detail).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	caas_conf := cfunc.GetCompNetworkById(detail.CaasId)
	if caas_conf.DeployComp == "" {
		c.SetJson(0, "", "租户信息有误！")
		return
	}

	var cnt int
	var unit_cntr models.UnitConfCntr
	err = initial.DB.Model(models.UnitConfCntr{}).Where("caas_team=? and caas_cluster=? and caas_stack=? " +
		"and service_name=? and deploy_comp=? and is_delete = 0", detail.TeamId, detail.ClusterUuid, detail.StackName,
		detail.ServiceName, caas_conf.DeployComp).Count(&cnt).First(&unit_cntr).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if cnt == 0 {
		unit_id = 1000000 + input.ServiceId
	}
	unit_id = unit_cntr.UnitId

	// 正在更新中的应用不允许再次更新
	cnt = 0
	initial.DB.Model(models.OprCntrUpgrade{}).Where("result = 2 and unit_id = ?", unit_id).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "镜像正在更新中，不允许再次点击！")
		return
	}
	// 校验镜像在harbor中是否存在
	err = harbor.HarborCheckImage(input.Image)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	// 初始化连接caas，获取相关数据
	opr := caas.CaasOpr{
		AgentConf: caas_conf,
		TeamId: detail.TeamId,
		ClustUuid: detail.ClusterUuid,
		StackName: detail.StackName,
		ServiceName: detail.ServiceName,
	}

	cntr_upgrade := operation.CntrUpgradeWithImage{
		Opr: opr,
		UnitId: unit_id,
		Image: input.Image,
		Operator: c.Username,
		SourceId: input.RecordId,
	}
	high_conc.JobQueue <- &cntr_upgrade

	c.SetJson(1, "", "镜像更新已成功进入队列，请耐心等待执行结果！")
}

type DevopsInput struct {
	ServiceId   int     `json:"service_id"`
	Image       string  `json:"image"`
	RecordId    string  `json:"record_id"'`          // devops唯一标致符是32位字符串类型，不是int型
}

// @Title 轮询接口，查询执行结果
// @Description 轮询接口，查询执行结果
// @Param	record_list	query	string	true	"记录列表，比如 aaabb,ccc,dd,ee"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	true	"调试模式"
// @Success 200 true or false
// @Failure 403
// @router /cntr/service/poll [get]
func (c *ExtCntrDeployController) Poll() {
	record_list := c.GetString("record_list")
	r_list := strings.Split(record_list, ",")
	
	type PollRet struct {
		RecordId    string    `json:"record_id"`
		Result      int       `json:"result"`
		Msg         string    `json:"msg"`
		Cost        int       `json:"cost"`
	}

	var ret []PollRet
	for _, v := range r_list {
		record_id := strings.Trim(v, " ")
		var per PollRet
		per.RecordId = record_id
		var cntr_upgrade models.OprCntrUpgrade
		err := initial.DB.Model(models.OprCntrUpgrade{}).Where("source_id=?", record_id).First(&cntr_upgrade).Error
		if err != nil {
			beego.Error(err.Error())
			per.Msg = err.Error()
			per.Result = 100
			ret = append(ret, per)
			continue
		}
		per.Result = cntr_upgrade.Result
		per.Msg = cntr_upgrade.Message
		per.Cost = cntr_upgrade.CostTime
		ret = append(ret, per)
	}

	c.SetJson(1, ret, "结果查询成功！")
}