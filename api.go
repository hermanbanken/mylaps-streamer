package main

import (
	"encoding/base64"
	"net/http"
	"time"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func (api *API) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(200)
	res.Write([]byte("API!"))
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
