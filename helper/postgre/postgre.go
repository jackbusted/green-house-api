package postgre

import (
	"green-house-api/helper/viper"
	"log"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DBMaster       *gorm.DB
	DBMainMaster   *gorm.DB
	DBReportMaster *gorm.DB
}

type NewPostgre struct {
	Username string
	Password string
	Host     string
	Port     int
	Name     string
}

func (self *NewPostgre) Connect() (*gorm.DB, error) {
	config := viper.NewViper()
	appName := config.GetString("app.name")

	timezone := "Asia/Jakarta"

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=" + self.Host + " user=" + self.Username + " password=" + self.Password + " dbname=" + self.Name + " port=" + strconv.Itoa(self.Port) + " sslmode=disable application_name=" + appName + " TimeZone=" + timezone,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		NowFunc: func() time.Time {
			loc, _ := time.LoadLocation(timezone)
			layout := "2006-01-02 15:04:05"
			t1 := time.Now()
			t2, _ := time.Parse(layout, t1.Format(layout))
			now, _ := time.ParseInLocation(layout, t2.Format(layout), loc)
			now, _ = time.Parse(layout, now.Format(layout))
			return now
		},
	})
	if err != nil {
		log.Panicln(err)
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Panicln(err)
		panic(err)
	}

	// Atur parameter connection pool
	sqlDB.SetMaxOpenConns(1500)                // Maksimum 1500 koneksi terbuka (aktif + idle)
	sqlDB.SetMaxIdleConns(200)                 // Maksimum 200 koneksi idle
	sqlDB.SetConnMaxLifetime(60 * time.Minute) // Koneksi hidup maksimal 60 menit
	sqlDB.SetConnMaxIdleTime(15 * time.Minute) // Koneksi idle ditutup setelah 15 menit

	log.Println("DB '" + self.Name + "' Connected!")
	return db, err
}
