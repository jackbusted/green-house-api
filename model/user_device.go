package model

import "time"

type UserDeviceModel struct {
	DefaultAttribute
	PersonalID     uint       `json:"personal_id"`
	UserID         uint       `json:"user_id"`
	CompanyID      uint       `json:"company_id"`
	AppName        string     `json:"app_name"`
	AppVersion     string     `json:"app_version"`
	AppPlatform    string     `json:"app_platform"`
	FcmToken       string     `json:"fcm_token"`
	DeviceID       string     `json:"device_id"`
	Timezone       string     `json:"timezone"`
	UserAgent      string     `json:"user_agent"`
	Manufacture    string     `json:"manufacture"`
	Brand          string     `json:"brand"`
	Model          string     `json:"model"`
	Imei           string     `json:"imei"`
	Serial         string     `json:"serial"`
	Os             string     `json:"os"`
	Status         string     `json:"status"`
	LoginAt        *time.Time `json:"login_at"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	LastActivityAt *time.Time `json:"last_activity_at"`
	InactiveAt     *time.Time `json:"inactive_at"`
	InactiveBy     uint       `json:"inactive_by"`
	LogOutAt       *time.Time `json:"log_out_at"`
	ExpiredAt      *time.Time `json:"expired_at"`
}

func (UserDeviceModel) TableName() string {
	return "user_devices"
}
