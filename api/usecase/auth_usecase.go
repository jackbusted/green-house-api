package usecase

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"green-house-api/api/request"
	"green-house-api/api/response"
	"green-house-api/helper"
	"green-house-api/helper/cache"
	"green-house-api/helper/logger"
	"green-house-api/model"
	"green-house-api/repository"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authUsecase struct {
	helper         helper.NewHelper
	logLoginRepo   repository.LogLoginRepoInterface
	userDeviceRepo repository.UserDeviceRepoInterface
}

type AuthUsecaseInterface interface {
	DoAuthentication(c echo.Context, req *request.Credentials) (*response.DoAuthentication, int, error)
	DoAuthenticationBySessionToken(c echo.Context, req *request.Credentials) (response.DoAuthentication, int, error)
}

func NewAuthUsecase(
	helper helper.NewHelper,
	logLoginRepo repository.LogLoginRepoInterface,
	userDeviceRepo repository.UserDeviceRepoInterface,
) AuthUsecaseInterface {
	return &authUsecase{
		helper:         helper,
		logLoginRepo:   logLoginRepo,
		userDeviceRepo: userDeviceRepo,
	}
}

func (t *authUsecase) DoAuthentication(c echo.Context, req *request.Credentials) (*response.DoAuthentication, int, error) {
	var (
		httpStatus = http.StatusOK
		resp       = response.DoAuthentication{}
		personal   = model.PersonalModel{}
		user       = model.UserModel{}
		err        = error(nil)
	)

	res := t.helper.DB.DBMaster.Where("email=? or work_id_number=? or code_external=? or mobile=?", req.Email, req.Email, req.Email, req.Email).First(&personal)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, http.StatusBadRequest, errors.New("not Found Account")
		} else {
			return nil, http.StatusInternalServerError, res.Error
		}
	}

	res = t.helper.DB.DBMaster.Where("id", personal.UserID).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, http.StatusBadRequest, errors.New("not Found Account")
		} else {
			return nil, http.StatusInternalServerError, res.Error
		}
	}

	// save personal user user has roles redis cache
	var cacheKey = fmt.Sprintf("personal:id=%v", personal.ID)
	cache.DeleteCache(cacheKey)
	err = cache.SetCache(cacheKey, &personal, 168*time.Hour)
	if err != nil {
		logger.Default().Println("Set Cache Error: " + err.Error())
		return nil, http.StatusInternalServerError, errors.New("failed to set cache: " + err.Error())
	}

	cacheKey = fmt.Sprintf("user:id=%v", user.ID)
	cache.DeleteCache(cacheKey)
	err = cache.SetCache(cacheKey, &user, 168*time.Hour)
	if err != nil {
		logger.Default().Println("Set Cache Error: " + err.Error())
		return nil, http.StatusInternalServerError, errors.New("failed to set cache: " + err.Error())
	}

	// CEK PASSWORD
	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password)); err != nil {
		return nil, http.StatusBadRequest, errors.New("email or password not match")
	}

	err = t.checkPersonalStatus(&personal)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	userDeviceData, httpStatus, err := t.checkUserDevice(c, &personal)
	if err != nil {
		return nil, httpStatus, err
	}

	token, httpStatus, err := t.generateJWTToken(c, &personal)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.AuthenticationID = req.Email
	req.AuthenticationType = "email"

	logLogin, err := t.insertLogLogin(c, &personal, userDeviceData, req)
	if err != nil {
		return nil, httpStatus, err
	}

	t.generateUserSession(&personal)

	resp.Token = token
	resp.LogLoginID = logLogin.ID

	return &resp, httpStatus, nil

}

