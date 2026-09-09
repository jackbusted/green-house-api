package router

import (
	"encoding/json"
	"errors"
	"fmt"

	// appMiddleware "green-house-api/api/middleware"
	"green-house-api/api/route"
	"green-house-api/helper"
	"green-house-api/helper/logger"
	"green-house-api/helper/postgre"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type NewRouter struct {
	E      *echo.Echo
	Helper helper.NewHelper
}

var hostname string
var mainLog string
var dateTime time.Time

func (self *NewRouter) Register() *NewRouter {
	self.E.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization, echo.HeaderAccept,
			"App-ID",
			"App-Name",
			"App-Platform",
			"App-Version",
			"Operating-System",
			"Client-Local-IP",
			"Client-Public-IP",
			"Client-Timezone",
			"Client-Timestamp",
			"Client-Device-ID",
			"Client-Manufacture",
			"Client-Brand",
			"Client-Model",
			"Client-Operating-System",
		},
		AllowCredentials: true,
		AllowMethods:     []string{echo.OPTIONS, echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	self.E.Use(middleware.BodyLimit("200M"))
	self.E.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	self.E.Validator = self.Helper.Validator.Validator()
	self.E.Use(middleware.Recover())
	self.E.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := err.Error()
		log.Println(message)
		he, ok := err.(*echo.HTTPError)
		if ok {
			code = he.Code
			message = he.Message.(string)
		}
		c.JSON(code, map[string]interface{}{
			"code":    code,
			"status":  "error",
			"message": message,
		})
	}

	dbMaster := postgre.NewPostgre{
		Username: self.Helper.Config.GetString("database.postgre.db_master.username"),
		Password: self.Helper.Config.GetString("database.postgre.db_master.password"),
		Host:     self.Helper.Config.GetString("database.postgre.db_master.host"),
		Port:     self.Helper.Config.GetInt("database.postgre.db_master.port"),
		Name:     self.Helper.Config.GetString("database.postgre.db_master.database"),
	}

	connDbMaster, err := dbMaster.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Master database: %v", err)
	}

	dbMainMaster := postgre.NewPostgre{
		Username: self.Helper.Config.GetString("database.postgre.db_main_master.username"),
		Password: self.Helper.Config.GetString("database.postgre.db_main_master.password"),
		Host:     self.Helper.Config.GetString("database.postgre.db_main_master.host"),
		Port:     self.Helper.Config.GetInt("database.postgre.db_main_master.port"),
		Name:     self.Helper.Config.GetString("database.postgre.db_main_master.database"),
	}

	connDbMainMaster, _ := dbMainMaster.Connect()

	dbReportMaster := postgre.NewPostgre{
		Username: self.Helper.Config.GetString("database.postgre.db_report_master.username"),
		Password: self.Helper.Config.GetString("database.postgre.db_report_master.password"),
		Host:     self.Helper.Config.GetString("database.postgre.db_report_master.host"),
		Port:     self.Helper.Config.GetInt("database.postgre.db_report_master.port"),
		Name:     self.Helper.Config.GetString("database.postgre.db_report_master.database"),
	}

	connDbReportMaster, _ := dbReportMaster.Connect()

	self.E.Use(middleware.RequestID())
	self.E.Use(customRecoveryMiddleware)

	self.E.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	if true == self.Helper.Config.GetBool(`app.debug`) {
		customMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				timeStarted := time.Now()
				err := next(c)
				status := c.Response().Status
				ip := c.RealIP()
				httpErr := new(echo.HTTPError)
				if errors.As(err, &httpErr) {
					status = httpErr.Code
				}

				fields := map[string]interface{}{
					"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
					"user_agent": c.Request().UserAgent(),
					"remote":     ip,
					"method":     c.Request().Method,
					"path":       c.Request().URL.Path,
					"query":      c.Request().URL.RawQuery,
					"status":     status,
					"latency":    int64(time.Since(timeStarted) / time.Millisecond),
				}
				req, _ := json.Marshal(fields)
				log.Println(string(req))

				if err != nil {
					return err
				}

				return nil
			}
		}

		self.E.Use(customMiddleware)
		self.E.HideBanner = true
		self.E.Debug = true
	} else {
		self.E.HideBanner = true
		self.E.Debug = false
	}

	self.Helper.DB.DBMaster = connDbMaster
	self.Helper.DB.DBMainMaster = connDbMainMaster
	self.Helper.DB.DBReportMaster = connDbReportMaster

	route := route.NewRoute{
		DBMaster:       connDbMaster,
		DBMainMaster:   connDbMainMaster,
		DBReportMaster: connDbReportMaster,
		Helper:         self.Helper,
		Config:         self.Helper.Config,
		MQTTClient:     self.Helper.MQTTClient,
	}

	group := self.E.Group("api/v1")
	/* group.Use(appMiddleware.JWTWithConfig(appMiddleware.JWTConfig{
		SigningMethod: "HS512",
		SigningKey:    []byte(self.Helper.Config.GetString("jwt.secret")),
		DBMaster:      self.Helper.DB.DBMaster,
	})) */

	route.AuthRoute(group)

	settingGroup := group.Group("/setting")
	route.PersonalRoute(settingGroup)

	reportGroup := group.Group("/report")
	route.DeviceActivityRoute(reportGroup)

	return self
}

func customRecoveryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	logger.Default().Println("start of customRecoveryMiddleware")
	defer logger.Default().Println("end of customRecoveryMiddleware")
	return func(c echo.Context) error {
		logger.Default().Println("start of customRecoveryMiddleware 2")
		defer logger.Default().Println("end of customRecoveryMiddleware 2")
		defer func() {
			if r := recover(); r != nil {
				errMsg := formatPanicMessage("[ECHO HANDLER]", r)
				logPanic(errMsg)

				// Return 500 error to client
				c.Error(echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error"))
			}
		}()
		return next(c)
	}
}

func formatPanicMessage(source string, r interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	stack := string(debug.Stack())
	return fmt.Sprintf(
		"*%s* 🔥\nTimestamp: `%s`\nError: `%v`\n\nStacktrace:\n```\n%s\n```",
		source, timestamp, r, stack,
	)
}

func logPanic(str string) {
	now := time.Now()
	mainLog = hostname + `_panic` + now.Format("20060102_1504") + `.log` // + now.Format("20060102")
	logfile := `./logs/` + mainLog
	// open file read/write | create if not exist | clear file at open if exists
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	f, _ := os.OpenFile(logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	// save existing stdout | MultiWriter writes to saved stdout and file
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
	log.Println(str)
	log.SetOutput(io.Discard)
	// close file after all writes have finished
}
