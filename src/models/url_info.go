package models

type UrlRole struct {
	Id         int       `gorm:"column:id" json:"id"`
	AppKey     string    `gorm:"column:app_key" json:"app_key"`
	SecretKey  string    `gorm:"column:secret_key" json:"secret_key"`
	Role       string    `gorm:"column:role" json:"role"`
	InsertTime string    `gorm:"column:insert_time" json:"insert_time"`
}

func (UrlRole) TableName() string {
	return "url_role"
}

type UrlAuth struct {
	Id         int       `gorm:"column:id" json:"id"`
	AppKey     string    `gorm:"column:app_key" json:"app_key"`
	SecretKey  string    `gorm:"column:secret_key" json:"secret_key"`
	Role       string    `gorm:"column:role" json:"role"`
	InsertTime string    `gorm:"column:insert_time" json:"insert_time"`
}

func (UrlRole) UrlAuth() string {
	return "url_Auth"
}