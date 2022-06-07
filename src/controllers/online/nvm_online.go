package online

import (
	"controllers"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"library/common"
	"models"
	"regexp"
	"strings"
	"time"
)

type NvmOnlineController struct {
	controllers.BaseController
}

func (c *NvmOnlineController) URLMapping() {
	c.Mapping("NvmOnlineList", c.NvmOnlineList)
	c.Mapping("NvmOnlineSave", c.NvmOnlineSave)
	c.Mapping("NvmOnlineDelete", c.NvmOnlineDelete)
	c.Mapping("NvmDeploy", c.NvmDeploy)
	c.Mapping("NvmShellLog", c.NvmShellLog)
	c.Mapping("NvmResultQuery", c.NvmResultQuery)
}

// @Title 获取非标虚机应用发布列表
// @Description 获取非标虚机应用发布列表
// @Param	unit_name   query	string	false	"发布单元英文名，支持模糊搜索"
// @Param	online_date	query	string	false	"上线日期"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Param	quick	query	string	false	"快速选择，''/not_start/not_finish/fail"
// @Success 200 {object} models.OnlineAllList
// @Failure 403
// @router /nvm/list [get]
func (c *NvmOnlineController) NvmOnlineList() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	unit_name := c.GetString("unit_name")
	online_date := c.GetString("online_date")
	quick := c.GetString("quick")
	page, rows := c.GetPageRows()
	cond := " a.is_delete = 0 and b.is_delete = 0 "
	if strings.TrimSpace(unit_name) != "" {
		cond += fmt.Sprintf(" and c.unit like '%%%s%%' ", unit_name)
	}
	if online_date != "" {
		cond += fmt.Sprintf(" and b.online_date = '%s' ", online_date)
	}
	if quick == "not_start" {
		cond += " and b.is_success = 10 "
	}
	if quick == "not_finish" {
		cond += " and b.is_success = 2 "
	}
	if quick == "fail" {
		cond += " and b.is_success = 0 "
	}

	// 组装数据
	type RetData struct {
		AllList      models.OnlineAllList  `json:"all_list"`
		NvmList      models.OnlineNvm      `json:"nvm_list"`
		UnitCnName   string   `json:"unit_cn_name"`
		UnitEnName   string   `json:"unit_en_name"`
		OperatorName string   `json:"operator_name"`
	}

	var cnt int
	var olist []models.OnlineAllList
	err := initial.DB.Table("online_nvm a").Select("b.*").
		Joins("left join online_all_list b ON a.online_id = b.id " +
			"left join unit_conf_list c ON b.unit_id = c.id").
		Where(cond).Count(&cnt).Order("a.id desc").Offset((page - 1)*rows).Limit(rows).Find(&olist).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	var data_ret []RetData
	for _, v := range olist {
		unit_info := cfunc.GetUnitInfoById(v.UnitId)
		var nvm_list models.OnlineNvm
		initial.DB.Model(models.OnlineNvm{}).Where("online_id=? and is_delete=0", v.Id).First(&nvm_list)
		nvm_list.ShellLog = ""
		per := RetData{
			AllList: v,
			NvmList: nvm_list,
			UnitCnName: unit_info.Name,
			UnitEnName: unit_info.Unit,
			OperatorName: cfunc.GetUserCnName(v.Operator),
		}
		data_ret = append(data_ret, per)
	}
	ret := map[string]interface{}{
		"cnt": cnt,
		"data": data_ret,
	}
	c.SetJson(1, ret, "数据获取成功！")
}

// @Title 新增非标虚机应用发布
// @Description 新增非标虚机应用发布
// @Param	body	body	online.OnlineNvmInput	true	"body形式的数据，非标虚机上线的基本元素"
// @Success 200 true or false
// @Failure 403
// @router /nvm/save [post]
func (c *NvmOnlineController) NvmOnlineSave() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	var nvm OnlineNvmInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &nvm)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	nvm.FileAddr = strings.TrimSpace(nvm.FileAddr)
	nvm.FileName = strings.TrimSpace(nvm.FileName)
	nvm.Sha256 = strings.TrimSpace(nvm.Sha256)
	nvm.IPList = strings.TrimSpace(nvm.IPList)
	// 校验
	var nvm_conf models.UnitConfNvm
	err = initial.DB.Model(models.UnitConfNvm{}).Where("is_delete=0 and unit_id=?", nvm.UnitId).First(&nvm_conf).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", "发布单元报错：" + err.Error())
		return
	}
	err = validNvmData(nvm, nvm_conf.Hosts)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(nvm.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的添加权限，请联系相关发布单元的负责人、开发人员和测试人员发布！")
		return
	}

	now := time.Now().Format(initial.DatetimeFormat)
	now_day := strings.Replace(now[0:10], "-", "", -1)
	now_time := now[11:16]
	if now_time < initial.DateSepLine {
		now_day = time.Now().AddDate(0, 0, -1).Format(initial.DateFormat)
	}
	if nvm.OnlineDate != "" {
		now_day = nvm.OnlineDate
	}
	if nvm.OnlineTime != "" {
		now_time = nvm.OnlineTime
	}
	// 赋初始值
	var online_main models.OnlineAllList
	var online_nvm models.OnlineNvm
	online_main.UnitId = nvm_conf.UnitId
	online_main.Branch = ""
	online_main.CommitId = ""
	online_main.ShortCommitId = ""
	online_main.OnlineDate = now_day
	online_main.OnlineTime = now_time
	online_main.Version = online_main.OnlineDate
	online_main.IsProcessing = 0
	online_main.IsSuccess = 10   // 10表示未开始
	online_main.IsDelete = 0
	online_main.Operator = c.UserId
	online_main.ExcuteTime = ""
	online_main.InsertTime = now
	online_main.ErrorLog = ""

	online_nvm.FileName = nvm.FileName
	online_nvm.FileAddr = nvm.FileAddr
	online_nvm.Sha256 = nvm.Sha256
	online_nvm.Host = nvm.IPList
	if online_nvm.Host == "" {
		online_nvm.Host = nvm_conf.Hosts
	}
	online_nvm.ShellLog = ""
	online_nvm.InsertTime = now
	online_nvm.IsDelete = 0

	// 录入
	tx := initial.DB.Begin()
	err = tx.Create(&online_main).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	online_nvm.OnlineId = online_main.Id
	err = tx.Create(&online_nvm).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "非标虚机上线单元创建成功！")
}