func (t *authUsecase) DoAuthenticationBySessionToken(c echo.Context, req *request.Credentials) (response.DoAuthentication, int, error) {
	logger.Default().Println("DoAuthenticationBySessionToken")
	var (
		httpStatus  = http.StatusOK
		resp        = response.DoAuthentication{}
		personal    = model.PersonalModel{}
		user        = model.UserModel{}
		err         = error(nil)
		userSession = model.UserSessionModel{}
	)

	req.AuthenticationID = req.Code
	req.AuthenticationType = "session_token"

	res := t.helper.DB.DBMaster.Where("token=? and expiry_time >= extract(epoch from now())", req.Code).First(&userSession)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return resp, http.StatusUnauthorized, errors.New("session token not found or expired")
	}

	res = t.helper.DB.DBMaster.Where("user_id", userSession.UserID).First(&personal)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return resp, http.StatusBadRequest, errors.New("not Found Account")
		} else {
			return resp, http.StatusInternalServerError, res.Error
		}
	}

	res = t.helper.DB.DBMaster.Where("id", personal.UserID).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return resp, http.StatusBadRequest, errors.New("not Found Account")
		} else {
			return resp, http.StatusInternalServerError, res.Error
		}
	}

	// save personal user user has roles redis cache
	var cacheKey = fmt.Sprintf("personal:id=%v", personal.ID)
	cache.DeleteCache(cacheKey)
	err = cache.SetCache(cacheKey, &personal, 168*time.Hour)
	if err != nil {
		logger.Default().Println("Set Cache Error: " + err.Error())
		return resp, http.StatusInternalServerError, errors.New("failed to set cache: " + err.Error())
	}

	cacheKey = fmt.Sprintf("user:id=%v", user.ID)
	cache.DeleteCache(cacheKey)
	err = cache.SetCache(cacheKey, &user, 168*time.Hour)
	if err != nil {
		logger.Default().Println("Set Cache Error: " + err.Error())
		return resp, http.StatusInternalServerError, errors.New("failed to set cache: " + err.Error())
	}

	err = t.checkPersonalStatus(&personal)
	if err != nil {
		return resp, http.StatusBadRequest, err
	}

	userDeviceData, httpStatus, err := t.checkUserDevice(c, &personal)
	if err != nil {
		return resp, httpStatus, err
	}

	token, httpStatus, err := t.generateJWTToken(c, &personal)
	if err != nil {
		return resp, http.StatusInternalServerError, err
	}

	logLogin, err := t.insertLogLogin(c, &personal, userDeviceData, req)
	if err != nil {
		return resp, httpStatus, err
	}

	resp.Token = token
	resp.LogLoginID = logLogin.ID

	return resp, httpStatus, nil
}

func (t *authUsecase) checkPersonalStatus(personal *model.PersonalModel) error {
	logger.Default().Println("checkPersonalStatus")
	if personal.IsActive != "1" {
		return errors.New("your account is inactive, please contact your administrator")
	}
	return nil
}

