package models

type SimpleProject struct {
	ID           int    `json:"id"`
	Description  string `json:"description"`
	SSHURLToRepo string `json:"ssh_url_to_repo"`
	HTTPURLToRepo  string  `json:"http_url_to_repo"`
	Name         string `json:"name"`
	PathWithNamespace  string  `json:"path_with_namespace"`
	CreatedAt string `json:"created_at"`
}

type SimpleBranch struct {
	Name           string     `json:"name"`
	AuthorName     string     `json:"author_name"`
	CreatedAt      string     `json:"created_at"`
}

type SimpleBranchList []SimpleBranch

func (a SimpleBranchList) Len() int {    // 重写 Len() 方法
	return len(a)
}
func (a SimpleBranchList) Swap(i, j int){     // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a SimpleBranchList) Less(i, j int) bool {    // 重写 Less() 方法， 从大到小排序
	return a[j].CreatedAt < a[i].CreatedAt
}
