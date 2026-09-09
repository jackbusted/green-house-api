package helper

import (
	mqttClient "green-house-api/api/mqtt"
	"green-house-api/helper/jwt"
	"green-house-api/helper/postgre"
	"green-house-api/helper/response"
	"green-house-api/helper/validator"
	"green-house-api/helper/viper"
)

type NewHelper struct {
	Response   response.ResponseHelper
	Validator  *validator.Validator
	Config     viper.Config
	Jwt        jwt.JwtHelper
	DB         postgre.Database
	MQTTClient *mqttClient.MqttClient
}
