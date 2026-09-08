package handler

import (
	"fmt"
	"green-house-api/api/request"
	"green-house-api/api/response"
	"green-house-api/api/usecase"
	"green-house-api/helper"
	"green-house-api/helper/logger"
	"green-house-api/model"
	"green-house-api/repository"

	"github.com/labstack/echo/v4"
)

type Auth struct {
	Helper         helper.NewHelper
	UserRepo       repository.UserRepoInterface
	PersonalRepo   repository.PersonalRepoInterface
	LogLoginRepo   repository.LogLoginRepoInterface
	AuthUsecase    usecase.AuthUsecaseInterface
	UserDeviceRepo repository.UserDeviceRepoInterface
}

func (a *Auth) Login(c echo.Context) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic occured :", r)
		}
	}()

	var (
		err                    error
		httpStatus             int
		credential             request.Credentials
		authenticationResponse *response.DoAuthentication
	)

	if err = c.Bind(&credential); err != nil {
		return a.Helper.Response.SendBadRequest(c, err.Error(), nil)
	}

	err = a.Helper.Validator.Validate(credential)
	if err != nil {
		return a.Helper.Response.SendBadRequest(c, err.Error(), nil)
	}

	authenticationResponse, httpStatus, err = a.AuthUsecase.DoAuthentication(c, &credential)
	if err != nil {
		return a.Helper.Response.SendCustomResponse(c, httpStatus, err.Error(), nil)
	}

	return a.Helper.Response.SendSuccess(c, "", authenticationResponse)
}

func (a *Auth) PostAuthentication(c echo.Context) error {
	var req struct {
		AuthenticationID string `validate:"required"`
	}
	var err error
	if err = c.Bind(&req); err != nil {
		return a.Helper.Response.SendBadRequest(c, err.Error(), nil)
	}

	if err = a.Helper.Validator.Validate(req); err != nil {
		return a.Helper.Response.SendBadRequest(c, err.Error(), nil)
	}

	// GET USERS
	user, err := a.UserRepo.Authentication(req.AuthenticationID)
	if err != nil {
		return a.Helper.Response.SendBadRequest(c, "Not Found Account", nil)
	}

	personal, err := a.PersonalRepo.FindOne(model.PersonalModel{UserID: user.ID})
	if err != nil {
		return a.Helper.Response.SendBadRequest(c, "Not Found Account", nil)
	}

	return a.Helper.Response.SendSuccess(c, "", map[string]interface{}{
		"user":     user,
		"personal": personal,
	})
}

func (a *Auth) PostAuthorization(c echo.Context) error {
	logger.Default().Println("start of PostAuthorization")
	defer logger.Default().Println("end of PostAuthorization")

	var req struct {
		UserID uint `validate:"required"`
	}
	var err error
	if err = c.Bind(&req); err != nil {
		return a.Helper.Response.SendBadRequest(c, err.Error(), nil)
	}

	if err = a.Helper.Validator.Validate(req); err != nil {
		return a.Helper.Response.SendBadRequest(c, err.Error(), nil)
	}
	// GET USERS
	user, err := a.UserRepo.FindOneByID(req.UserID)
	if err != nil {
		return a.Helper.Response.SendUnauthorized(c, "Not Found User", nil)
	}

	personal, err := a.PersonalRepo.FindOne(model.PersonalModel{UserID: user.ID})
	if err != nil {
		return a.Helper.Response.SendUnauthorized(c, "Not Found Personal", nil)
	}

	return a.Helper.Response.SendSuccess(c, "", map[string]interface{}{
		"user":     user,
		"personal": personal,
	})
}