func (t *authUsecase) checkUserDevice(c echo.Context, personal *model.PersonalModel) (*model.UserDeviceModel, int, error) {
	logger.Default().Println("start of checkUserDevice")
	defer func() {
		logger.Default().Println("end of checkUserDevice")
	}()

	appName := c.Request().Header.Get("App-Name")
	appVersion := c.Request().Header.Get("App-Version")
	appPlatform := c.Request().Header.Get("App-Platform")
	deviceID := c.Request().Header.Get("Client-Device-ID")
	manufacture := c.Request().Header.Get("Client-Manufacture")
	brand := c.Request().Header.Get("Client-Brand")
	clientModel := c.Request().Header.Get("Client-Model")
	os := c.Request().Header.Get("Client-Operating-System")
	timezone := c.Request().Header.Get("Client-Timezone")
	userAgent := c.Request().Header.Get("User-Agent")

	now := t.getTimeNow()

	var err error
	var userDeviceData model.UserDeviceModel
	res := t.helper.DB.DBMaster.Where(map[string]interface{}{
		"user_id":      personal.UserID,
		"personal_id":  personal.ID,
		"app_name":     appName,
		"app_platform": appPlatform,
		"device_id":    deviceID,
	}).Where("status in ('Inactive','Active','Log Out')").Last(&userDeviceData)
	if res.Error != nil {
		return &userDeviceData, http.StatusInternalServerError, errors.New("create log user device (1): " + res.Error.Error())
	}

	logger.Default().Println("userDeviceData.ID : ", userDeviceData.ID)
	logger.Default().Println("deviceID : ", deviceID)

	if userDeviceData.ID == 0 {
		userDeviceData.PersonalID = personal.ID
		userDeviceData.UserID = personal.UserID
		userDeviceData.AppName = appName
		userDeviceData.AppVersion = appVersion
		userDeviceData.AppPlatform = appPlatform
		userDeviceData.DeviceID = deviceID
		userDeviceData.Timezone = timezone
		userDeviceData.UserAgent = userAgent
		userDeviceData.Manufacture = manufacture
		userDeviceData.Brand = brand
		userDeviceData.Model = clientModel
		userDeviceData.Os = os
		userDeviceData.LoginAt = &now
		userDeviceData.LastLoginAt = &now
		userDeviceData.LastActivityAt = &now
		userDeviceData.Status = "Active"
		res = t.helper.DB.DBMaster.Create(&userDeviceData)
		if res.Error != nil {
			return &userDeviceData, http.StatusInternalServerError, errors.New("create log user device (2): " + res.Error.Error())
		}
	} else {
		logger.Default().Println("userDeviceData.ID : ", userDeviceData.ID)
		logger.Default().Println("deviceID : ", deviceID)

		if userDeviceData.DeviceID == deviceID {
			logger.Default().Println("userDeviceData.DeviceID == deviceID")
			userDeviceData.LastLoginAt = &now
			userDeviceData.LastActivityAt = &now
			userDeviceData.UserAgent = userAgent
			userDeviceData.AppVersion = appVersion
			userDeviceData.Status = "Active"

			err := t.userDeviceRepo.Updates(userDeviceData, "id", userDeviceData.ID)
			if err != nil {
				return &userDeviceData, http.StatusInternalServerError, errors.New("update log user device (3): " + err.Error())
			}

		} else if userDeviceData.DeviceID != deviceID && appPlatform == "web" {
			logger.Default().Println(`serDeviceData.DeviceID != deviceID && appPlatform == "web"`)
			var userDeviceData model.UserDeviceModel
			res = t.helper.DB.DBMaster.Where(map[string]interface{}{
				"user_id":      personal.UserID,
				"personal_id":  personal.ID,
				"app_name":     appName,
				"app_platform": appPlatform,
				"device_id":    deviceID,
			}).Where("(status ='Active' or status='Log Out')").First(&userDeviceData)
			if res.Error != nil {
				return &userDeviceData, http.StatusInternalServerError, errors.New("create log user device (4): " + res.Error.Error())
			}

			if userDeviceData.ID == 0 {
				userDeviceData.PersonalID = personal.ID
				userDeviceData.UserID = personal.UserID
				userDeviceData.AppName = appName
				userDeviceData.AppVersion = appVersion
				userDeviceData.AppPlatform = appPlatform
				userDeviceData.DeviceID = deviceID
				userDeviceData.Timezone = timezone
				userDeviceData.UserAgent = userAgent
				userDeviceData.Manufacture = manufacture
				userDeviceData.Brand = brand
				userDeviceData.Model = clientModel
				userDeviceData.Os = os
				userDeviceData.LoginAt = &now
				userDeviceData.LastLoginAt = &now
				userDeviceData.LastActivityAt = &now
				userDeviceData.Status = "Active"
				userDeviceData, err = t.userDeviceRepo.CreateAndFirst(userDeviceData)
				if err != nil {
					return &userDeviceData, http.StatusInternalServerError, errors.New("create log user device (5): " + err.Error())
				}
			} else {
				userDeviceData.LastLoginAt = &now
				userDeviceData.LastActivityAt = &now
				userDeviceData.UserAgent = userAgent
				userDeviceData.AppVersion = appVersion
				userDeviceData.Status = "Active"
				err := t.userDeviceRepo.Updates(userDeviceData, "id", userDeviceData.ID)
				if err != nil {
					log.Printf(err.Error())
					return &userDeviceData, http.StatusInternalServerError, errors.New("update log user device (6): " + err.Error())
				}
			}
		} else {
			return &userDeviceData, http.StatusUnauthorized, errors.New("Device not recognize, contact your admin to inactive your old device.")
		}
	}

	return &userDeviceData, http.StatusOK, nil
}

