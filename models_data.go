package main

import "time"

// Activity is a grouper for sessions
type Activity struct {
	ID        uint64           `json:"id"`
	Name      string           `json:"name"`
	StartTime time.Time        `json:"startTime"`
	EndTime   time.Time        `json:"endTime"`
	Location  ActivityLocation `json:"location"`
}

// ActivityLocation docs
type ActivityLocation struct {
	Name        string `json:"name"`
	TrackLength uint   `json:"trackLength"`
	Sport       string `json:"sport"`
}

// ActivityDetail contains stats & sessions
type ActivityDetail struct {
	Stats    ActivityStats `json:"stats"`
	Sessions []Session     `json:"sessions"`
}

// ActivityStats contains the lap count & more
type ActivityStats struct {
	LapCount uint `json:"lapCount"`
}

// Session contains the laps
type Session struct {
	ID        uint64        `json:"id"`
	StartTime time.Time     `json:"dateTimeStart"`
	Duration  time.Duration `json:"duration"`
	Laps      []Lap         `json:"laps"`
}

// Lap has a start time & duration
type Lap struct {
	Nr        uint          `json:"nr"`
	SessionID uint          `json:"sessionID"`
	StartTime time.Time     `json:"dateTimeStart"`
	Duration  time.Duration `json:"duration"`
}

// FinishTime calculates the finish time: time of event
func (lap *Lap) FinishTime() time.Time {
	return lap.StartTime.Add(lap.Duration)
}
