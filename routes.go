package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var cache = Cache{}

// GetTokenRoute calls GetToken with correct arguments from the HTTP request
func (api *API) GetTokenRoute(w http.ResponseWriter, r *http.Request) {
	if user, pass, ok := r.BasicAuth(); ok {
		log.Infof("Basic auth login")
		token, err := api.GetToken(user, pass)
		if err == nil {
			var expiration = time.Now().Add(365 * 24 * time.Hour)
			var Host = r.Host
			log.Infof("Using '%s' as cookie domain", Host)
			var cookieA = http.Cookie{Name: "accesstoken", Value: token.AccessToken, Expires: expiration, Path: "/"}
			http.SetCookie(w, &cookieA)
			var cookieB = http.Cookie{Name: "userid", Value: token.UserID, Expires: expiration, Path: "/"}
			http.SetCookie(w, &cookieB)
			http.Redirect(w, r, r.Referer(), 307)
			jsonResult, _ := json.Marshal(token)
			defer log.Printf(string(jsonResult))
			return
		}
		http.Error(w, fmt.Sprintf("Something went wrong: %v", err), http.StatusUnauthorized)
	} else {
		log.Infof("Invalid login")
		w.Header().Set("WWW-Authenticate", "Basic realm=MyLaps")
		http.Error(w, "Authorization failed", http.StatusUnauthorized)
		return
	}
}

// GetEventStreamRoute streams events
func (api *API) GetEventStreamRoute(w http.ResponseWriter, r *http.Request) {
	accessToken, _, ok := getAuth(w, r)
	if !ok {
		return
	}
	api.GetEventStream(w, r, accessToken, 4)
}

// GetEventStreamOfOtherRoute streams events
func (api *API) GetEventStreamOfOtherRoute(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userid")
	users, err := getUsers(r)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	if savedUser, ok := users[userID]; ok {
		api.GetEventStream(w, r, savedUser.AccessToken, 1)
	} else {
		w.WriteHeader(404)
	}
}

// GetEventStream retrieves anyones stream
func (api *API) GetEventStream(w http.ResponseWriter, r *http.Request, accessToken string, replayActivityAmount uint) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(200)
	w.Write([]byte("retry: 0\n\n"))
	start := time.Now()

	// Ensure to stop if the client disconnects
	var stop = make(chan int)
	notifier, hasNotify := w.(http.CloseNotifier)
	if hasNotify { // on AppEngine this is false
		notify := notifier.CloseNotify()
		go func() {
			<-notify
			stop <- 1
		}()
	}

	lastEventID := r.Header.Get("Last-Event-ID")
	userID, activityID, lapID := parseEventID(lastEventID)
	if lapID > 0 {
		replayActivityAmount = 1
	}
	stream := api.StreamEvents(bootstrap{accessToken, userID, activityID, lapID, replayActivityAmount}, stop)

	var hasSendData = false
	var write = func(data string) {
		w.Write([]byte(data))
		tryFlush(w)
	}

	for e := range stream {
		log.Debugf("Event: %v", e)
		if profile, isProfile := e.Data.(*ProfileJSON); isProfile {
			data := fmt.Sprintf(`{ "type": "profile", "name": "%s" }`, profile.NickName)
			write(fmt.Sprintf("id: %s\nevent: update\ndata: %s\n\n\n", profile.UserID, data))
			userID = profile.UserID
		}
		if act, isActivity := e.Data.(Activity); isActivity {
			data := fmt.Sprintf(`{ "type": "activity", "id": "%d", "start": "%s", "location": { "name": "%s", "trackLength": %d, "sport": "%s" } }`,
				act.ID, act.StartTime.Format(time.RFC3339),
				act.Location.Name, act.Location.TrackLength, act.Location.Sport)
			write(fmt.Sprintf("id: %s:%d\nevent: update\ndata: %s\n\n\n", userID, act.ID, data))
			activityID = act.ID
		}
		if lap, isLap := e.Data.(Lap); isLap {
			data := fmt.Sprintf(`{ "type": "lap", "duration": "%s", "start": "%s" }`, lap.Duration, lap.StartTime.Format(time.RFC3339))
			write(fmt.Sprintf("id: %s:%d:%d\nevent: update\ndata: %s\n\n\n", userID, activityID, lapID, data))
			lapID++
			hasSendData = true
		}
		if err, isError := e.Data.(error); isError {
			data := fmt.Sprintf(`{ "type": "error", "message": "%s" }`, err)
			write(fmt.Sprintf("event: update\ndata: %s\n\n\n", data))
		}

		// AppEngine longpolling:
		_, isWait := e.Data.(Waiting)
		if isWait && (hasSendData || time.Since(start) > 10*time.Second) {
			// We got data after some time: now stop stream and cause the actual flush
			stop <- 1
		}
	}
}

// GetLastRoute polls
func (api *API) GetLastRoute(w http.ResponseWriter, r *http.Request) {
	accessToken, userID, ok := getAuth(w, r)
	if !ok || userID == "" {
		http.Redirect(w, r, "/", 307)
		return
	}

	var activityID uint64
	var err error

	// Last activity is cached for 1 minute
	var result = &Activity{}
	var isActivity bool
	result, isActivity = cache.GetOrElseWithin(fmt.Sprintf("Activities-%s", userID), func() interface{} {
		list, _ := api.GetActivities(accessToken, userID, 3, 0, 0)
		if len(list) > 0 {
			return &list[0]
		}
		return Activity{}
	}, 1*time.Minute).(*Activity)

	if isActivity && result != nil {
		activityID = result.ID
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), 500)
		return
	}

	if activityID == 0 {
		http.NotFound(w, r)
		return
	}

	var activity *ActivityDetail
	activity, err = api.GetActivity(accessToken, activityID, 0, 0, 0)
	ss := activity.Sessions
	ll := ss[len(ss)-1].Laps
	last := ll[len(ll)-1]

	if accept := r.Header.Get("Accept"); accept != "" {
		data := fmt.Sprintf(`{"laps": %d, "lastStart": "%s", "lastDuration": "%s" }`, activity.Stats.LapCount, last.StartTime.Format(time.RFC3339), last.Duration)
		w.Write([]byte(data))
	} else {
		data := fmt.Sprintf(`%d=%s|%s`, activity.Stats.LapCount, durationToDigits(last.Duration), last.StartTime.Format(time.RFC3339))
		w.Write([]byte(data))
	}
}
