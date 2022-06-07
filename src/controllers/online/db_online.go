package online

import (
	"controllers"
	"strings"
	"fmt"
	"models"
	"initial"
	"github.com/astaxie/beego"
	"library/cfunc"
	"encoding/json"
	"time"
	"library/git"
	"library/common"
	"library/database"
	"path"
)

type DBOnlineController struct {
	controllers.BaseController
}

func (c *DBOnlineController) URLMapping() {
	c.Mapping("DBList", c.DBList)
	c.Mapping("DBSave", c.DBSave)
	c.Mapping("DBDel", c.DBDel)
	c.Mapping("DBDetail", c.DBDetail)
	c.Mapping("DBPullDir", c.DBPullDir)
	c.Mapping("DBDeploy", c.DBDeploy)
	c.Mapping("DBResultQuery", c.DBResultQuery)
	// 刷新sha值和结果
	c.Mapping("DBFreshSha", c.DBFreshSha)
	c.Mapping("DBFreshResult", c.DBFreshResult)
	// 单文件操作
	c.Mapping("RowSqlExec", c.RowSqlExec)
	c.Mapping("RowSqlInfo", c.RowSqlInfo)
	c.Mapping("RowSqlDel", c.RowSqlDel)
}

// @Title 获取数据库发布列表
// @Description 获取数据库发布列表
// @Param	unit_name	    query	string	false	"发布单元英文名，支持模糊搜索"
// @Param	online_date	query	string	false	"上线日期"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Param	quick	query	string	true	"快速选择，''/not_start/not_finish/fail"
// @Success 200 {object} models.OnlineAllList
// @Failure 403
// @router /db/list [get]
func (c *DBOnlineController) DBList() {
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
		DbList       models.OnlineDbList   `json:"db_list"`
		UnitCnName   string   `json:"unit_cn_name"`
		UnitEnName   string   `json:"unit_en_name"`
		OperatorName string   `json:"operator_name"`
	}

	var cnt int
	var olist []models.OnlineAllList
	err := initial.DB.Table("online_db_list a").Select("b.*").
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
		var dlist models.OnlineDbList
		initial.DB.Model(models.OnlineDbList{}).Where("online_id=? and is_delete=0", v.Id).First(&dlist)
		per := RetData{
			AllList: v,
			DbList: dlist,
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
// @router /db/save [post]
func (c *DBOnlineController) DBSave() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	var o_db OnlineCntrInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &o_db)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(o_db.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的添加权限，请联系相关发布单元的负责人、开发人员和测试人员发布！")
		return
	}
	// 校验
	var db_conf models.UnitConfDb
	err = initial.DB.Model(models.UnitConfDb{}).Where("is_delete=0 and unit_id=?", o_db.UnitId).First(&db_conf).Error
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
	if o_db.OnlineDate != "" {
		now_day = o_db.OnlineDate
	}
	if o_db.OnlineTime != "" {
		now_time = o_db.OnlineTime
	}
	// 赋初始值
	var online_main models.OnlineAllList
	var online_db models.OnlineDbList
	online_main.UnitId = db_conf.UnitId
	online_main.Branch = o_db.Branch
	//online_main.CommitId = ""
	//online_main.ShortCommitId = ""
	online_main.OnlineDate = now_day
	online_main.OnlineTime = now_time
	online_main.Version = o_db.OnlineDate
	online_main.IsProcessing = 0
	online_main.IsSuccess = 10   // 10表示未开始
	online_main.IsDelete = 0
	online_main.Operator = c.UserId
	online_main.ExcuteTime = ""
	online_main.InsertTime = now
	online_main.ErrorLog = ""

	online_db.IsDelete = 0
	online_db.DirName = ""
	online_db.IsDirClear = 0
	online_db.IsPullDir = 10

	if strings.Trim(o_db.Sha, " ") == "" {
		// 获取最新sha值
		branch_detail := git.GetBranchDetail(db_conf.GitId, o_db.Branch)
		if branch_detail == nil {
			c.SetJson(0, "", "分支输入有错，请重新输入！")
			return
		}
		online_main.CommitId = branch_detail.Commit.ID
		online_main.ShortCommitId = branch_detail.Commit.ShortID
	} else {
		commit_detail := git.GetCommitDetail(db_conf.GitId, o_db.Sha)
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
	online_db.OnlineId = online_main.Id
	err = tx.Create(&online_db).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "数据库上线单元创建成功！")
}

