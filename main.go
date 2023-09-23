package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/HeyyMrDJ/reflex/internal/routing"
    "github.com/gorilla/mux"
    "github.com/redis/go-redis/v9"

)

var redisClient *redis.Client


func main() {
    RE_PORT := os.Getenv("RE_PORT")
    // Initialize the Redis client
    InitializeRedisClient()
    router := mux.NewRouter()
    routing.configureRoutes(router)

   

    fmt.Println("Serving on port:", RE_PORT)
    log.Fatal(http.ListenAndServe(":" + RE_PORT, nil))
}


func InitializeRedisClient() {
    REDIS_ADDR := os.Getenv("REDIS_ADDR")
    REDIS_PORT := os.Getenv("REDIS_PORT")

    redisClient = redis.NewClient(&redis.Options{
        Addr:       REDIS_ADDR + ":" + REDIS_PORT, // Replace with your Redis server address
        Password: "",              // No password by default
        DB:       0,               // Default DB
    })
}


