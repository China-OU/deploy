package online

import (
	"strings"
	"models"
	"initial"
	"controllers"
	"library/git"
	"encoding/json"
)

// @Title 数据库部署更新sha值，只能用于同分支的sha值更新
// @Description 数据库部署更新sha值，只能用于同分支的sha值更新
// @Param	online_id  query	string	true	"数据库发布记录的自增id"
// @Param	sha  query	string	true	"对应分支的sha值"
// @Success 200 true or false
// @Failure 403
// @router /db/fresh/sha [post]
func (c *DBOnlineController) DBFreshSha() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	type InputData struct {
		Branch    string  `json:"branch"`
		OnlineId  string  `json:"online_id"`
		Sha       string  `json:"sha"`
	}
	var input InputData
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	var online models.OnlineAllList
	var online_db models.OnlineDbList
	var info models.UnitConfDb
	var base_info models.UnitConfList
	err = initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", input.OnlineId).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if len(input.Sha) > 6 && strings.Contains(online.CommitId, input.Sha) {
		c.SetJson(0, "", "sha值没有变化，无需更新！")
		return
	}
	err = initial.DB.Model(models.OnlineDbList{}).Where("online_id=? and is_delete=0", input.OnlineId).First(&online_db).Error
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
	if online.IsSuccess == 2 {
		c.SetJson(0, "", "脚本正在执行中，不能更新sha值！")
		return
	}
	if online_db.IsPullDir == 2 {
		c.SetJson(0, "", "文件正在拉取，不能更新sha值！")
		return
	}

	// 校验sha值对不对
	commit_detail := git.GetCommitDetail(info.GitId, input.Sha)
	if commit_detail == nil {
		c.SetJson(0, "", "sha值输入有错，请重新输入！")
		return
	}

	opr := DBAgentOpr{
		Online: online,
		OnlineDetail: online_db,
		Conf: info,
		UnitInfo: base_info,
		Operator: c.UserId,
		Err: nil,
		FreSha: commit_detail.ID,
	}
	opr.UpdateDirInfo(map[string]interface{}{"is_pull_dir": 2})
	// 拉取文件，返回文件内容
	flag := opr.FreshGitDir()
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
	if !opr.UpdateAllListSha(commit_detail.ShortID) {
		c.SetJson(0, "", "主列表更新sha值出现错误，sha值更新失败！")
		return
	}
	c.SetJson(1, "", "sha值更新成功！")
}

// @Title 子记录全部执行成功，主记录也要更新成功
// @Description 子记录全部执行成功，主记录也要更新成功
// @Param	online_id   query	string	ture	"数据库应用的上线id"
// @Success 200  true or false
// @Failure 403
// @router /db/fresh/result [post]
func (c *DBOnlineController) DBFreshResult() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id := c.GetString("online_id")
	var online_list models.OnlineAllList
	var cnt int
	var cnt_success int
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", online_id).First(&online_list).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	// 有错误或未执行的脚本，不更改状态
	err = initial.DB.Model(models.OnlineDbLog{}).Where("online_id=? and is_delete=0 and file_sha=? and is_success in (0, 2, 10)", online_id,
		online_list.CommitId).Count(&cnt).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if cnt > 0 {
		c.SetJson(0, "", "有脚本没有发布成功，状态不变！")
		return
	}
	// 无执行成功的脚本，不允许更改
	err = initial.DB.Model(models.OnlineDbLog{}).Where("online_id=? and is_delete=0 and file_sha=? and is_success=1", online_id,
		online_list.CommitId).Count(&cnt_success).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if cnt_success == 0 {
		c.SetJson(0, "", "没有发布成功的脚本，状态不变！")
		return
	}

	// 更新发布结果
	tx := initial.DB.Begin()
	err = tx.Model(models.OnlineAllList{}).Where("id=?", online_id).Update("is_success", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	c.SetJson(1, "", "数据库发布结果刷新成功！")
}

