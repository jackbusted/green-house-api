package model

import "time"

type UserModel struct {
	DefaultAttribute
	Name              string     `gorm:"type:varchar(191)" json:"name"`
	Email             string     `gorm:"type:varchar(191)" json:"email"`
	EmailVerifiedAt   *time.Time `json:"-"`
	Password          string     `gorm:"type:varchar(191)"`
	Avatar            string     `json:"-"`
	RememberToken     string     `gorm:"type:varchar(100)" json:"-"`
	Source            string     `gorm:"column:source"`
	Username          string     `gorm:"type:varchar(255)" json:"-"`
	Mobile            string     `gorm:"type:varchar(255)" json:"-"`
	Pin               string     `gorm:"type:varchar(255)" json:"-"`
	IDNumber          string     `gorm:"type:varchar(255)" json:"-"`
	Code              string     `gorm:"type:varchar(255)" json:"code"`
	EmailNotification string     `gorm:"type:varchar(255)" json:"email_notification"`
}

func (UserModel) TableName() string {
	return "users"
}