// @Title 删除标数据库应用发布列表
// @Description 删除标数据库应用发布列表
// @Param	online_id	    query	string	true	"数据库发布单元的id"
// @Success 200  true or false
// @Failure 403
// @router /db/del [post]
func (c *DBOnlineController) DBDel() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", online_id).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的删除权限！")
		return
	}
	if online.IsSuccess != 10 {
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
	err = tx.Model(models.OnlineDbList{}).Where("online_id=?", online_id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	// 详情也要清零，online_db_log
	err = tx.Model(models.OnlineDbLog{}).Where("online_id=?", online_id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()

	// 删除的时候清除目录
	var online_db models.OnlineDbList
	err = initial.DB.Model(models.OnlineDbList{}).Where("online_id=?", online_id).First(&online_db).Error
	if err != nil  {
		c.SetJson(0, "", err.Error())
		return
	}
	if online_db.IsPullDir != 10 {
		var conf models.UnitConfDb
		var base_info models.UnitConfList
		err = initial.DB.Model(models.UnitConfDb{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&conf).Error
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		err = initial.DB.Model(models.UnitConfList{}).Where("id=? and is_offline=0", online.UnitId).First(&base_info).Error
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		agent_opr := database.DBAgentOpr{
			DeployComp: conf.DeployComp,
		}
		if !agent_opr.GetAgentInfo() {
			c.SetJson(0, "", "agent信息获取失败！")
			return
		}
		base_dir := beego.AppConfig.String("file_base_dir")
		unit_en := strings.Replace(strings.ToLower(base_info.Unit), "_", "-", -1)
		rm_dir := path.Join(base_dir, conf.Type, unit_en, common.GetString(online_id) )
		err = agent_opr.RmAgentDir(rm_dir)
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	}

	c.SetJson(1, "", "数据删除成功！")
}

// @Title 删除标数据库应用发布列表
// @Description 删除标数据库应用发布列表
// @Success 200  true or false
// @Failure 403
// @router /db/detail/:id [get]
func (c *DBOnlineController) DBDetail() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id := c.Ctx.Input.Param(":id")
	search := c.GetString("sname")
	stype := c.GetString("quick")
	cond := fmt.Sprintf(" online_id=%s ", online_id)
	if strings.TrimSpace(search) != "" {
		cond += fmt.Sprintf(" and file_name like '%%%s%%' ", search)
	}
	var detail []models.OnlineDbLog
	var order_list []string
	switch stype {
		case "all":
			order_list = []string{"ddl", "trig", "pkg_type", "pkg_head", "pkg_body", "dml"}
		case "ddl":
			order_list = []string{"ddl"}
		case "dml":
			order_list = []string{"dml"}
		case "trig":
			order_list = []string{"trig"}
		case "pkg":
			order_list = []string{"pkg_type", "pkg_head", "pkg_body"}
		default:
			order_list = []string{"ddl", "trig", "pkg_type", "pkg_head", "pkg_body", "dml"}
	}
	for _, v := range order_list {
		order_str := "file_name asc"
		if common.InList(v, []string{"pkg_type", "pkg_head", "pkg_body"}) {
			order_str = "id asc"
		}
		var data []models.OnlineDbLog
		initial.DB.Model(models.OnlineDbLog{}).Where(cond + fmt.Sprintf(" and sql_type='%s' ", v)).Select("id, online_id, " +
			"file_name, file_sha, is_success, sql_type, execute_time, proxy_user, start_time, insert_time, is_delete, operator").
			Order(order_str).Find(&data)
		if len(data) > 0 {
			detail = append(detail, data...)
		}
	}

	var online models.OnlineAllList
	var online_db models.OnlineDbList
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", online_id).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.OnlineDbList{}).Where("online_id=? and is_delete=0", online_id).First(&online_db).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	// 获取online_db_log的数据，进行展示，如果 online.CommitId != log.CommitId，显示文件已删除
	if online_db.IsDirClear == 1 {
		for i:=0; i<len(detail); i++ {
			detail[i].IsDelete = 1
		}
	} else {
		for i:=0; i<len(detail); i++ {
			if detail[i].FileSha != online.CommitId {
				detail[i].IsDelete = 1
			}
		}
	}

	info := cfunc.GetUnitInfoById(online.UnitId)
	var conf models.UnitConfDb
	initial.DB.Model(models.UnitConfDb{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&conf)
	ret := map[string]interface{}{
		"dbname": info.Unit,
		"dbtype": conf.Type,
		"branch": online.Branch,
		"sha": online.ShortCommitId,
		"detail": detail,
		"cnt": len(detail),
	}
	c.SetJson(1, ret, "详情数据获取成功！")
}

// @Title 定时任务获取数据库发布结果
// @Description 定时任务获取数据库发布结果
// @Param	online_list	    query	string	ture	"数据库应用的上线id列表，比如`1,2,3`"
// @Success 200  true or false
// @Failure 403
// @router /db/result/query [get]
func (c *DBOnlineController) DBResultQuery() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_list := c.GetString("online_list")
	type Ret struct {
		OnlineId      int  `json:"online_id"`
		PullResult    int  `json:"pull_result"`
		DeployResult  int  `json:"deploy_result"`
	}
	online_arr := strings.Split(online_list, ",")
	var ret []Ret
	for _, v := range online_arr {
		if strings.TrimSpace(v) == "" {
			continue
		}
		var online_list models.OnlineAllList
		var db_list models.OnlineDbList
		err := initial.DB.Model(models.OnlineAllList{}).Where("id=?", v).First(&online_list).Error
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		err = initial.DB.Model(models.OnlineDbList{}).Where("online_id=?", v).First(&db_list).Error
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
		ret = append(ret, Ret{
			OnlineId: common.GetInt(v),
			PullResult: db_list.IsPullDir,
			DeployResult: online_list.IsSuccess,
		})
	}
	c.SetJson(1, ret, "数据库发布结果获取成功！")
}