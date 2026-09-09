package request

type DeviceReportRequest struct {
	DeviceID     uint    `json:"device_id"`
	SwitchStatus string  `json:"switch_status"`
	Temperature  float64 `json:"temperature"`
	Humidity     float64 `json:"humidity"`
}
