package model

import "time"

type LogLoginModel struct {
	DefaultAttribute
	IPSource     string
	UserAgent    string
	Date         time.Time
	Time         time.Time
	LogStatus    int
	UserID       uint
	CompanyID    uint
	Name         string
	Email        string
	Sources      string
	LogoutDate   time.Time
	LogoutTime   time.Time
	PersonalID   uint
	Data         MapStringInterface
	UserDeviceID uint
	Description  string
}

func (LogLoginModel) TableName() string {
	return "log_logins"
}
