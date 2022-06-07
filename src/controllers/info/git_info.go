package info

import (
	"controllers"
	"library/git"
)

type GitInfoController struct {
	controllers.BaseController
}

func (c *GitInfoController) URLMapping() {
	c.Mapping("SearchInfo", c.SearchInfo)
	c.Mapping("GetBranchList", c.GetBranchList)
	//c.Mapping("GetCommitList", c.GetCommitList)
}

// @Title git gitlab info by info, like aml/amldb
// @Description 获取git信息，最多获取10条
// @Param	search	query	string	true	"git搜索字段"
// @Success 200 {object} []models.SimpleProject
// @Failure 403
// @router /search [get]
func (c *GitInfoController) SearchInfo() {
	search := c.GetString("search")
	git_prj_list := git.SearchByNamespace(search)
	c.SetJson(1, git_prj_list, "git信息获取成功！")
}

// @Title 获取git最近一年的分支列表，按时间降序排
// @Description 获取git最近一年的分支列表，按时间降序排
// @Param	git_id	query	string	true	"git的id"
// @Success 200 {object} []models.SimpleBranch
// @Failure 403
// @router /branch/list [get]
func (c *GitInfoController) GetBranchList() {
	git_id := c.GetString("git_id")
	bl := git.GetAllBranchList(git_id)
	c.SetJson(1, bl, "分支列表获取成功！")
}
