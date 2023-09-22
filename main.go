package main

import (
	"fmt"
	"html/template"
	"log"
	"context"
	"net/http"

    "github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
    "io/ioutil"
)

const PORT string = ":9069"
var ctx = context.Background()
var redisClient *redis.Client

func main() {
    // Initialize the Redis client
    InitializeRedisClient()
	router := mux.NewRouter()

	// Serve static files (CSS, JS, etc.) from the "static" directory.
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Define your other routes here.
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/create", serveCreate)
	router.HandleFunc("/update", serveUpdate)
	router.HandleFunc("/deleteme", deleteme).Methods("POST")
	router.HandleFunc("/delete", serveDelete).Methods("GET")
	router.HandleFunc("/set", createFlex).Methods("POST")
	router.HandleFunc("/{route}", serveFlex).Methods("GET")
	router.HandleFunc("/{route}", createFlex).Methods("POST")
	router.HandleFunc("/{route}", deleteFlex).Methods("DELETE")
	router.HandleFunc("/{route}", updateFlex).Methods("PUT")

	http.Handle("/", router)

    fmt.Println("Serving on port:", PORT)
	log.Fatal(http.ListenAndServe(":9069", nil))
}

func NotFound(w http.ResponseWriter, r *http.Request, route string) {
    // Set a custom message in the response body
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprintln(w, "Path not found:", route)
}

func InitializeRedisClient() {
    redisClient = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379", // Replace with your Redis server address
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

func DeleteRedisValue(key string) (error) {
    _, err := redisClient.Del(context.Background(), key).Result()
    if err != nil {
        return err
    }
    return nil
}

func serveHome(w http.ResponseWriter, r *http.Request) {
    type keyValue struct {
        Key     string
        Value   string
    }
    var cursor uint64
    var my_keys []keyValue
    for {
        var keys []string
        var err error
        keys, cursor, err = redisClient.Scan(ctx, cursor, "*", 0).Result()
        if err != nil {
            panic(err)
        }

        for _, key := range keys {
            val, _ := GetRedisValue(key)
            kv := keyValue {key, val}
            fmt.Println("key", val)
            my_keys = append(my_keys, kv)
        }

        if cursor == 0 { // no more keys
            break
        }
    }

	tmpl, err := template.ParseFiles("templates/list.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    if len(my_keys) < 1 {
        kv := keyValue {"TEST", "ME"}
        my_keys = append(my_keys, kv)
    }
	err = tmpl.ExecuteTemplate(w, "base.html", my_keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func serveFlex(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
	route := vars["route"]
    val, err := GetRedisValue(route)
    if err != nil {
        NotFound(w, r, route)
    }

    http.Redirect(w, r, val, 301)
}

func createFlex(w http.ResponseWriter, r *http.Request) {
    //vars := mux.Vars(r)
    //route := vars["route"]
    
    // Extract the value to set from the request body
    value := r.PostFormValue("flex")
    //route := r.PostFormValue("route")
    route := r.Form.Get("route")
    fmt.Println(route, value)
    
    err := SetRedisValue(route, value)
    if err != nil {
        fmt.Fprintln(w, "Failed to set Redis value:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Redirect to the home page with a 200 status code
    fmt.Fprintln(w, "Redis value set successfully")
    fmt.Fprintln(w, route, value)
}

func deleteFlex(w http.ResponseWriter, r *http.Request) {
    fmt.Println("deleteflex", r.PostFormValue("flex"))
    vars := mux.Vars(r)
    route := vars["route"]

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        // Handle the error
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    fmt.Println("Request Body:", string(body))
    
    // Extract the value to set from the request body
    value := r.PostFormValue("flex")
    fmt.Println(value)
    
    err = DeleteRedisValue(route)
    if err != nil {
        fmt.Fprintln(w, "Failed to delete Redis value:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Redirect to the home page with a 200 status code
    fmt.Fprintln(w, "Redis value deleted successfully")
    fmt.Fprintln(w, route)
}

func updateFlex(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    route := vars["route"]
    
    // Extract the value to set from the request body
    value := r.PostFormValue("flex")
    fmt.Println(value)

    _, err := GetRedisValue(route)
    if err != nil {
        NotFound(w, r, route)
        return
    }
    
    err = SetRedisValue(route, value)
    if err != nil {
        fmt.Fprintln(w, "Failed to set Redis value:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Redirect to the home page with a 200 status code
    fmt.Fprintln(w, "Redis value updated successfully")
    fmt.Fprintln(w, route)
}

func serveCreate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func serveUpdate(w http.ResponseWriter, r *http.Request) {
    type keyValue struct {
        Key     string
        Value   string
    }
    var cursor uint64
    var my_keys []keyValue
    for {
        var keys []string
        var err error
        keys, cursor, err = redisClient.Scan(ctx, cursor, "*", 0).Result()
        if err != nil {
            panic(err)
        }

        for _, key := range keys {
            val, _ := GetRedisValue(key)
            kv := keyValue {key, val}
            fmt.Println("key", val)
            my_keys = append(my_keys, kv)
        }

        if cursor == 0 { // no more keys
            break
        }
    }

	tmpl, err := template.ParseFiles("templates/update.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", my_keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func serveDelete(w http.ResponseWriter, r *http.Request) {
    fmt.Println("serveDelete")
    type keyValue struct {
        Key     string
        Value   string
    }
    var cursor uint64
    var my_keys []keyValue
    for {
        var keys []string
        var err error
        keys, cursor, err = redisClient.Scan(ctx, cursor, "*", 0).Result()
        if err != nil {
            panic(err)
        }

        for _, key := range keys {
            val, _ := GetRedisValue(key)
            kv := keyValue {key, val}
            fmt.Println("key", val)
            my_keys = append(my_keys, kv)
        }

        if cursor == 0 { // no more keys
            break
        }
    }

	tmpl, err := template.ParseFiles("templates/delete.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", my_keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
func deleteme(w http.ResponseWriter, r *http.Request) {
    fmt.Println("deleteflex", r.PostFormValue("flex"))
    vars := mux.Vars(r)
    route := vars["route"]
    
    // Extract the value to set from the request body
    value := r.PostFormValue("flex")
    fmt.Println("VALUE is:",value)
    
    err := DeleteRedisValue(value)
    if err != nil {
        fmt.Fprintln(w, "Failed to delete Redis value:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Redirect to the home page with a 200 status code
    fmt.Fprintln(w, "Redis value deleted successfully")
    fmt.Fprintln(w, route)
}

