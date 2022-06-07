package caas_cntr

import (
	"controllers/operation"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/jinzhu/gorm"
	high_conc "high-conc"
	"initial"
	"library/caas"
	"library/harbor"
	"models"
	"regexp"
	"strings"
	"time"
)

type CntrConfig struct {
	Cpu      int
	MemLimit int
}

// @Title EditService
// @Description 更新堆栈服务
// @Param	body	body	caas.InitServiceWebData 	true	"json body"
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/init [put]
func (c *CntrController) ReInitService() {
	c.IsEdit = true
	c.InitService()
}

// @Title InitService
// @Description 新增堆栈服务
// @Param	body	body	caas.InitServiceWebData 	true	"json body"
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/init [post]
func (c *CntrController) InitService() {
	runMode := beego.AppConfig.String("runmode")
	var input caas.InitServiceWebData
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}

	input.ServiceName = strings.ToLower(input.ServiceName)
	input.ServiceName = strings.Replace(input.ServiceName, "_", "-", -1)
	if len(input.ServiceName) > 30 {
		c.SetJson(0, "", "服务名错误：长度不允许超过30个字符")
		return
	}
	serviceNameRe := `^[a-z0-9-]{3,50}$`
	if matched, err := regexp.Match(serviceNameRe, []byte(input.ServiceName)); err != nil {
		c.SetJson(0, "", "regexp compile err：服务名错误")
		return
	} else if !matched {
		c.SetJson(0, "", "服务名只允许用小写字母和中划线")
		return
	}
	if runMode == "prd" {
		if input.InstanceNum > 5 || input.InstanceNum < 2 {
			c.SetJson(0, "", "prd实例数标配为2~5个")
			return
		}
	} else {
		if input.InstanceNum > 2 {
			c.SetJson(0, "", "实例数不能超过2")
			return
		}
	}
	// CPU和内存限制已去掉
	// 环境变量配置检查：推荐大写，名称必须以下划线分开，不能使用其他字符
	if !input.Environment.OK() {
		c.SetJson(0, "", "环境变量只允许大小写、中下划线和数字，不支持其它格式")
		return
	}
	// 故障恢复策略检查
	if err := input.HealthCheck.OK(); err != nil {
		c.SetJson(0, "", "健康检查"+err.Error())
		return
	}
	// 卷类型检查
	for _, v := range input.Volume {
		if err := v.OK(); err != nil {
			c.SetJson(0, "", "卷存储配置："+err.Error())
			return
		}
	}
	// 权限校验
	unit, err := models.UnitConfList{}.GetOneById(input.UnitId)
	if err != nil {
		c.SetJson(0, "", "发布单元不存在："+err.Error())
		return
	}
	// 权限校验
	isDeploySingle := true
	if strings.Contains(c.Role, "admin") {
		isDeploySingle = false
	}
	if runMode == "prd" {
		if isDeploySingle {
			c.SetJson(0, "", "生产环境只有管理员能操作")
			return
		}
	} else if isDeploySingle && !strings.Contains(unit.Leader, c.UserId) { // 非生产环境
		c.SetJson(0, "", "非生产环境只有管理员和开发负责人能操作")
		return
	}
	var cnt int
	err = initial.DB.Model(&models.OprCntrInit{}).Where("is_delete = 0 and unit_id = ?", input.UnitId).Count(&cnt).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", "数据库错误："+err.Error())
		return
	}
	if cnt > 0 && !c.IsEdit { // 可再次编辑，但不能重新创建
		c.SetJson(0, "", "唯一性错误：请在原记录上进行初始化或编辑！")
		return
	}
	err, agentConf := operation.GetCaasConfig(input.Comp)
	// 唯一性校验2:存在于容器服务详情表中的不能再创建,可以再次编辑
	cnt = 0
	err = initial.DB.Model(&models.CaasConfDetail{}).
		Where("caas_id=? and team_id=? "+
			"and cluster_uuid=? and stack_name=? "+
			"and service_name=? and is_delete=0",
			agentConf.Id, input.TeamId, input.ClusterUuid,
			input.StackName, input.ServiceName).
		Count(&cnt).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", "数据库错误："+err.Error())
		return
	}
	if !c.IsEdit && cnt > 0 {
		c.SetJson(0, "", "该服务可能在caas已经存在，请联系Deploy！")
		return

	}
	// 校验镜像在harbor中是否存在
	err = harbor.HarborCheckImage(input.Image)
	if err != nil {
		beego.Info(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	opr := caas.CaasOpr{
		AgentConf:   agentConf,
		TeamId:      input.TeamId,
		ClustUuid:   input.ClusterUuid,
		StackName:   input.StackName,
		ServiceName: input.ServiceName,
	}
	cntrInit := CntrInit{
		Opr:      opr,
		Operator: c.UserId,
		WebData:  input,
	}
	input.Scheduler = nil
	defaultScheduler := models.Scheduler{
		Policy:   "mustNot",
		Selector: "node",
		Key:      "bad",
		Value:    "true",
	}
	input.Scheduler = append(input.Scheduler, &defaultScheduler)
	if c.IsEdit {
		cntrEdit := CntrEdit{
			Opr:      opr,
			Operator: c.UserId,
			WebData:  input,
		}
		high_conc.JobQueue <- &cntrEdit
		c.SetJson(1, "", "服务配置更新中！")
		return
	}
	high_conc.JobQueue <- &cntrInit
	c.SetJson(1, "", "服务初始化中！")
}

func CpuMemConf() map[int]CntrConfig {
	cntrConfigMap := make(map[int]CntrConfig, 6)
	cntrConfigMap[1] = CntrConfig{1, 1024}
	cntrConfigMap[2] = CntrConfig{1, 2048}
	cntrConfigMap[3] = CntrConfig{2, 2048}
	cntrConfigMap[4] = CntrConfig{2, 4096}
	cntrConfigMap[5] = CntrConfig{4, 4096}
	cntrConfigMap[6] = CntrConfig{4, 8192}
	return cntrConfigMap
}

func (c *CntrConfig) String() string {
	return fmt.Sprintf("%dC %dG", c.Cpu, c.MemLimit/1024)
}

// @Title ListInitServiceTask
// @Description 获取堆栈初始化任务记录
// @Param search query string false "搜索关键字,按服务名搜索"
// @Param limit query int false "记录数"
// @Param page query int false "页码"
// @Success 200 {object} models.OprCntrInit
// @Failure 403
// @router /cntr/init/list [get]
func (c *CntrController) ListCntrInit() {
	page, _ := c.GetInt("page", 1)
	limit, _ := c.GetInt("limit", 10)
	search := c.GetString("search", "")
	isAdmin := false
	if strings.Contains(c.Role, "admin") {
		isAdmin = true
	}
	total, initList, err := models.OprCntrInit{}.List(isAdmin, c.UserId, search, page, limit)
	if err != nil {
		c.SetJson(0, "", "数据错误:"+err.Error())
		return
	}
	resData := map[string]interface{}{
		"total":    total,
		"initList": initList,
	}
	c.SetJson(1, resData, "")
}

// @Title GetInitServiceTask
// @Description 获取堆栈初始化任务记录
// @Param id query int false "记录ID"
// @Success 200 {object} models.OprCntrInit
// @Failure 403
// @router /cntr/init [get]
func (c *CntrController) GetCntrInit() {
	id, _ := c.GetInt("id", 0)
	if id == 0 {
		c.SetJson(0, "", "id 错误")
		return
	}
	isAdmin := false
	if strings.Contains(c.Role, "admin") {
		isAdmin = true
	}
	item, err := models.OprCntrInit{}.GetOneById(isAdmin, c.UserId, id)
	if err != nil {
		c.SetJson(0, "", "记录不存在或权限不足:"+err.Error())
		return
	}
	c.SetJson(1, item, "")
}

// @Title ListAppUnit
// @Description 获取app-unit列表，初始化容器使用
// @Param search query string false "搜索关键字,按unit英文名搜索"
// @Param limit query int false "记录数"
// @Param page query int false "页码"
// @Param appType query string false "应用类型，app（默认）或 web"
// @Success 200 {object} models.UnitConfList
// @Failure 403
// @router /cntr/init/unit-list [get]
func (c *CntrController) ListAppUnit() {
	uid := c.UserId
	if strings.Contains(c.Role, "admin") {
		uid = ""
	}
	page, _ := c.GetInt("page", 1)
	limit, _ := c.GetInt("limit", 10)
	search := c.GetString("search", "")
	appType := c.GetString("appType", "app")
	total, unitList, err := models.UnitConfList{}.Find(search, uid, appType, page, limit)
	if err != nil {
		c.SetJson(0, "", "数据错误:"+err.Error())
		return
	}
	resData := map[string]interface{}{
		"total":    total,
		"initList": unitList,
	}
	c.SetJson(1, resData, "")
}

// @Title DeleteInitRecord
// @Description 删除初始化记录
// @Param id query int true	"记录ID"
// @Success 200 {object} {}
// @Failure 403
// @router /cntr/init [delete]
func (c *CntrController) DeleteInitRecord() {
	id, _ := c.GetInt("id", 0)
	if id == 0 {
		c.SetJson(0, "", "ID不能为0")
		return
	}
	// 权限校验
	if !strings.Contains(c.Role, "admin") {
		c.SetJson(0, "", "权限不足，请联系管理员！")
		return
	}

	var initLog models.OprCntrInit
	if err := initial.DB.Table("opr_cntr_init").Select(`id, is_delete`).First(&initLog, id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.SetJson(0, "", "记录不存在！")
			return
		}
		c.SetJson(0, "", err.Error())
		return
	}
	tx := initial.DB.Begin()
	if err := tx.Model(&initLog).Updates(map[string]interface{}{
		"is_delete": true,
	}).Error; err != nil {
		tx.Rollback()
		c.SetJson(0, "", "删除失败："+err.Error())
		return
	}
	if err := c.insertOprRecord(tx, "删除记录", "1", "0", id); err != nil {
		tx.Rollback()
		c.SetJson(0, "", "删除失败："+err.Error())
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.SetJson(0, "", "删除失败："+err.Error())
		return
	}
	c.SetJson(1, "", "删除成功！")

}

