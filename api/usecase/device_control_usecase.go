package usecase

import (
	"green-house-api/api/request"
	"green-house-api/api/response"
	"green-house-api/helper"
	"green-house-api/model"
	"green-house-api/repository"
	"log"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type deviceControlUsecase struct {
	helper            helper.NewHelper
	deviceControlRepo repository.DeviceControlRepoInterface
}

type DeviceControlUsecaseInterface interface {
	ReportDeviceStatus(c echo.Context, req request.DeviceReportRequest) (response.DeviceReportResp, error)
}

func NewDeviceControlUsecase(helper helper.NewHelper, deviceControlRepo repository.DeviceControlRepoInterface) DeviceControlUsecaseInterface {
	return &deviceControlUsecase{
		helper:            helper,
		deviceControlRepo: deviceControlRepo,
	}
}

func (t *deviceControlUsecase) ReportDeviceStatus(c echo.Context, req request.DeviceReportRequest) (response.DeviceReportResp, error) {
	current := t.getTimeNow()

	var (
		err        error
		data       model.DeviceIdentity
		resp       response.DeviceReportResp
		dbStatus   string = "connected"
		mqttStatus string = "connected"
	)

	data, err = t.deviceControlRepo.FindOneByID(req.DeviceID)
	if err != nil {
		return resp, err
	}

	switchStatus := false
	if strings.ToLower(req.SwitchStatus) == "on" {
		switchStatus = true
	}

	if !t.helper.MQTTClient.IsConnected() {
		mqttStatus = "disconnected"
	}

	dataReport := model.DeviceReport{}
	dataReport.DeviceID = data.ID
	dataReport.SwitchStatus = switchStatus
	dataReport.DeviceStatus = "healthy" // device condition. eg : healthy
	dataReport.DatabaseStatus = dbStatus
	dataReport.MqttStatus = mqttStatus
	dataReport.Temperature = req.Temperature
	dataReport.Humidity = req.Humidity
	dataReport.CreatedAt = &current
	dataReport.UpdatedAt = &current
	err = t.deviceControlRepo.CreateData(dataReport)
	if err != nil {
		return resp, err
	}

	resp.DeviceID = data.ID
	resp.SwitchAction = "on"
	resp.DatabaseStatus = dbStatus
	resp.MqttStatus = mqttStatus

	log.Println("start publish MQTT")
	err = t.helper.MQTTClient.Publish("greenhouse/test", resp)
	if err != nil {
		return resp, err
	}
	log.Println("success publish MQTT")

	return resp, nil
}

func (t *deviceControlUsecase) getTimeNow() time.Time {
	layout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Asia/Jakarta")
	nowWIT := time.Now().In(loc)
	now, _ := time.Parse(layout, nowWIT.Format(layout))
	return now
}