func (t *authUsecase) generateJWTToken(c echo.Context, personal *model.PersonalModel) (string, int, error) {
	jti, _ := json.Marshal(map[string]interface{}{
		"personal_id": personal.ID,
		"user_id":     personal.UserID,
	})

	token, err := t.helper.Jwt.CreateJwtToken(t.helper.Config.GetString("jwt.secret"), string(jti))
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("failed to generate token")
	}

	return token, http.StatusOK, nil
}

func (t *authUsecase) insertLogLogin(c echo.Context, personal *model.PersonalModel, userDeviceData *model.UserDeviceModel, req *request.Credentials) (model.LogLoginModel, error) {
	now := t.getTimeNow()
	dataLogLogin := model.LogLoginModel{}
	dataLogLogin.IPSource = ""
	dataLogLogin.UserAgent = c.Request().Header.Get("User-Agent")
	dataLogLogin.Date = now
	dataLogLogin.Time = now
	dataLogLogin.LogStatus = 0
	dataLogLogin.UserID = personal.UserID
	dataLogLogin.Name = personal.Name
	dataLogLogin.Email = personal.Email
	dataLogLogin.Sources = ""
	dataLogLogin.PersonalID = personal.ID
	dataLogLogin.UserID = userDeviceData.ID
	dataLogLogin.Data = map[string]interface{}{
		"authentication_type": req.AuthenticationType,
		"authentication_id":   req.AuthenticationID,
		"app_id":              c.Request().Header.Get("App-ID"),
		"app_name":            c.Request().Header.Get("App-Name"),
		"app_platform":        c.Request().Header.Get("App-Platform"),
		"app_version":         c.Request().Header.Get("App-Version"),
		"timezone":            c.Request().Header.Get("Client-Timezone"),
		"timestamp":           c.Request().Header.Get("Client-Timestamp"),
		"operating_system":    c.Request().Header.Get("Client-Operating-System"),
		"manufacture":         c.Request().Header.Get("Client-Manufacture"),
		"brand":               c.Request().Header.Get("Client-Brand"),
		"model":               c.Request().Header.Get("Client-Model"),
		"device_id":           c.Request().Header.Get("Client-Device-ID"),
		"local_ip":            c.Request().Header.Get("Client-Local-IP"),
		"public_ip":           c.Request().Header.Get("Client-Public-IP"),
		"error":               "",
		"source_app_id":       req.SourceAppID,
	}

	return t.logLoginRepo.Create(dataLogLogin)
}

func (t *authUsecase) generateUserSession(personal *model.PersonalModel) (int, error) {
	res := t.helper.DB.DBMaster.Create(&model.UserSessionModel{
		UserID:     personal.UserID,
		Token:      t.generateAccessToken(personal.Name),
		ExpiryTime: time.Now().Add(8760 * time.Hour).Unix(),
	})
	if res.Error != nil {
		return http.StatusInternalServerError, errors.New("create user session : " + res.Error.Error())
	}

	return http.StatusOK, nil
}

func (t *authUsecase) generateAccessToken(str string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(hash)
}

func (t *authUsecase) getTimeNow() time.Time {
	layout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Asia/Jakarta")
	nowWIT := time.Now().In(loc)
	now, _ := time.Parse(layout, nowWIT.Format(layout))
	return now
}
