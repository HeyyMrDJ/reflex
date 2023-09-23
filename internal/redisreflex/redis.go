package redisreflex

import (
    "github.com/redis/go-redis/v9"
    "context"
    "os"
)

var ctx = context.Background()
var redisClient *redis.Client

func InitializeRedisClient() {
    REDIS_ADDR := os.Getenv("REDIS_ADDR")
    REDIS_PORT := os.Getenv("REDIS_PORT")

    redisClient = redis.NewClient(&redis.Options{
        Addr:       REDIS_ADDR + ":" + REDIS_PORT, // Replace with your Redis server address
        Password: "",              // No password by default
        DB:       0,               // Default DB
    })
}

func SetRedisValue(key string, value string) error {
    err := redisClient.Set(context.Background(), key, value, 0).Err()
    return err
}

func GetRedisValue(key string) (string, error) {
    val, err := redisClient.Get(context.Background(), key).Result()
    if err != nil {
        return "", err
    }
    return val, nil
}

func DeleteRedisValue(key string) error {
    _, err := redisClient.Del(context.Background(), key).Result()
    if err != nil {
        return err
    }
    return nil
}


