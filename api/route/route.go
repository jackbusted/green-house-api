package route

import (
	"green-house-api/helper"
	viperHelper "green-house-api/helper/viper"

	"gorm.io/gorm"
)

type NewRoute struct {
	DBMaster       *gorm.DB
	DBMainMaster   *gorm.DB
	DBReportMaster *gorm.DB
	Helper         helper.NewHelper
	Config         viperHelper.Config
}