func (c *CntrController) insertOprRecord(tx *gorm.DB, oprAction, newVal, oldVal string, CntrConfigId int) error {
	var log models.OprCntrLog
	log.Operator = c.UserId
	log.InsertTime = time.Now().Format(initial.DatetimeFormat)
	log.RelTable = "cntr_init_log"
	log.RelId = CntrConfigId
	log.OprAction = oprAction
	log.OldVal = oldVal
	log.NewVal = newVal
	return tx.Create(&log).Error
}

// @Title ListOprLog
// @Description 获取初始化操作列表，
// @Param search query string false "搜索关键字,服务名搜索"
// @Param limit query int false "记录数"
// @Param page query int false "页码"
// @Success 200 {object} models.UnitConfList
// @Failure 403
// @router /cntr/init/opr-log [get]
func (c *CntrController) ListOprLog() {
	page, _ := c.GetInt("page", 1)
	limit, _ := c.GetInt("limit", 10)
	search := c.GetString("search", "")
	total, logList, err := models.OprCntrLog{}.Find(search, page, limit)
	if err != nil {
		c.SetJson(0, "", "数据错误:"+err.Error())
		return
	}
	resData := map[string]interface{}{
		"total":    total,
		"initList": logList,
	}
	c.SetJson(1, resData, "")
}
