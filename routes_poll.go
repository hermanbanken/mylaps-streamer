package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func parseUnixString(unix string, orElse time.Time) time.Time {
	i, err := strconv.ParseInt(unix, 10, 64)
	if err != nil {
		return orElse
	}
	return time.Unix(i, 0)
}

var startOfTime = time.Time{}

// PollLastPublicRoute adds to saved db
func (api *API) PollLastPublicRoute(w http.ResponseWriter, r *http.Request) {
	users, err := getUsers(r)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	var since = parseUnixString(r.URL.Query().Get("lastEventId"), time.Time{})
	since.Add(1 * time.Second)
	log.Infof("Polling since %s - %s", since, r.URL.Query().Get("lastEventId"))

	// 1549391400
	var replayTime = parseUnixString(r.URL.Query().Get("replay"), time.Now())

	var wg sync.WaitGroup
	wg.Add(1)
	quitChannel := make(chan LastLap)
	emit := func(lastLap LastLap) {
		log.Infof("Emitting lap %v with finish %d", lastLap, lastLap.Finish.Unix())
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("%d=%s", lastLap.Finish.Unix(), durationToDigits(lastLap.Lap.Duration))))
		wg.Done()
		quitChannel <- lastLap
		log.Infof("Stopped")
	}

	logNoEmit := func(lastLap LastLap) {
		if lastLap.Lap == nil {
			log.Infof("Not emitting empty lap %v", lastLap)
		}
		if lastLap.Finish.Before(since) {
			log.Infof("Not emitting non-recent lap %v", lastLap)
		}
		if lastLap.Duration > 50*time.Second {
			log.Infof("Not emitting slow lap %v\n", lastLap)
		}
	}

	for _, usr := range users {
		go func(user SavedUser) {
			relativeTime := replayTime

			// Determine the last activity skated
			activity, isActivity := cache.GetOrElseWithin(fmt.Sprintf("LastActivity-%s", user.UserID), func() interface{} {
				list, _ := api.GetActivities(user.AccessToken, user.UserID, 1, 0, 0)
				if len(list) > 0 {
					return &list[0]
				}
				return Activity{}
			}, 30*time.Second).(*Activity)

			if !isActivity || activity.ID == 0 {
				return
			}

			getLastLap := func(startingLapIndex uint) LastLap {
				// This retrieves everything, so we cache this!!
				detail, err := api.GetActivity(user.AccessToken, activity.ID, 0, startingLapIndex, 0)
				if err != nil {
					return LastLap{nil, -1, time.Time{}, 0}
				}
				detail = detail.filterFast(relativeTime, 50*time.Second) // allow replaying
				var toSkipLater = 0
				for si, s := range detail.Sessions {
					for li, l := range s.Laps {
						toSkipLater++
						if si == len(detail.Sessions)-1 && li == len(s.Laps)-1 {
							return LastLap{&l, toSkipLater, l.FinishTime(), l.Duration}
						}
					}
				}
				return LastLap{nil, 0, time.Time{}, 0}
			}

			// Determine the last lap skated
			previousLap, isLap := cache.GetOrElseWithin(fmt.Sprintf("LastLap-%d-%s", activity.ID, user.UserID), func() interface{} {
				return getLastLap(0)
			}, 30*time.Second).(*LastLap)

			var lastID int
			if isLap && previousLap.Lap != nil && previousLap.Index > 1 {
				lastID = previousLap.Index - 1
			}

			lastLap := getLastLap(uint(lastID))
			if lastLap.Finish.After(since) && lastLap.Duration < 50*time.Second {
				emit(lastLap)
				return
			}
			logNoEmit(lastLap)

			for {
				select {
				case <-quitChannel:
					return
				default:
					log.Infof("Poll iteration %s", relativeTime)
					lastLap = getLastLap(uint(lastID))
					if lastLap.Lap != nil && lastLap.Finish.After(since) && lastLap.Duration < 50*time.Second {
						emit(lastLap)
						return
					}
					logNoEmit(lastLap)
					time.Sleep(1 * time.Second)
					relativeTime = relativeTime.Add(1 * time.Second)
				}
			}
		}(usr)
	}

	wg.Wait()
}

// LastLap captures the continuation token structure
type LastLap struct {
	Lap      *Lap
	Index    int
	Finish   time.Time
	Duration time.Duration
}
