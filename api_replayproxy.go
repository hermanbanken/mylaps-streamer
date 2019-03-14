package main

import (
	"time"
)

// ReplayAPI is same as API but it accepts a current time
type ReplayAPI struct {
	API
	now time.Time
}

type activityList []Activity

func (acts *activityList) filter(now time.Time) *[]Activity {
	var result []Activity
	for _, act := range *acts {
		if act.StartTime.After(now) {
			return nil
		}
		act.EndTime = now
		result = append(result, act)
	}
	return &result
}

func (detail *ActivityDetail) filter(now time.Time) *ActivityDetail {
	var cloned = detail
	var sessions []Session
	for _, session := range cloned.Sessions {
		if session.StartTime.Before(now) {
			var laps []Lap
			for _, lap := range session.Laps {
				if lap.FinishTime().Before(now) {
					laps = append(laps, lap)
				}
			}
			session.Laps = laps
			sessions = append(sessions, session)
		}
	}
	cloned.Sessions = sessions
	return cloned
}

func (detail *ActivityDetail) filterFast(now time.Time, maxDuration time.Duration) *ActivityDetail {
	var cloned = detail
	var sessions []Session
	for _, session := range cloned.Sessions {
		if session.StartTime.Before(now) {
			var laps []Lap
			for _, lap := range session.Laps {
				if lap.FinishTime().Before(now) && lap.Duration < maxDuration {
					laps = append(laps, lap)
				}
			}
			session.Laps = laps
			sessions = append(sessions, session)
		}
	}
	cloned.Sessions = sessions
	return cloned
}

// GetActivities gets the latest activities
func (api *ReplayAPI) GetActivities(accessToken, userID string, count, offset, requestID uint) ([]Activity, error) {
	list, err := api.API.GetActivities(accessToken, userID, count, offset, requestID)
	if err != nil {
		return nil, err
	}
	wrapped := activityList(list)
	return *wrapped.filter(api.now), nil
}

// GetActivity gets the latest activity details: stats, sessions & laps
func (api *ReplayAPI) GetActivity(accessToken string, activityID uint64) (*ActivityDetail, error) {
	detail, err := api.API.GetActivity(accessToken, activityID, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	return (*detail).filter(api.now), nil
}
