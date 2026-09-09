package model

type DeviceReport struct {
	DefaultAttribute
	DeviceID       uint
	SwitchStatus   bool
	DeviceStatus   string
	DatabaseStatus string
	MqttStatus     string
	Temperature    float64
	Humidity       float64
}

func (DeviceReport) TableName() string {
	return "device_reports"
}
