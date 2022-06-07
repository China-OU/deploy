package models

// 全部上线单元列表，总表
type OnlineAllList struct {
	Id             int       `gorm:"column:id" json:"id"`
	UnitId         int       `gorm:"column:unit_id" json:"unit_id"`
	Branch         string    `gorm:"column:branch" json:"branch"`
	CommitId       string    `gorm:"column:commit_id" json:"commit_id"`
	ShortCommitId  string    `gorm:"column:short_commit_id" json:"short_commit_id"`
	OnlineDate     string    `gorm:"column:online_date" json:"online_date"`
	OnlineTime     string    `gorm:"column:online_time" json:"online_time"`
	Version        string    `gorm:"column:version" json:"version"`
	IsProcessing   int       `gorm:"column:is_processing" json:"is_processing"`
	IsSuccess      int       `gorm:"column:is_success" json:"is_success"`
	IsDelete       int       `gorm:"column:is_delete" json:"is_delete"`
	Operator       string    `gorm:"column:operator" json:"operator"`
	ExcuteTime     string    `gorm:"column:excute_time" json:"excute_time"`
	InsertTime     string    `gorm:"column:insert_time" json:"insert_time"`
	ErrorLog       string    `gorm:"column:error_log" json:"error_log"`
	SourceId       string    `gorm:"column:source_id" json:"source_id"`
}

func (OnlineAllList) TableName() string {
	return "online_all_list"
}

// 标准容器上线单元列表，总表
type OnlineStdCntr struct {
	Id               int       `gorm:"column:id" json:"id"`
	OnlineId         int       `gorm:"column:online_id" json:"online_id"`
	OprCntrId        int       `gorm:"column:opr_cntr_id" json:"opr_cntr_id"`
	JenkinsName      string    `gorm:"column:jenkins_name" json:"jenkins_name"`
	JenkinsSuccess   int       `gorm:"column:jenkins_success" json:"jenkins_success"`
	JenkinsImage     string    `gorm:"column:jenkins_image" json:"jenkins_image"`
	JenkinsCostTime  int       `gorm:"column:jenkins_cost_time" json:"jenkins_cost_time"`
	IsDelete         int       `gorm:"column:is_delete" json:"is_delete"`
	InsertTime       string    `gorm:"column:insert_time" json:"insert_time"`
}

func (OnlineStdCntr) TableName() string {
	return "online_std_cntr"
}

// db上线单元列表，总表
type OnlineDbList struct {
	Id               int       `gorm:"column:id" json:"id"`
	OnlineId         int       `gorm:"column:online_id" json:"online_id"`
	DirName          string    `gorm:"column:dir_name" json:"dir_name"`
	IsDirClear       int       `gorm:"column:is_dir_clear" json:"is_dir_clear"`
	IsPullDir        int       `gorm:"column:is_pull_dir" json:"is_pull_dir"`
	IsDelete         int       `gorm:"column:is_delete" json:"is_delete"`
}

func (OnlineDbList) TableName() string {
	return "online_db_list"
}


// db文件详情列表
type OnlineDbLog struct {
	Id               int       `gorm:"column:id" json:"id"`
	OnlineId         int       `gorm:"column:online_id" json:"online_id"`
	FileName         string    `gorm:"column:file_name" json:"file_name"`
	FileSha          string    `gorm:"column:file_sha" json:"file_sha"`
	FilePath         string    `gorm:"column:file_path" json:"file_path"`
	FileContent      string    `gorm:"column:file_content" json:"file_content"`
	IsSuccess        int       `gorm:"column:is_success" json:"is_success"`
	Message          string    `gorm:"column:message" json:"message"`
	SqlType          string    `gorm:"column:sql_type" json:"sql_type"`
	ExecuteTime      int       `gorm:"column:execute_time" json:"execute_time"`
	ProxyUser        string    `gorm:"column:proxy_user" json:"proxy_user"`
	Command          string    `gorm:"column:command" json:"command"`
	StartTime        string    `gorm:"column:start_time" json:"start_time"`
	InsertTime       string    `gorm:"column:insert_time" json:"insert_time"`
	IsDelete         int       `gorm:"column:is_delete" json:"is_delete"`
	Operator         string    `gorm:"column:operator" json:"operator"`
}

func (OnlineDbLog) TableName() string {
	return "online_db_log"
}