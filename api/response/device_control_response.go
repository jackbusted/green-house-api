package response

type DeviceReportResp struct {
	DeviceID       uint   `json:"device_id"`
	SwitchAction   string `json:"switch_action"`
	DatabaseStatus string `json:"database_status"`
	MqttStatus     string `json:"mqtt_status"`
}
