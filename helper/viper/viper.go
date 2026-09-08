package viper

import (
	"strings"

	"github.com/spf13/viper"
)

type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	Init()
	GetStringSlice(key string) []string
	GetStringMapString(key string) map[string]string
}

type config struct{}

func NewViper() Config {
	v := &config{}
	v.Init()
	return v
}

func (v *config) Init() {
	viper.SetEnvPrefix(`test`)
	viper.AutomaticEnv()

	replacer := strings.NewReplacer(`.`, `_`)
	viper.SetEnvKeyReplacer(replacer)
	viper.SetConfigType(`json`)
	viper.SetConfigFile(`config.json`)

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func (v *config) GetString(key string) string {
	return viper.GetString(key)
}

func (v *config) GetInt(key string) int {
	return viper.GetInt(key)
}

func (v *config) GetBool(key string) bool {
	return viper.GetBool(key)
}

func (v *config) GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func (v *config) GetStringMapString(key string) map[string]string {
	return viper.GetStringMapString(key)
}
