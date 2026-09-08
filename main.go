package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"green-house-api/api/config/redis"
	"green-house-api/helper"
	"green-house-api/helper/cache"
	"green-house-api/helper/jwt"
	"green-house-api/helper/response"
	"green-house-api/helper/validator"
	"green-house-api/helper/viper"
	"green-house-api/router"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type App struct {
	config viper.Config
	helper helper.NewHelper
}

var app App
var dateTime time.Time
var hostname string
var mainLog string

func init() {
	config := viper.NewViper()

	helper := helper.NewHelper{
		Response:  response.ResponseHelper{},
		Config:    config,
		Jwt:       jwt.JwtHelper{},
		Validator: validator.NewValidator(),
	}
	app = App{
		config: config,
		helper: helper,
	}
}

func main() {
	runtime.GC()
	e := echo.New()
	e.Use(middleware.BodyLimit("200M"))
	loc, _ := time.LoadLocation("Asia/Jakarta")
	// handle err
	time.Local = loc // -> this is setting the global timezone

	defer func() {
		if r := recover(); r != nil {
			errMsg := formatPanicMessage("[MAIN]", r)

			log.Println(errMsg)
			logPanic(errMsg)
		}
	}()

	router := router.NewRouter{
		E:      e,
		Helper: app.helper,
	}
	router.Register()

	var zLog zerolog.Logger

	cache.InitCache(app.helper.Config)
	if cache.GetCacheType() == cache.RedisCache {
		err := redis.Connect(&zLog, app.helper.Config)
		if err != nil {
			zLog.Fatal().Stack().Err(err).Msg("Error redis connection")
		}

		zLog.Info().Msg("redisClient " + fmt.Sprint(redis.GetRedis()))
		zLog.Print()
	}

	hostname = router.Helper.Response.GetHostname()
	dateTime = time.Now()
	fn := logOutput()
	defer fn()

	port := flag.Int("port", app.config.GetInt(`app.port`), "port")
	host := flag.String("host", app.config.GetString(`app.host`), "host")
	flag.Parse()

	log.Println("Branch : " + router.Helper.Response.GetBranch())
	log.Println("Hash : " + router.Helper.Response.GetHash(router.Helper.Response.GetBranch()))
	log.Println("Updated : " + router.Helper.Response.GetUpdated())
	log.Println("Hostname : " + router.Helper.Response.GetHostname())
	serverAddr := *host + ":" + strconv.Itoa(*port)
	if app.config.GetBool(`app.debug`) {
		log.Println("Service RUN on DEBUG mode - HOST: " + serverAddr)
	} else {
		log.Println("Service RUN on PRODUCTION mode - HOST: " + serverAddr)
	}

	router.E.Start(serverAddr)
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
	f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer f.Close()
	// save existing stdout | MultiWriter writes to saved stdout and file
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
	log.Println(str)
	log.SetOutput(io.Discard)
	// close file after all writes have finished
}

func logOutput() func() {
	mainLog = hostname + `_logfile_now.log`
	logfile := `./logs/` + mainLog
	// open file read/write | create if not exist | clear file at open if exists
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	f, _ := os.OpenFile(logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	// save existing stdout | MultiWriter writes to saved stdout and file
	out := os.Stdout
	mw := io.MultiWriter(out, f)

	// get pipe reader and writer | writes to pipe writer come out pipe reader
	r, w, _ := os.Pipe()

	// replace stdout,stderr with pipe writer | all writes to stdout, stderr will go through pipe instead (fmt.print, log)
	os.Stdout = w
	os.Stderr = w

	// writes with log.Print should also write to mw
	log.SetOutput(mw)

	//create channel to control exit | will block until all copies are finished
	exit := make(chan bool)
	// name, _ := r.Readdirnames(0)
	// nameFile := make(chan string)
	go func() {
		// copy all reads from pipe to multiwriter, which writes to stdout and file
		_, _ = io.Copy(mw, r)
		// when r or w is closed copy will finish and true will be sent to channel
		exit <- true
	}()

	// function to be deferred in main until program exits
	return func() {
		// close writer then block on exit channel | this will let mw finish writing before the program exits
		_ = w.Close()
		<-exit
		// close file after all writes have finished
		_ = f.Close()
	}
}
