package models

import "time"

type DBAccount struct {
    ID              int         `gorm:"column:id" json:"id"`
    Dialect         string      `gorm:"column:dialect" json:"dialect"`
    Corp            string      `gorm:"column:corp" json:"corp"`
    Host            string      `gorm:"column:host" json:"host"`
    Port            uint        `gorm:"column:port" json:"port"`
    Schema          string      `gorm:"column:schema" json:"schema"`
    Username        string      `gorm:"column:username" json:"username"`
    Key             string      `gorm:"column:key" json:"key"`
    EncryptedPWD    string      `gorm:"column:encrypted_pwd" json:"encrypted_pwd"`
    CreateTime      string      `gorm:"column:create_time" json:"create_time"`
    UpdateTime      time.Time   `gorm:"column:update_time;default:current_time on update current_time" json:"update_time"`
    Expired         int         `gorm:"column:expired" json:"expired"`
    ExpireTime      time.Time   `gorm:"column:expire_time" json:"expire_time"`
}

func (DBAccount) TableName() string {
    return "conf_db_account"
}
