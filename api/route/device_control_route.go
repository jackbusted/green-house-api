package route

import (
	"green-house-api/api/handler"
	"green-house-api/api/usecase"
	"green-house-api/repository"

	"github.com/labstack/echo/v4"
)

func (r *NewRoute) DeviceActivityRoute(g *echo.Group) {
	deviceControlRepo := repository.NewDeviceControlRepo(r.DBMaster)

	h := handler.DeviceControlHandler{
		Helper:               r.Helper,
		DeviceControlUsecase: usecase.NewDeviceControlUsecase(r.Helper, deviceControlRepo),
		DeviceControlRepo:    deviceControlRepo,
	}

	g.POST("/device-control", h.DeviceReport)
}
