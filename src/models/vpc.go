package models

type VpcExt struct {
	Id         int       `gorm:"column:id" json:"id"`
	Vpc        string    `gorm:"column:vpc" json:"vpc"`
	VpcName    string    `gorm:"column:vpc_name" json:"vpc_name"`
	Operator   string    `gorm:"column:operator" json:"operator"`
	InsertTime string    `gorm:"column:insert_time" json:"insert_time"`
	IsDelete   string    `gorm:"column:is_delete" json:"is_delete"`
}

func (VpcExt) TableName() string {
	return "vpc_ext"
}