package model

import "time"

type DeviceIdentity struct {
	DefaultAttribute
	Code     string
	Name     string
	Batch    string
	IsActive bool
	InTime   time.Time
}

func (DeviceIdentity) TableName() string {
	return "device_identities"
}
