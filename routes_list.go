package main

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var loadedUsers = false

// GoPublicRoute adds to saved db
func (api *API) GoPublicRoute(w http.ResponseWriter, r *http.Request) {
	accessToken, _, ok := getAuth(w, r)
	if !ok {
		w.WriteHeader(401)
		w.Write([]byte("Not Authorized"))
		return
	}
	profile, _ := api.GetClaims(accessToken)
	if profile != nil {
		storeUser(SavedUser{profile.NickName, profile.UserID, accessToken, time.Now()}, r)
		http.Redirect(w, r, "/", 301)
	} else {
		w.WriteHeader(401)
		w.Write([]byte("Not Authorized"))
	}
}

// ListStreamsRoute streams events
func (api *API) ListStreamsRoute(w http.ResponseWriter, r *http.Request) {
	users, err := getUsers(r)
	if err == nil {
		loadedUsers = true
	} else {
		w.WriteHeader(500)
		return
	}

	type listedUser struct {
		UserID     string `json:"userId"`
		ActivityID uint64 `json:"activityId"`
	}
	var results []listedUser
	for _, user := range users {
		activities, err := api.GetActivities(user.AccessToken, user.UserID, 1, 0, 0)
		if err == nil && len(activities) > 0 {
			results = append(results, listedUser{user.UserID, activities[0].ID})
		}
	}

	if results == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		log.Errorf("JSON write error %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
