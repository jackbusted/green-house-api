package handler

import (
	"fmt"
	"green-house-api/api/request"
	"green-house-api/api/response"
	"green-house-api/api/usecase"
	"green-house-api/helper"
	"green-house-api/repository"

	"github.com/labstack/echo/v4"
)

type DeviceControlHandler struct {
	Helper               helper.NewHelper
	DeviceControlUsecase usecase.DeviceControlUsecaseInterface
	DeviceControlRepo    repository.DeviceControlRepoInterface
}

func (t *DeviceControlHandler) DeviceReport(c echo.Context) error {
	var (
		err     error
		request request.DeviceReportRequest
		result  response.DeviceReportResp
	)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic occured :", r)
		}
	}()

	err = c.Bind(&request)
	if err != nil {
		return t.Helper.Response.SendError(c, err.Error(), err.Error())
	}

	err = t.Helper.Validator.Validate(request)
	if err != nil {
		return t.Helper.Response.SendError(c, err.Error(), err.Error())
	}

	result, err = t.DeviceControlUsecase.ReportDeviceStatus(c, request)
	if err != nil {
		return t.Helper.Response.SendBadRequest(c, err.Error(), err.Error())
	}

	return t.Helper.Response.SendSuccess(c, "", result)
}
