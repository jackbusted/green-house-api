package route

import (
	"green-house-api/api/handler"
	"green-house-api/api/usecase"
	"green-house-api/repository"

	"github.com/labstack/echo/v4"
)

func (r *NewRoute) AuthRoute(g *echo.Group) {
	userRepo := repository.NewUserRepo(r.DBMaster)
	logLoginRepo := repository.NewLogLoginRepo(r.DBMaster)
	userDeviceRepo := repository.NewUserDeviceRepo(r.DBMaster)

	auth := handler.Auth{
		Helper:       r.Helper,
		UserRepo:     userRepo,
		PersonalRepo: repository.NewPersonalRepo(r.DBMaster),
		LogLoginRepo: logLoginRepo,

		AuthUsecase:    usecase.NewAuthUsecase(r.Helper, logLoginRepo, userDeviceRepo),
		UserDeviceRepo: repository.NewUserDeviceRepo(r.DBMaster),
	}

	g.POST("/login", auth.Login)
	g.POST("/authentication", auth.PostAuthentication)
	g.POST("/authorization", auth.PostAuthorization)
}
