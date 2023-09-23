package routing

import (
    "github.com/gorilla/mux"
    "net/http"
)

func ConfigureRoutes(router *mux.Router){
     // Serve static files (CSS, JS, etc.) from the "static" directory.
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Define your other routes here.
    router.HandleFunc("/", serveHome)
    router.HandleFunc("/create", serveCreate)
    router.HandleFunc("/update", serveUpdate)
    router.HandleFunc("/deleteme", deleteFlex).Methods("POST")
    router.HandleFunc("/delete", serveDelete).Methods("GET")
    router.HandleFunc("/set", createFlex).Methods("POST")
    router.HandleFunc("/set/{route}", createFlex).Methods("POST")
    router.HandleFunc("/{route}", serveFlex).Methods("GET")
    router.HandleFunc("/{route}", createFlex).Methods("POST")
    router.HandleFunc("/{route}", deleteFlex).Methods("DELETE")
    router.HandleFunc("/{route}", updateFlex).Methods("PUT")

    InitializeRedisClient()
}
