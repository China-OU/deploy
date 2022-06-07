package models

type HarborSync struct {
	Id            int       `gorm:"column:id" json:"id"`
	ImageUrl      string    `gorm:"column:image_url" json:"image_url"`
	Result        int       `gorm:"column:result" json:"result"`
	Message       string    `gorm:"column:message" json:"message"`
	CostTime      int       `gorm:"column:cost_time" json:"cost_time"`
	ApplyPerson   string    `gorm:"column:apply_person" json:"apply_person"`
	Operator      string    `gorm:"column:operator" json:"operator"`
	InsertTime    string    `gorm:"column:insert_time" json:"insert_time"`
	IsDelete      int       `gorm:"column:is_delete" json:"is_delete"`
	SourceId      string    `gorm:"column:source_id" json:"source_id"`
}

func (HarborSync) TableName() string {
	return "harbor_sync"
}