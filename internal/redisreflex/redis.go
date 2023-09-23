package redisreflex

import (
    "github.com/redis/go-redis/v9"
    "context"
)

func SetRedisValue(key string, value string, redisClient *redis.Client, ctx context.Context) error {
    err := redisClient.Set(ctx, key, value, 0).Err()
    return err
}

func GetRedisValue(key string, redisClient *redis.Client, ctx context.Context) (string, error) {
    val, err := redisClient.Get(ctx, key).Result()
    if err != nil {
        return "", err
    }
    return val, nil
}

func DeleteRedisValue(key string, redisClient *redis.Client, ctx context.Context) error {
    _, err := redisClient.Del(ctx, key).Result()
    if err != nil {
        return err
    }
    return nil
}


