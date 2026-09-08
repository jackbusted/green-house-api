package repository

import (
	"green-house-api/model"

	"gorm.io/gorm"
)

type userDeviceRepo struct {
	DB        *gorm.DB
	TableName string
}

type UserDeviceRepoInterface interface {
	Finds(where interface{}, args ...interface{}) ([]model.UserDeviceModel, error)
	CreateAndFirst(data model.UserDeviceModel) (model.UserDeviceModel, error)
	Updates(values interface{}, query interface{}, args ...interface{}) error
}

func NewUserDeviceRepo(db *gorm.DB) UserDeviceRepoInterface {
	var model model.UserDeviceModel
	return &userDeviceRepo{
		DB:        db,
		TableName: model.TableName(),
	}
}

func (t *userDeviceRepo) Finds(where interface{}, args ...interface{}) ([]model.UserDeviceModel, error) {
	var data []model.UserDeviceModel
	res := t.DB.Where(where, args).Find(&data)
	if res.Error != nil {
		return data, res.Error
	}
	return data, nil
}

func (t *userDeviceRepo) CreateAndFirst(data model.UserDeviceModel) (model.UserDeviceModel, error) {
	res := t.DB.Create(&data)
	if res.Error != nil {
		return data, res.Error
	}
	return data, nil
}

func (t *userDeviceRepo) Updates(values interface{}, query interface{}, args ...interface{}) error {
	data := model.UserDeviceModel{}
	res := t.DB.Model(&data).Where(query, args).Updates(values)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
