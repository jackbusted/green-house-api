package logger

import (
	"green-house-api/helper/viper"
	"io"
	"log"
	"os"
	"time"

	gorm "gorm.io/gorm/logger"
)

var logger log.Logger

var basePrefix string

const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix                    // move the "prefix" from the beginning of the line to before the message
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

type loggerJson struct {
	Prefix  string `json:"prefix"`
	Message string `json:"message"`
}

func Default() *log.Logger { return &logger }

func init() {
	create()
}

func create() {
	now := time.Now()
	file, err := openLogFile("./logs/"+now.Format("2006/01/02"), "main.log")
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, file)
	config := viper.NewViper()
	basePrefix = config.GetString("app.name")
	logger.SetOutput(mw)
	SetPrefix("")
	SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
}

func Reset() {
	create()
}

func openLogFile(path string, name string) (*os.File, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return nil, err
		}
	}
	logFile, err := os.OpenFile(path+"/"+name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func SetPrefix(prefix string) {
	if prefix != "" {
		logger.SetPrefix("[" + basePrefix + "." + prefix + "] ")
	} else {
		logger.SetPrefix("[" + basePrefix + "] ")
	}
}

func SetOutput(name string) {
	now := time.Now()
	file, err := openLogFile("./logs/"+now.Format("2006/01/02"), name+".log")
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, file)
	logger.SetOutput(mw)
	logger.SetOutput(io.Discard)
}

func SetFlags(flag int) {
	logger.SetFlags(flag)
}

func GormLog() gorm.Interface {

	now := time.Now()
	fileError, err := openLogFile("./logs/"+now.Format("2006/01/02"), "main.log")
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, fileError)
	newLogger := gorm.New(
		log.New(mw, "[gorm]\r\n", log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
		gorm.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gorm.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	return newLogger
}
