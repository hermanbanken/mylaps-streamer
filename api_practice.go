package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// GetActivities gets the latest activities
func (api *API) GetActivities(accessToken, userID string, count, offset, requestID uint) ([]Activity, error) {
	url := fmt.Sprintf("https://%s/accounts/%s/training/activities?count=%d&offset=%d&requestId=%d", api.Endpoints.Practice, userID, count, offset, requestID)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("apiKey", api.APIKey)
	req.Header.Add("User-Agent", "mylaps-streamer/1.0.0")

	res, err := netClient.Do(req)
	if res == nil {
		return nil, err
	}
	log.Debugf("[GetActivities] HTTP%d", res.StatusCode)

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var body struct {
			Activities []Activity `json:"activities"`
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		if err != nil {
			log.Errorf("[GetActivities] Unmarshal incorrect: %s", err)
			return nil, err
		}
		log.Infof("[GetActivities] Result: length=%d", len(body.Activities))
		return body.Activities, nil
	}

	// Other statuscodes
	log.Debugf("curl -X POST %s -f -H 'Authorization: Bearer %s' -H 'apiKey: %s'", url, accessToken, api.APIKey)
	return nil, errors.New(string(fmt.Sprintf("%s", res.Status)))
}

// GetActivity gets the latest activity details: stats, sessions & laps
func (api *API) GetActivity(accessToken string, activityID uint64, count, numlaps, offset uint) (*ActivityDetail, error) {
	url := fmt.Sprintf("https://%s/training/activities/%d/sessions?count=%d&numlaps=%d&offset=%d&requestId=0", api.Endpoints.Practice, activityID, count, numlaps, offset)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("apiKey", api.APIKey)
	req.Header.Add("User-Agent", "mylaps-streamer/1.0.0")

	res, err := netClient.Do(req)
	if res == nil {
		return nil, err
	}
	log.Debugf("[GetActivity] HTTP%d", res.StatusCode)

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var body apiActivityDetail
		err = json.NewDecoder(res.Body).Decode(&body)
		if err != nil {
			log.Errorf("[GetActivity] Unmarshal incorrect: %s", err)
			return nil, err
		}
		detail := body.ToActivity()
		contained := 0
		for _, session := range detail.Sessions {
			contained += len(session.Laps)
		}
		log.Debugf("[GetActivity] laps=%d,of which contained=%d", detail.Stats.LapCount, contained)
		return &detail, nil
	}

	// Other statuscodes
	log.Debugf("curl -X POST %s -f -H 'Authorization: Bearer %s' -H 'apiKey: %s'", url, accessToken, api.APIKey)
	return nil, errors.New(string(fmt.Sprintf("%s", res.Status)))
}

type apiActivityDetail struct {
	Stats    ActivityStats `json:"stats"`
	Sessions []apiSession  `json:"sessions"`
}

func (d *apiActivityDetail) ToActivity() ActivityDetail {
	var sessions = make([]Session, len(d.Sessions))
	for i, s := range d.Sessions {
		sessions[i] = s.ToSession()
	}
	return ActivityDetail{
		Stats:    d.Stats,
		Sessions: sessions,
	}
}

type apiSession struct {
	ID        uint64    `json:"id"`
	StartTime time.Time `json:"dateTimeStart"`
	Duration  string    `json:"duration"`
	Laps      []apiLap  `json:"laps"`
}

func (d *apiSession) ToSession() Session {
	var laps = make([]Lap, len(d.Laps))
	for i, s := range d.Laps {
		laps[i] = s.ToLap()
	}
	duration, _ := parseDuration(d.Duration)
	return Session{
		ID:        d.ID,
		StartTime: d.StartTime,
		Duration:  duration,
		Laps:      laps,
	}
}

// Lap has a start time & duration
type apiLap struct {
	Nr        uint      `json:"nr"`
	StartTime time.Time `json:"dateTimeStart"`
	Duration  string    `json:"duration"`
}

func (d *apiLap) ToLap() Lap {
	duration, _ := parseDuration(d.Duration)
	return Lap{
		Nr:        d.Nr,
		StartTime: d.StartTime,
		Duration:  duration,
	}
}
