package models

import "time"

type OnlineStdVM struct {
    ID              int         `gorm:"column:id" json:"id"`
    OnlineID        int         `gorm:"column:online_id" json:"online_id"`
    JenkinsName     string      `gorm:"column:jenkins_name" json:"jenkins_name"`
    BuildStatus     int         `gorm:"column:build_status" json:"build_status"`
    UpgradeStatus   int         `gorm:"column:upgrade_status" json:"upgrade_status"`
    ArtifactURL     string      `gorm:"column:artifact_url" json:"artifact_url"`
    BuildDuration   int         `gorm:"column:build_duration" json:"build_duration"`
    UpgradeDuration int         `gorm:"column:upgrade_duration" json:"upgrade_duration"`
    UpgradeLogs     string      `gorm:"column:upgrade_logs" json:"upgrade_logs"`
    CreateTime      string      `gorm:"column:create_time" json:"create_time"`
    UpdateTime      time.Time   `gorm:"column:update_time;default:current_time on update current_time" json:"update_time"`
    IsDelete        int         `gorm:"column:is_delete" json:"is_delete"`
}

func (*OnlineStdVM) TableName() string {
    return "online_std_vm"
}

type OnlineNvm struct {
    ID              int         `gorm:"column:id" json:"id"`
    OnlineId        int         `gorm:"column:online_id" json:"online_id"`
    FileAddr        string      `gorm:"column:file_addr" json:"file_addr"`
    FileName        string      `gorm:"column:file_name" json:"file_name"`
    Sha256          string      `gorm:"column:sha_256" json:"sha_256"`
    Host            string      `gorm:"column:host" json:"host"`
    ShellLog        string      `gorm:"column:shell_log" json:"shell_log"`
    InsertTime      string      `gorm:"column:insert_time" json:"insert_time"`
    IsDelete        int         `gorm:"column:is_delete" json:"is_delete"`
}

func (OnlineNvm) TableName() string {
    return "online_nvm"
}