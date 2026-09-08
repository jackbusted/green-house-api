package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	redisConfig "green-house-api/api/config/redis"
	"green-house-api/helper/logger"
	viperHelper "green-house-api/helper/viper"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheType string

const (
	LocalCache CacheType     = "local" // Use local in-memory cache
	RedisCache CacheType     = "redis" // Use Redis
	ttl        time.Duration = 24 * time.Hour
)

var localCache sync.Map
var redisClient *redis.Client
var cacheType CacheType
var prefixKey string

func InitCache(config viperHelper.Config) {
	redisClient = redisConfig.GetRedis()
	prefixKey = config.GetString("app.env") + ":" + config.GetString("app.name")

	switch config.GetString("app.cache_type") {
	case "redis":
		logger.Default().Print("Cache Type: Redis")
		cacheType = RedisCache
	default:
		logger.Default().Print("Cache Type: Local")
		cacheType = LocalCache
	}
}

func SetCache(key string, value any, t time.Duration) error {
	redisClient = redisConfig.GetRedis()
	key = prefixKey + ":" + key

	// Convert struct to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return errors.New("Set Cache Error encoding JSON: " + err.Error())
	}
	tempTtl := t
	if tempTtl <= 0 {
		tempTtl = ttl
	}
	msgTtl := "duration cache => "
	if tempTtl.Hours()/24 >= 1 {
		msgTtl += fmt.Sprint(tempTtl.Hours()/24, "d ")
		msgTtl += fmt.Sprint(int(tempTtl.Hours())%24, "h ")
	} else {
		msgTtl += fmt.Sprint(tempTtl.Hours(), "h ")
	}
	logger.Default().Println(msgTtl)
	switch cacheType {
	case RedisCache:
		err := redisClient.Set(context.Background(), key, data, tempTtl).Err()
		if err != nil {
			return err
		}
	default: // LocalCache
		localCache.Store(key, data)

		logger.Default().Println("SetCache", key)
	}
	return nil
}

func GetCache(key string, data interface{}) error {
	key = prefixKey + ":" + key
	var val any
	var err error
	var found bool
	redisClient = redisConfig.GetRedis()
	// zlog.Info().Msg("redisClientGet " + fmt.Sprint(redisClient))
	switch cacheType {
	case RedisCache:
		val, err = redisClient.Get(context.Background(), key).Result()
		if err != nil {
			return err
		}
		log.Println("GetCache", key, "val", val.(string))
		err = json.Unmarshal([]byte(val.(string)), &data)
		if err != nil {
			return errors.New("Get Cache Error decoding JSON: " + err.Error())
		}
	default: // LocalCache
		if val, found = localCache.Load(key); !found {
			return errors.New("data not found")
		}

		err = json.Unmarshal(val.([]byte), &data)
		if err != nil {
			return errors.New("Get Cache Error decoding JSON: " + err.Error())
		}
	}

	return nil
}

func DeleteCache(key string) {
	key = prefixKey + ":" + key
	switch cacheType {
	case RedisCache:
		redisClient = redisConfig.GetRedis()
		logger.Default().Println("DeleteCache", key)
		logger.Default().Println("redisClient", redisClient)
		res, err := redisClient.Del(context.Background(), key).Result()

		if err != nil {
			logger.Default().Println("Error deleting cache:", err)
			return
		}

		println("Number of keys deleted:", res)
	default: // LocalCache
		localCache.Delete(key)
	}
}

func GetCacheType() CacheType {
	return cacheType
}
