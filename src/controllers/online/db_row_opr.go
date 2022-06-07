package online

import (
	"strings"
	"models"
	"initial"
	"controllers"
	"library/common"
	"fmt"
	"github.com/astaxie/beego"
	"time"
)

// @Title 脚本单独执行，不使用并发
// @Description 脚本单独执行，不使用并发
// @Param	row_id	    query	string	true	"脚本记录的自增id"
// @Success 200 true or false
// @Failure 403
// @router /db/row/exec [post]
func (c *DBOnlineController) RowSqlExec() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	row_id, _ := c.GetInt("row_id")
	var online models.OnlineAllList
	var detail models.OnlineDbLog
	var info models.UnitConfDb
	err := initial.DB.Model(models.OnlineDbLog{}).Where("id=? and is_delete=0", row_id).First(&detail).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", detail.OnlineId).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.UnitConfDb{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&info).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	// 发布前判断，1权限判断，2状态判断，3时间判断
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的部署权限！")
		return
	}
	if detail.IsSuccess == 1 && (detail.SqlType == "ddl" || detail.SqlType == "dml") {
		// pkg和trig可以重复发布
		c.SetJson(0, "", "发布成功，不允许再次发布！")
		return
	}
	if detail.IsSuccess == 2 {
		c.SetJson(0, "", "正在发布，请稍后点击！")
		return
	}
	if detail.IsSuccess == 0 && detail.SqlType == "ddl" {
		c.SetJson(0, "", "ddl不能重复执行，需要将没有执行的语句抽离到新的脚本再次执行！")
		return
	}
	if online.OnlineTime != "" && beego.AppConfig.String("runmode") == "prd" {
		// 只有生产环境才验证发布时间
		ready := common.ReadyToRelease(online.OnlineTime, initial.DateSepLine)
		if !ready {
			c.SetJson(0, "", fmt.Sprintf("发布时间为%s,请在指定时间进行更新!", online.OnlineTime))
			return
		}
	}

	row_dp := DBDeploy{
		OnlineList: online,
		Conf: info,
		Operator: c.UserId,
	}
	detail_arr := []models.OnlineDbLog{detail}
	terr := row_dp.ExecuteSql(detail_arr, detail.SqlType, true)
	time.Sleep(100 * time.Millisecond)
	// 重新获取数据
	initial.DB.Model(models.OnlineDbLog{}).Where("id=? and is_delete=0", row_id).First(&detail)
	ret_code := 0
	ret_msg := "执行失败，请查看详情！"
	if detail.IsSuccess == 1 {
		ret_code = 1
		ret_msg = "执行成功"
	}
	if terr != nil {
		ret_msg = "执行失败，错误为：" + terr.Error()
	}
	c.SetJson(ret_code, "已执行", ret_msg)
}

// @Title 获取脚本的详细信息，包括内容、执行结果和git地址
// @Description 获取脚本的详细信息，包括内容、执行结果和git地址
// @Param	row_id	    query	string	true	"脚本记录的自增id"
// @Success 200  true or false
// @Failure 403
// @router /db/row/info [get]
func (c *DBOnlineController) RowSqlInfo() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	row_id := c.GetString("row_id")
	var detail models.OnlineDbLog
	err := initial.DB.Model(models.OnlineDbLog{}).Where("id=? and is_delete=0", row_id).First(&detail).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	// 获取脚本git地址
	var all_list models.OnlineAllList
	var db_info models.UnitConfDb
	initial.DB.Model(models.OnlineAllList{}).Where("id=?", detail.OnlineId).First(&all_list)
	initial.DB.Model(models.UnitConfDb{}).Where("unit_id=? and is_delete=0", all_list.UnitId).First(&db_info)
	base_url := strings.Split(db_info.GitUrl, ".git")[0]
	db_name_arr := strings.Split(base_url, "/")
	file_path_arr := strings.Split(detail.FilePath, db_name_arr[len(db_name_arr)-1])
	file_git_url := fmt.Sprintf("%s/blob/%s%s", base_url, detail.FileSha, file_path_arr[len(file_path_arr)-1])
	ret_map := map[string]interface{}{
		"detail": detail,
		"file_git_url": file_git_url,
	}

	c.SetJson(1, ret_map, "详情数据获取成功！")
}

// @Title 删除单个脚本
// @Description 删除单个脚本
// @Param	online_id	    query	string	true	"数据库发布单元的id"
// @Success 200  true or false
// @Failure 403
// @router /db/row/del [post]
func (c *DBOnlineController) RowSqlDel() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	row_id, _ := c.GetInt("row_id")
	var online models.OnlineAllList
	var detail models.OnlineDbLog
	err := initial.DB.Model(models.OnlineDbLog{}).Where("id=? and is_delete=0", row_id).First(&detail).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", detail.OnlineId).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的删除权限！")
		return
	}
	if detail.IsSuccess == 1 {
		c.SetJson(0, "", "脚本已执行成功，不允许删除！")
		return
	}

	tx := initial.DB.Begin()
	err = tx.Model(models.OnlineDbLog{}).Where("id=?", row_id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "脚本删除成功！")
}
