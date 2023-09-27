package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/HeyyMrDJ/reflex/internal/routing"
)



func main() {
    RE_PORT := os.Getenv("RE_PORT")
    routing.GetMostUsed()
    router := mux.NewRouter()
    routing.ConfigureRoutes(router)

    http.Handle("/", router)
    http.Handle("/metrics", promhttp.Handler())

    fmt.Println("Serving on port:", RE_PORT)
    log.Fatal(http.ListenAndServe(":" + RE_PORT, nil))
}
