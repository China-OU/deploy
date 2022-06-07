package git

import (
	"models"
	"github.com/xanzy/go-gitlab"
	"time"
	"github.com/astaxie/beego"
	"sort"
	"initial"
	"library/common"
)

// 获取全部分支列表
func GetAllBranchList(git_id string) []models.SimpleBranch {
	gitToken := beego.AppConfig.String("gitToken")
	git := gitlab.NewClient(nil, gitToken)
	git.SetBaseURL("http://git.dev.cmrh.com/api/v4/")
	p := &gitlab.ListBranchesOptions{}
	p.PerPage = 50
	p.Page = 1
	now := time.Now().AddDate(0, -6, 0).Format(initial.DatetimeFormat)

	var bl []models.SimpleBranch
	for {
		branch_list, resp, err := git.Branches.ListBranches(git_id, p)
		if err != nil {
			beego.Error(err)
			return []models.SimpleBranch{}
		}
		beego.Info(len(branch_list))
		beego.Info(resp.TotalPages)
		// 去掉半年之前的分支，保留三个特殊分支
		for _, v := range branch_list {
			if v.Commit.CreatedAt.Format(initial.DatetimeFormat) > now || common.InList(v.Name, []string{"master", "release", "dev"}) {
				bl = append(bl, models.SimpleBranch{
					Name: v.Name,
					AuthorName: v.Commit.AuthorName,
					CreatedAt: v.Commit.CreatedAt.Format(initial.DatetimeFormat),
				})
			}
		}
		// 退出条件
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		p.Page = resp.NextPage
	}

	sort.Sort(models.SimpleBranchList(bl))
	return bl
}

// 根据sha值获取提交详情
func GetCommitDetail(git_id, sha string) *gitlab.Commit {
	gitToken := beego.AppConfig.String("gitToken")
	git := gitlab.NewClient(nil, gitToken)
	git.SetBaseURL("http://git.dev.cmrh.com/api/v4/")

	commit, _, err := git.Commits.GetCommit(git_id, sha)
	if err != nil {
		beego.Error(err)
		return nil
	}
	return commit
}

// 获取分支的详情，包括最新提交
func GetBranchDetail(git_id, branch string) *gitlab.Branch {
	gitToken := beego.AppConfig.String("gitToken")
	git := gitlab.NewClient(nil, gitToken)
	git.SetBaseURL("http://git.dev.cmrh.com/api/v4/")

	branch_detail, _, err := git.Branches.GetBranch(git_id, branch)
	if err != nil {
		beego.Error(err)
		return nil
	}
	return branch_detail
	
}
