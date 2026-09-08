package repository

import (
	"errors"
	"green-house-api/model"

	"gorm.io/gorm"
)

type userRepo struct {
	DB        *gorm.DB
	TableName string
}

type UserRepoInterface interface {
	Authentication(authenticationID string) (model.UserModel, error)
	FindOneByID(ID uint) (model.UserModel, error)
}

func NewUserRepo(db *gorm.DB) UserRepoInterface {
	var model model.UserModel
	return &userRepo{
		DB:        db,
		TableName: model.TableName(),
	}
}

func (t *userRepo) Authentication(authenticationID string) (model.UserModel, error) {
	data := model.UserModel{}
	err := t.DB.Where("( email=? )", authenticationID).First(&data)
	if err.Error != nil {
		return data, err.Error
	}
	return data, nil
}

func (t *userRepo) FindOneByID(ID uint) (model.UserModel, error) {
	data := model.UserModel{}
	err := t.DB.Table(t.TableName).Preload("UserHasRoles").First(&data, ID)
	if err.Error != nil {
		return data, t.getError(err.Error)
	}
	return data, nil
}

func (t *userRepo) getError(err error) error {
	return errors.New(t.TableName + ": " + err.Error())
}
