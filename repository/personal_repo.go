package repository

import (
	"green-house-api/model"

	"gorm.io/gorm"
)

type personalRepo struct {
	DB        *gorm.DB
	TableName string
}

type PersonalRepoInterface interface {
	FindOne(data model.PersonalModel) (model.PersonalModel, error)
}

func NewPersonalRepo(db *gorm.DB) PersonalRepoInterface {
	var model model.PersonalModel
	return &personalRepo{
		DB:        db,
		TableName: model.TableName(),
	}
}

func (t *personalRepo) FindOne(data model.PersonalModel) (model.PersonalModel, error) {
	err := t.DB.Where(&data).First(&data)
	if err.Error != nil {
		return data, err.Error
	}
	return data, nil
}
