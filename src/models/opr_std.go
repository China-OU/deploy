package models

type OprCntrUpgrade struct {
	Id            int       `gorm:"column:id" json:"id"`
	UnitId        int       `gorm:"column:unit_id" json:"unit_id"`
	OldImage      string    `gorm:"column:old_image" json:"old_image"`
	NewImage      string    `gorm:"column:new_image" json:"new_image"`
	Result        int       `gorm:"column:result" json:"result"`
	Operator      string    `gorm:"column:operator" json:"operator"`
	Message       string    `gorm:"column:message" json:"message"`
	OnlineDate    string    `gorm:"column:online_date" json:"online_date"`
	CostTime      int       `gorm:"column:cost_time" json:"cost_time"`
	InsertTime    string    `gorm:"column:insert_time" json:"insert_time"`
	SourceId      string    `gorm:"column:source_id" json:"source_id"`
}

func (OprCntrUpgrade) TableName() string {
	return "opr_cntr_upgrade"
}

