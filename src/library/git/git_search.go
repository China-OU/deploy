package git

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/xanzy/go-gitlab"
	"models"
)

func SearchByNamespace(search string) []models.SimpleProject {
	gitToken := beego.AppConfig.String("gitToken")
	git := gitlab.NewClient(nil, gitToken)
	git.SetBaseURL("http://git.dev.cmrh.com/api/v4/")

	// 参数
	opt := &gitlab.ListProjectsOptions{}
	opt.Page = 1
	opt.PerPage = 20
	opt.Search = gitlab.String(search)
	opt.OrderBy = gitlab.String("name")
	opt.Sort = gitlab.String("asc")

	var pulic_project []*gitlab.Project
	for {
		ps, resp, err := git.Projects.ListProjects(opt)
		if err != nil {
			beego.Error(err)
			return []models.SimpleProject{}
		}
		//kkkk, _ := json.Marshal(ps)
		//beego.Info(string(kkkk))
		// 去掉个人项目
		for _, v := range ps {
			if v.Namespace.Kind != "user" {
				pulic_project = append(pulic_project, v)
			}
		}
		// 退出条件
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		if len(pulic_project) >= 10 {
			break
		}
		opt.Page = resp.NextPage
	}

	// 取所需值，无用值过滤
	var simplePrj []models.SimpleProject
	project_bytes, _ := json.Marshal(pulic_project)
	if err := json.Unmarshal([]byte(project_bytes), &simplePrj); err != nil {
		beego.Error("err:", err)
	}
	return simplePrj
}

func SearchByGitId(git_id string) gitlab.Project {
	gitToken := beego.AppConfig.String("gitToken")
	git := gitlab.NewClient(nil, gitToken)
	git.SetBaseURL("http://git.dev.cmrh.com/api/v4/")

	ps, _, err := git.Projects.GetProject(git_id, nil)
	if err != nil {
		beego.Error(err)
		return gitlab.Project{}
	}
	beego.Info(ps.HTTPURLToRepo)
	return *ps
}
