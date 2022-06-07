package online

import (
	"strings"
	"controllers"
	"models"
	"initial"
	"fmt"
	"library/common"
	"github.com/astaxie/beego"
	"high-conc"
)

// @Title 数据库拉取文件目录
// @Description 数据库拉取文件目录
// @Param	online_id	    query	string	true	"数据库发布记录的自增id"
// @Success 200 true or false
// @Failure 403
// @router /db/pulldir [post]
func (c *DBOnlineController) DBPullDir() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	var online_db models.OnlineDbList
	var info models.UnitConfDb
	var base_info models.UnitConfList
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
	err = initial.DB.Model(models.UnitConfDb{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&info).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.UnitConfList{}).Where("id=? and is_offline=0", online.UnitId).First(&base_info).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	// 发布前判断
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的部署权限！")
		return
	}
	if online.IsSuccess != 10 {
		c.SetJson(0, "", "只有未发布的数据库才可以拉取git脚本文件！")
		return
	}
	if online_db.IsPullDir == 1 {
		c.SetJson(0, "", "文件已拉取，无需重复拉取！")
		return
	}
	if online_db.IsPullDir == 2 {
		c.SetJson(0, "", "文件正在拉取，请稍等！")
		return
	}

	opr := DBAgentOpr{
		Online: online,
		OnlineDetail: online_db,
		Conf: info,
		UnitInfo: base_info,
		Operator: c.UserId,
		Err: nil,
	}
	opr.UpdateDirInfo(map[string]interface{}{"is_pull_dir": 2})
	// 拉取文件，返回文件内容
	flag := opr.PullGitDirAct()
	pull_result := 0
	if flag {
		pull_result = 1
	}
	umap := map[string]interface{}{
		"is_pull_dir": pull_result,
		"dir_name": online_db.OnlineId,
	}
	opr.UpdateDirInfo(umap)
	if !flag {
		c.SetJson(0, "", "数据库脚本拉取失败，错误为：" + opr.Err.Error())
		return
	}
	c.SetJson(1, "", "数据库脚本拉取成功！")
}

// @Title 数据库部署
// @Description 数据库部署
// @Param	online_id	    query	string	true	"数据库发布记录的自增id"
// @Success 200 true or false
// @Failure 403
// @router /db/deploy [post]
func (c *DBOnlineController) DBDeploy() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	var online_db models.OnlineDbList
	var info models.UnitConfDb
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
	err = initial.DB.Model(models.UnitConfDb{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&info).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	// 发布前判断
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的部署权限！")
		return
	}
	if info.ConnResult != 1 {
		c.SetJson(0, "", "数据库信息不可信，请先在配置页面进行数据库连通性测试！")
		return
	}
	if online.IsSuccess != 10 {
		msg := ""
		if online.IsSuccess == 0 {
			msg = "发布失败，请进入详情页进行部署！"
		}
		if online.IsSuccess == 2 {
			msg = "正在发布，请稍后点击！"
		}
		if online.IsSuccess == 1 {
			msg = "发布成功，不允许再次点击！"
		}
		c.SetJson(0, "", msg)
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

	// 文件拉取操作
	if online_db.IsDirClear == 1 {
		c.SetJson(0, "", "数据库目录已经被清除，请重新创建上线单元！")
		return
	}
	db_dp := DBDeploy{
		OnlineList: online,
		Conf: info,
		Operator: c.UserId,
	}
	high_conc.JobQueue <- &db_dp
	c.SetJson(1, "", "数据库已进入队列，正在进行有序发布！")
}
