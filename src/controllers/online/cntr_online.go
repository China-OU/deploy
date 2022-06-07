package online

import (
	"controllers"
	"controllers/operation"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cfunc"
	"library/common"
	"library/git"
	"models"
	"strings"
	"time"
)

type StdCntrOnlineController struct {
	controllers.BaseController
}

func (c *StdCntrOnlineController) URLMapping() {
	c.Mapping("CntrList", c.CntrList)
	c.Mapping("CntrSave", c.CntrSave)
	c.Mapping("CntrDel", c.CntrDel)
	c.Mapping("CntrBuild", c.CntrBuild)
	c.Mapping("CntrBuildLog", c.CntrJenkLog)
	c.Mapping("CntrUpgrade", c.CntrUpgrade)
	c.Mapping("CntrResultQuery", c.CntrResultQuery)
}

// @Title 获取标准容器应用发布列表
// @Description 获取标准容器应用发布列表
// @Param	unit_id	    query	string	false	"发布单元的id"
// @Param	online_date	query	string	false	"上线日期"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Param	quick	query	string	false	"快速选择，''/not_start/not_finish/fail"
// @Success 200 {object} models.OnlineAllList
// @Failure 403
// @router /cntr/list [get]
func (c *StdCntrOnlineController) CntrList() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	unit_id := c.GetString("unit_id")
	online_date := c.GetString("online_date")
	quick := c.GetString("quick")
	page, rows := c.GetPageRows()
	cond := " a.is_delete = 0 and b.is_delete = 0 "
	if strings.TrimSpace(unit_id) != "" {
		cond += fmt.Sprintf(" and b.unit_id = '%s' ", unit_id)
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

	var cnt int
	var olist []models.OnlineAllList
	err := initial.DB.Table("online_std_cntr a").Select("b.*").
		Joins("left join online_all_list b ON a.online_id = b.id").
		Where(cond).Count(&cnt).Order("a.id desc").Offset((page - 1)*rows).Limit(rows).Find(&olist).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	// 组装数据
	type RetData struct {
		AllList      models.OnlineAllList  `json:"all_list"`
		StdCntr      models.OnlineStdCntr  `json:"std_cntr"`
		CntrUpgrade  models.McpUpgradeList `json:"cntr_upgrade"`
		UnitCnName   string   `json:"unit_cn_name"`
		UnitEnName   string   `json:"unit_en_name"`
		OperatorName string   `json:"operator_name"`
	}
	var data_ret []RetData
	for _, v := range olist {
		unit_info := cfunc.GetUnitInfoById(v.UnitId)
		var cntr_info models.OnlineStdCntr
		var cntr_upgrade models.McpUpgradeList
		initial.DB.Model(models.OnlineStdCntr{}).Where("online_id=? and is_delete=0", v.Id).First(&cntr_info)
		// 判断从哪个表取数据
		_, cconf :=  operation.GetCntrConfig(common.GetString(v.UnitId))
		if cconf.McpConfId > 0 {
			initial.DB.Model(models.McpUpgradeList{}).Where("id=?", cntr_info.OprCntrId).First(&cntr_upgrade)
			if cntr_upgrade.UnitId == 0 {
				// 早期的数据还是取自于 opr_cntr_upgrade
				initial.DB.Table("opr_cntr_upgrade").Where("id=?", cntr_info.OprCntrId).First(&cntr_upgrade)
			}
		} else {
			initial.DB.Table("opr_cntr_upgrade").Where("id=?", cntr_info.OprCntrId).First(&cntr_upgrade)
		}
		per := RetData{
			AllList: v,
			StdCntr: cntr_info,
			CntrUpgrade: cntr_upgrade,
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

// @Title 新增标准容器应用发布
// @Description 新增标准容器应用发布
// @Param	body	body	online.OnlineCntrInput	true	"body形式的数据，标准容器上线的基本元素"
// @Success 200 true or false
// @Failure 403
// @router /cntr/save [post]
func (c *StdCntrOnlineController) CntrSave() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	var o_cntr OnlineCntrInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &o_cntr)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(o_cntr.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的添加权限，请联系相关发布单元的负责人、开发人员和测试人员发布！")
		return
	}
	// 校验
	var cntr_conf models.UnitConfCntr
	err = initial.DB.Model(models.UnitConfCntr{}).Where("is_delete=0 and unit_id=?", o_cntr.UnitId).First(&cntr_conf).Error
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", "发布单元报错：" + err.Error())
		return
	}

	now := time.Now().Format(initial.DatetimeFormat)
	now_day := strings.Replace(now[0:10], "-", "", -1)
	now_time := now[11:16]
	if now_time < initial.DateSepLine {
		now_day = time.Now().AddDate(0, 0, -1).Format(initial.DateFormat)
	}
	if o_cntr.OnlineDate != "" {
		now_day = o_cntr.OnlineDate
	}
	if o_cntr.OnlineTime != "" {
		now_time = o_cntr.OnlineTime
	}
	// 赋初始值
	var online_main models.OnlineAllList
	var online_cntr models.OnlineStdCntr
	online_main.UnitId = cntr_conf.UnitId
	online_main.Branch = o_cntr.Branch
	//online_main.CommitId = ""
	//online_main.ShortCommitId = ""
	online_main.OnlineDate = now_day
	online_main.OnlineTime = now_time
	online_main.Version = o_cntr.OnlineDate
	online_main.IsProcessing = 0
	online_main.IsSuccess = 10   // 10表示未开始
	online_main.IsDelete = 0
	online_main.Operator = c.UserId
	online_main.ExcuteTime = ""
	online_main.InsertTime = now
	online_main.ErrorLog = ""

	//online_cntr.OnlineId = 0
	online_cntr.JenkinsSuccess = 10
	online_cntr.IsDelete = 0
	online_cntr.InsertTime = now


	if strings.Trim(o_cntr.Sha, " ") == "" {
		// 获取最新sha值
		branch_detail := git.GetBranchDetail(cntr_conf.GitId, o_cntr.Branch)
		if branch_detail == nil {
			c.SetJson(0, "", "分支输入有错，请重新输入！")
			return
		}
		online_main.CommitId = branch_detail.Commit.ID
		online_main.ShortCommitId = branch_detail.Commit.ShortID
	} else {
		commit_detail := git.GetCommitDetail(cntr_conf.GitId, o_cntr.Sha)
		if commit_detail == nil {
			c.SetJson(0, "", "sha值输入有错，请重新输入！")
			return
		}
		online_main.CommitId = commit_detail.ID
		online_main.ShortCommitId = commit_detail.ShortID
	}

	// 录入
	tx := initial.DB.Begin()
	err = tx.Create(&online_main).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	online_cntr.OnlineId = online_main.Id
	err = tx.Create(&online_cntr).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "标准容器上线单元创建成功！")
}

