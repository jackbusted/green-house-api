package repository

import (
	"green-house-api/model"

	"gorm.io/gorm"
)

type deviceControlRepo struct {
	DB        *gorm.DB
	TableName string
}

type DeviceControlRepoInterface interface {
	FindOneByID(ID uint) (model.DeviceIdentity, error)
	CreateData(data model.DeviceReport) error
}

func NewDeviceControlRepo(db *gorm.DB) DeviceControlRepoInterface {
	var model model.DeviceReport
	return &deviceControlRepo{
		DB:        db,
		TableName: model.TableName(),
	}
}

func (t *deviceControlRepo) FindOneByID(ID uint) (model.DeviceIdentity, error) {
	var err error
	data := model.DeviceIdentity{}
	err = t.DB.Where("id = ?", ID).First(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (t *deviceControlRepo) CreateData(data model.DeviceReport) error {
	var err error

	err = t.DB.Create(&data).Error
	if err != nil {
		return err
	}

	return nil
}
