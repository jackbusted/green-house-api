package model

import "time"

type PersonalModel struct {
	DefaultAttribute
	UserID             uint
	CompanyID          uint
	WorkGroupID        uint
	WorkIDNumber       string `gorm:"type:varchar(255)"`
	IDNumber           string `gorm:"type:varchar(255)"`
	Name               string `gorm:"type:varchar(255)"`
	Address            string `gorm:"type:varchar(255)"`
	Mobile             string `gorm:"type:varchar(255)"`
	Source             string `gorm:"type:varchar(255)"`
	Country            string `gorm:"type:varchar(255)"`
	Province           string `gorm:"type:varchar(255)"`
	City               string `gorm:"type:varchar(255)"`
	District           string `gorm:"type:varchar(255)"`
	SubDistrict        string `gorm:"type:varchar(255)"`
	PostalCode         string `gorm:"type:varchar(255)"`
	PasswordChangeDate time.Time
	IsShowTos          bool
	RegistrationDate   time.Time
	CodeExternal       string `gorm:"type:varchar(255)"`
	IsActive           string `gorm:"type:varchar(255)"`
	Email              string `gorm:"type:varchar(255)"`
	Remark             string `gorm:"type:varchar(255)"`
}

func (PersonalModel) TableName() string {
	return "personals"
}
