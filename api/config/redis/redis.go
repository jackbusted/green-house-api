package redis

import (
	"context"
	"fmt"
	viperHelper "green-house-api/helper/viper"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var (
	redisClient *redis.Client
	ctx         context.Context
)

type RedisConfig struct {
	Host     string
	Password string
}

func Connect(zLog *zerolog.Logger, config viperHelper.Config) error {
	zLog.Info().Msg("Connecting to Redis")
	zLog.Info().Msg("REDIS_HOST : " + config.GetString("redis.host"))
	zLog.Info().Msg("REDIS_PASSWORD : " + config.GetString("redis.password"))

	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.GetString("redis.host"),
		Password: config.GetString("redis.password"),
		DB:       0,
	})

	zLog.Info().Msg("REDIS CONTEXT WITH TIMEOUT ")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zLog.Info().Msg("REDIS CLIENT PING ")
	_, err := redisClient.Ping(ctx).Result()
	zLog.Info().Msg("REDIS CLIENT PING " + fmt.Sprint(redisClient))
	if err != nil {
		zLog.Error().Err(err).Msg("Failed to connect to Redis")
		return err
	}

	zLog.Info().Msg("Successfully connected to Redis")
	return nil
}

func GetRedis() *redis.Client {
	return redisClient
}