type OnlineCntrInput struct {
	UnitId   int     `json:"unit_id"`
	Branch   string  `json:"branch"`
	Sha      string  `json:"sha"`
	OnlineDate  string `json:"online_date"`
	OnlineTime  string `json:"olnine_time"`
}

// @Title 删除标准容器应用发布列表
// @Description 删除标准容器应用发布列表
// @Param	online_id	    query	string	true	"标准容器发布单元的id"
// @Success 200  true or false
// @Failure 403
// @router /cntr/del [post]
func (c *StdCntrOnlineController) CntrDel() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	var cntr_info models.OnlineStdCntr
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", online_id).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.OnlineStdCntr{}).Where("online_id=? and is_delete=0", online_id).First(&cntr_info).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的删除权限！")
		return
	}
	flag := false
	// 各种情况。             is_success      jenkins_success   opr_cntr_id
	// 1、未构建。                      10                10              0
	// 2、构建中                        2                 2               0
	// 3、构建成功，未发布。             2                1               0
	// 4、构建不成功。                  2                 0              0
	// 5、构建成功，发布失败。          0                 1              num
	// 6、发布成功                     1                 1              num
	// 其中 1, 3, 4 都可以删除
	if cntr_info.OprCntrId == 0 && cntr_info.JenkinsSuccess != 2 {
		flag = true
	}
	if flag == false {
		c.SetJson(0, "", "已执行完成或者执行中的任务，不允许删除！")
		return
	}

	tx := initial.DB.Begin()
	err = tx.Model(models.OnlineAllList{}).Where("id=?", online_id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	err = tx.Model(models.OnlineStdCntr{}).Where("online_id=?", online_id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "数据删除成功！")
}

// @Title 定时任务获取标准容器应用发布结果
// @Description 定时任务获取标准容器应用发布结果
// @Param	online_list	    query	string	ture	"标准容器应用的上线列表，比如`1,2,3`"
// @Success 200  true or false
// @Failure 403
// @router /cntr/result/query [get]
func (c *StdCntrOnlineController) CntrResultQuery() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_list := c.GetString("online_list")
	type Ret struct {
		OnlineId      int  `json:"online_id"`
		BuildResult   int  `json:"build_result"`
		UpgradeResult int  `json:"upgrade_result"`
	}
	online_arr := strings.Split(online_list, ",")
	var ret []Ret
	for _, v := range online_arr {
		if strings.TrimSpace(v) == "" {
			continue
		}
		var online models.OnlineAllList
		var online_cntr models.OnlineStdCntr
		initial.DB.Model(models.OnlineAllList{}).Where("id=?", v).First(&online)
		err := initial.DB.Model(models.OnlineStdCntr{}).Where("online_id=?", v).First(&online_cntr).Error
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		build_result := online_cntr.JenkinsSuccess
		upgrade_result := 10
		if online_cntr.OprCntrId > 0 {
			_, cconf :=  operation.GetCntrConfig(common.GetString(online.UnitId))
			if cconf.McpConfId > 0 {
				var opr_cntr models.McpUpgradeList
				err := initial.DB.Model(models.McpUpgradeList{}).Where("id=?", online_cntr.OprCntrId).First(&opr_cntr).Error
				if err != nil {
					c.SetJson(0, "", err.Error())
					return
				}
				upgrade_result = opr_cntr.Result
			} else {
				var old_upgrade models.OprCntrUpgrade
				err := initial.DB.Model(models.OprCntrUpgrade{}).Where("id=?", online_cntr.OprCntrId).First(&old_upgrade).Error
				if err != nil {
					c.SetJson(0, "", err.Error())
					return
				}
				upgrade_result = old_upgrade.Result
			}
		}
		ret = append(ret, Ret{
			OnlineId: common.GetInt(v),
			BuildResult: build_result,
			UpgradeResult: upgrade_result,
		})
	}
	c.SetJson(1, ret, "标准容器应用发布结果获取成功！")
}
