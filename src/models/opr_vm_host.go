package models

import "time"

type OprVMHost struct {
    ID          int       `gorm:"column:id" json:"id"`
    UnitID      int       `gorm:"column:unit_id" json:"unit_id"`
    Unit        string    `gorm:"column:unit" json:"unit"`
    UnitCN      string    `gorm:"column:unit_cn" json:"unit_cn"`
    Hosts       string    `gorm:"column:hosts" json:"hosts"`
    Type        string    `gorm:"column:type" json:"type"`
    Command     string    `gorm:"column:command" json:"command"`
    Args        string    `gorm:"column:args" json:"args"`
    ExecPath    string    `gorm:"column:exec_path" json:"exec_path"`
    ExecUser    string    `gorm:"column:exec_user" json:"exec_user"`
    Operator    string    `gorm:"column:operator" json:"operator"`
    Status      int       `gorm:"column:status" json:"status"`
    Logs        string    `gorm:"column:logs" json:"logs"`
    CreateTime  string    `gorm:"column:create_time" json:"create_time"`
    UpdateTime  time.Time `gorm:"column:update_time;default:current_time on update current_time" json:"update_time"`
    Deleted     int       `gorm:"column:deleted" json:"deleted"`
}

func (OprVMHost) TableName() string {
    return "opr_vm_host"
}
