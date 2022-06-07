package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"initial"
)

type ConfCaasService struct {
	Id                int       `gorm:"column:id" json:"id"`
	UnitId            int       `gorm:"column:unit_id" json:"unit_id"`
	InstanceNum       int       `gorm:"column:instance_num" json:"instance_num"`
	CpuLimit          int       `gorm:"column:cpu_limit" json:"cpu_limit"`
	MemLimit          int       `gorm:"column:mem_limit" json:"mem_limit"`
	Image             string    `gorm:"column:image" json:"image"`
	LogConfig         string    `gorm:"column:log_config" json:"log_config"`
	HealthCheck       string    `gorm:"column:health_check" json:"health_check"`
	Scheduler         string    `gorm:"column:scheduler" json:"scheduler"`
	Env               string    `gorm:"column:env" json:"env"`
	Volume            string    `gorm:"column:volume" json:"volume"`
	IsAlwaysPullImage string     `gorm:"column:is_always_pull_image" json:"is_always_pull_image"`
	SyncTime          string    `gorm:"column:sync_time" json:"sync_time"`
	IsDelete          uint8     `gorm:"column:is_delete" json:"is_delete"`
}

func (ConfCaasService) TableName() string {
	return "conf_caas_service"
}

func (ConfCaasService)UpdateOrCreate(item ConfCaasService) error  {
	var old ConfCaasService
	cond := fmt.Sprintf("is_delete = 0 and unit_id = %d", item.UnitId)
	err := initial.DB.Table(old.TableName()).First(&old,cond).Error
	if gorm.IsRecordNotFoundError(err) {
		// create
		tx := initial.DB.Begin()
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	}
	if err != nil {
		return err
	}
	// update
	tx := initial.DB.Begin()
	if err := tx.Model(&old).Updates(item).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}