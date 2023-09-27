package routing

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/HeyyMrDJ/reflex/internal/redisreflex"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisClient *redis.Client

type Message struct {
    Action  string
    Flex    string
    Route   string
}

func getAllKeys() map[string]string {
    var cursor uint64
    result := make(map[string]string)
    
    for {
        var keys []string
        var err error
        keys, cursor, err = redisClient.Scan(ctx, cursor, "*", 0).Result()
        if err != nil {
            panic(err)
        }

        for _, key := range keys {
            val, _ := redisreflex.GetRedisValue(key, *&redisClient, ctx)
            result[key] = val
        }

        if cursor == 0 { // no more keys
            break
        }
    }

    return result
}


func getPinned() map[string]string{
    var pinnedFlexes = map[string]string{
        "ado": "https://dev.azure.com/",
        "gh": "https://github.com/",
        "yt": "https://www.youtube.com/",
    }
    return pinnedFlexes
}

func NotFound(w http.ResponseWriter, r *http.Request, route string) {
    // Set a custom message in the response body
    w.WriteHeader(http.StatusNotFound)
    notFoundCount.Inc()
    fmt.Fprintln(w, "Path not found:", route)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
    type Content struct {
        Flexes  map[string]string
        PinnedFlexes map[string]string
    }

    my_keys := getAllKeys()
    pinnedKeys := getPinned()
    content := Content{my_keys, pinnedKeys}


    tmpl, err := template.ParseFiles("templates/list.html", "templates/base.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if len(my_keys) < 1 {
        my_keys = map[string]string{"TEST": "ME"}
    }
    err = tmpl.ExecuteTemplate(w, "base.html", content)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func serveFlex(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    route := vars["route"]

    val, err := redisreflex.GetRedisValue(route, *&redisClient, ctx)
    if err != nil {
        NotFound(w, r, route)
        return
    }

    recordRouteHit(route) // Record the route hit
    http.Redirect(w, r, val, 301)
}

func createFlex(w http.ResponseWriter, r *http.Request) {
    //fmt.Println(r.Form.Get("route"))
    vars := mux.Vars(r)
    log.Print("Action: CREATE ", "Route: ", vars["route"]," Value: ", r.PostFormValue("flex"))
    route := vars["route"]
    if route == "" {
        route = r.Form.Get("route")
    }
    value := r.PostFormValue("flex")

    log.Print("Creating flex: ", route)
    err := redisreflex.SetRedisValue(route, value, *&redisClient, ctx)
    if err != nil {
        fmt.Fprintln(w, "Failed to set Redis value:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        log.Fatal("Failed to create route: ", route)
        return
    }

    log.Print("Flex created successfully: ", route)
    // Redirect to the home page with a 200 status code
    //fmt.Fprintln(w, "Redis value set successfully")
    //fmt.Fprintln(w, "Route:",route,"\tValue:", value)

    tmpl, err := template.ParseFiles("templates/body.html", "templates/base.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    msg := Message{"created", route, value}
    err = tmpl.ExecuteTemplate(w, "base.html", msg)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

}

func deleteFlex(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    route := vars["route"]
    value := r.PostFormValue("flex")
    if route == "" {
        route = r.PostFormValue("flex")
    }
    fmt.Println(value)
    log.Print("Action: DELETE ", "Route: ", vars["route"]," Value: ", r.PostFormValue("flex"))
    fmt.Println("deleteflex", r.PostFormValue("flex"))


    log.Print("Deleting flex: ", route)
    err := redisreflex.DeleteRedisValue(route, *&redisClient, ctx)
    if err != nil {
        fmt.Fprintln(w, "Failed to delete Redis value:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        log.Fatal("Failed to delete Redis value: ", route)
        return
    }

    log.Print("Flex deleted successfully:", route)
    // Redirect to the home page with a 200 status code
    tmpl, err := template.ParseFiles("templates/body.html", "templates/base.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    msg := Message{"deleted", route, value}
    err = tmpl.ExecuteTemplate(w, "base.html", msg)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func updateFlex(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    route := vars["route"]

    value := r.PostFormValue("flex")
    fmt.Println(value)

    log.Print("Updating flex: ", value)
    _, err := redisreflex.GetRedisValue(route, *&redisClient, ctx)
    if err != nil {
        NotFound(w, r, route)
        return
    }

    err = redisreflex.SetRedisValue(route, value, *&redisClient, ctx)
    if err != nil {
        fmt.Fprintln(w, "Failed to set Redis value:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        log.Fatal("Failed to update flex")
        return
    }

    log.Print("Flex updated successfully")
    tmpl, err := template.ParseFiles("templates/body.html", "templates/base.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    msg := Message{"updated", route, value}
    err = tmpl.ExecuteTemplate(w, "base.html", msg)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
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
    my_keys := getAllKeys()
    
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
    my_keys := getAllKeys()
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

func recordRouteHit(route string) {
    routeHits.WithLabelValues(route).Inc()
}


func InitializeRedisClient() {
    log.Print("Starting Redis Client")
    REDIS_ADDR := os.Getenv("REDIS_ADDR")
    REDIS_PORT := os.Getenv("REDIS_PORT")

    redisClient = redis.NewClient(&redis.Options{
        Addr:       REDIS_ADDR + ":" + REDIS_PORT, // Replace with your Redis server address
        Password: "",              // No password by default
        DB:       0,               // Default DB
    })
    log.Print("Redis Client started")
}

var (
    // Create a counter metric to track 404 responses.
    notFoundCount = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "myapp_http_404_total",
            Help: "Total number of 404 responses.",
        },
    )
)

var (
    // Create a counter metric to track route hits.
    routeHits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "myapp_route_hits_total",
            Help: "Total number of hits for dynamic routes.",
        },
        []string{"route"}, // This label will store the route name.
    )
)