type OnlineNvmInput struct {
	UnitId      int     `json:"unit_id"`
	FileAddr    string  `json:"file_addr"`
	FileName    string  `json:"file_name"`
	Sha256      string  `json:"sha256"`
	IPList      string  `json:"ip_list"`
	OnlineDate  string  `json:"online_date"`
	OnlineTime  string  `json:"olnine_time"`
}

func validNvmData(nvm OnlineNvmInput, host string) error {
	match, _ := regexp.MatchString("^https://pan2021.cmrh.com/s/[0-9a-zA-Z]*/download$", nvm.FileAddr)
	if !match {
		return errors.New("非标虚机下载包地址有误，必须为 https://pan2021.cmrh.com/s/xxxxx/download 才能下载！")
	}
	match, _ = regexp.MatchString("^[0-9a-zA-Z]{8,64}$", nvm.Sha256)
	if !match {
		return errors.New("sha值必须为8到64位的大写、小写或数字！")
	}
	match, _ = regexp.MatchString("\\.(tar.gz|zip|jar|war)$", nvm.FileName)
	if !match {
		return errors.New("文件名只支持tar.gz|zip|jar|war四种格式！")
	}
	host_arr := strings.Split(host, ";")
	for _, v := range strings.Split(nvm.IPList, ";") {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if !common.CheckIp(v) {
			return errors.New("非标虚机的ip地址有误！")
		}
		if !common.InList(v, host_arr) {
			return errors.New("ip地址必须为非标虚机配置列表的子集！")
		}
	}
	return nil
}

// @Title 删除非标虚机应用发布列表
// @Description 删除非标虚机应用发布列表
// @Param	online_id	    query	string	true	"非标虚机发布单元的id"
// @Success 200  true or false
// @Failure 403
// @router /nvm/del [post]
func (c *NvmOnlineController) NvmOnlineDelete() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	var nvm models.OnlineNvm
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", online_id).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.OnlineNvm{}).Where("online_id=? and is_delete=0", online_id).First(&nvm).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的删除权限！")
		return
	}
	if online.IsSuccess == 1 || online.IsSuccess == 2 {
		c.SetJson(0, "", "执行成功或者正在执行中的任务，不允许删除！")
		return
	}

	tx := initial.DB.Begin()
	err = tx.Model(models.OnlineAllList{}).Where("id=?", online_id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	err = tx.Model(models.OnlineNvm{}).Where("online_id=?", online_id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "数据删除成功！")
}

// @Title 定时任务获取非标虚机应用发布结果
// @Description 定时任务获取非标虚机应用发布结果
// @Param	online_list	    query	string	ture	"非标虚机应用的上线列表，比如`1,2,3`"
// @Success 200  true or false
// @Failure 403
// @router /nvm/result/query [get]
func (c *NvmOnlineController) NvmResultQuery() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_list := c.GetString("online_list")
	type Ret struct {
		OnlineId       int  `json:"online_id"`
		DeployResult   int  `json:"deploy_result"`
	}
	online_arr := strings.Split(online_list, ",")
	var ret []Ret
	for _, v := range online_arr {
		if strings.TrimSpace(v) == "" {
			continue
		}
		var online models.OnlineAllList
		var nvm models.OnlineNvm
		err := initial.DB.Model(models.OnlineAllList{}).Where("id=?", v).First(&online).Error
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		err = initial.DB.Model(models.OnlineNvm{}).Where("online_id=?", v).First(&nvm).Error
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		ret = append(ret, Ret{
			OnlineId: common.GetInt(v),
			DeployResult: online.IsSuccess,
		})
	}
	c.SetJson(1, ret, "非标虚机应用发布结果获取成功！")
}
