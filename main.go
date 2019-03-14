package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
)

// API contains all paths and secrets needed to access the API
type API struct {
	Endpoints struct {
		Users    string
		Practice string
	}
	AppKey    string
	AppSecret string
	APIKey    string
}

var api = API{
	Endpoints: struct {
		Users    string
		Practice string
	}{"usersandproducts-api.speedhive.com", "practice-api.speedhive.com/api/v1"},
}

var httpPort uint

func init() {
	// Signify options
	flag.StringVar(&api.AppKey, "appkey", "SpeedhiveIosApp", "AppKey")
	flag.StringVar(&api.AppSecret, "appsecret", "VYvyLnu5Egwwr8CwxQemBTaOjfbs8MiYyVIvMGOS", "AppSecret")
	flag.StringVar(&api.APIKey, "apikey", "SpeedhiveIosApp-91be0d68-294c-4281-aa12-df55275a51cd", "ApiKey")

	// Application options
	PORT, err := strconv.ParseUint(os.Getenv("PORT"), 10, 64)
	if err != nil {
		PORT = 8080
	}
	flag.UintVar(&httpPort, "port", uint(PORT), "port to bind to")
}

func main() {
	flag.Parse()

	// AppEngine
	http.HandleFunc("/_ah/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	// API
	http.HandleFunc("/api/login", api.GetTokenRoute)
	http.HandleFunc("/api/public", api.GoPublicRoute)
	http.HandleFunc("/api/streams", api.ListStreamsRoute)
	http.HandleFunc("/api/stream/me", api.GetEventStreamRoute)
	http.HandleFunc("/api/stream/other", api.GetEventStreamOfOtherRoute)
	http.HandleFunc("/api/poll", api.GetLastRoute)
	http.HandleFunc("/api/poll/public", api.PollLastPublicRoute)
	http.Handle("/api/", &api)

	// Statics during localdev
	backToHome := func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", 301) }
	http.HandleFunc("/docs", backToHome)
	http.HandleFunc("/docs/", backToHome)
	fs := http.FileServer(http.Dir("docs"))
	http.Handle("/", fs)

	start(httpPort)
}
