package main

import (
	"fmt"
	"time"

	"mylaps-streamer/multicast"

	log "github.com/sirupsen/logrus"
)

type bootstrap struct {
	AccessToken          string
	UserID               string
	ActivityID           uint64
	LapID                uint
	ReplayActivityAmount uint
}

// Waiting indicates the stream will pause for some duration for a delayed poll
type Waiting struct {
	Duration time.Duration
}

// StreamEvents creates a channel that generates live events
func (api *API) StreamEvents(bootstrap bootstrap, stop <-chan int) (channel <-chan multicast.Event) {
	logSafeBoostrap := bootstrap
	logSafeBoostrap.AccessToken = "*redacted*"
	log.Infof("Starting stream %+v", logSafeBoostrap)
	var out = make(chan multicast.Event)
	channel = out

	var accessToken = bootstrap.AccessToken
	var activityID = bootstrap.ActivityID
	var userID = bootstrap.UserID
	var lastReadLap Lap
	lastReadLap.Nr = bootstrap.LapID

	go func() {
		if accessToken == "" {
			out <- multicast.Event{T: time.Now(), Data: fmt.Errorf("Missing accessToken")}
			close(out)
			return
		}

		// It would be very weird to include all laps of N old events & then skip the first X laps of the latest event.
		if bootstrap.LapID > 0 && bootstrap.ReplayActivityAmount > 1 {
			out <- multicast.Event{T: time.Now(), Data: fmt.Errorf("Either specify LapID > 0 OR ReplayActivityAmount > 1, results would be weird otherwise")}
			close(out)
			return
		}

		// Start by getting the userID if not present
		if userID == "" {
			profile, _ := api.GetClaims(accessToken)
			if profile != nil {
				out <- multicast.Event{T: time.Time{}, Data: profile}
				userID = profile.UserID
			}
			log.Infof("Got profile %s", profile)
		}

		// Still will be called each time
		readNextBatch := func(activityID uint64, lastLap *Lap) {
			activity, err := api.GetActivity(accessToken, activityID, 0, lastLap.Nr, 0)
			if err != nil {
				out <- multicast.Event{T: time.Now(), Data: err}
				close(out)
				return
			}
			for _, session := range activity.Sessions {
				for _, lap := range session.Laps {
					lap.Nr = lastLap.Nr + 1
					*lastLap = lap
					out <- multicast.Event{T: lap.FinishTime(), Data: lap}
				}
			}
			return
		}

		// We have no starting activity: so read 2 fully!
		if activityID <= 0 {
			log.Info("Getting initial activities")
			activities, err := api.GetActivities(accessToken, userID, bootstrap.ReplayActivityAmount, 0, 0)
			log.Infof("Activities: %d/%d", len(activities), bootstrap.ReplayActivityAmount)
			for len(activities) == 0 {
				if err != nil {
					out <- multicast.Event{T: time.Now(), Data: err}
					close(out)
					return
				}
				log.Info("Waiting on activities to start")
				wait(5*time.Second, out)
			}

			// Start with oldest activity
			reverse(activities)
			for i, act := range activities {
				log.Info("Processing activity")
				activityID = act.ID
				out <- multicast.Event{T: act.StartTime, Data: act}
				// Read all but last: that might still be ongoing, so see below
				if i != len(activities)-1 {
					lastReadLap.Nr = 0 // fresh activity: fresh lap count
					readNextBatch(act.ID, &lastReadLap)
				} else {
					lastReadLap.Nr = bootstrap.LapID // resume from bootstrapped LapID
				}
			}
		}

		// Keep polling last activity
		for {
			select {
			case <-stop:
				log.Infof("Stopping stream")
				close(out)
				return
			default:
				readNextBatch(activityID, &lastReadLap)
				wait(time.Second, out)
			}
		}
	}()

	return
}

func wait(td time.Duration, out chan<- multicast.Event) {
	out <- multicast.Event{T: time.Now(), Data: Waiting{td}}
	time.Sleep(td)
}
