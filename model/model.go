package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type DefaultAttribute struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedBy *uint          `gorm:"column:created_by" json:"-" `
	UpdatedBy *uint          `gorm:"column:updated_by" json:"-" `
	DeletedBy *uint          `gorm:"column:deleted_by" json:"-" `
	CreatedAt *time.Time     `gorm:"column:created_at" json:"CreatedAt" `
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"UpdatedAt" `
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type MapStringInterface map[string]interface{}

func (msi MapStringInterface) Value() (driver.Value, error) {
	valueString, err := json.Marshal(msi)
	return string(valueString), err
}

func (msi *MapStringInterface) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &msi); err != nil {
		return err
	}
	return nil
}
