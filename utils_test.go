package main

import "testing"

func TestParseEventIDFull(t *testing.T) {
	userID, activityID, lapID := parseEventID("MYLAPS-GA-yada:12345678:34")
	if userID != "MYLAPS-GA-yada" {
		t.Errorf("Got %s, expected MYLAPS-GA-yada", userID)
	}
	if activityID != 12345678 {
		t.Errorf("Got %d, expected 12345678", activityID)
	}
	if lapID != 34 {
		t.Errorf("Got %d, expected 34", lapID)
	}
}

func TestParseEventIDNoLap(t *testing.T) {
	userID, activityID, lapID := parseEventID("MYLAPS-GA-yada:12345678")
	if userID != "MYLAPS-GA-yada" {
		t.Errorf("Got %s, expected MYLAPS-GA-yada", userID)
	}
	if activityID != 12345678 {
		t.Errorf("Got %d, expected 12345678", activityID)
	}
	if lapID != 0 {
		t.Errorf("Got %d, expected 0", lapID)
	}
}

func TestParseEventIDNoActivity(t *testing.T) {
	userID, activityID, lapID := parseEventID("MYLAPS-GA-yada")
	if userID != "MYLAPS-GA-yada" {
		t.Errorf("Got %s, expected MYLAPS-GA-yada", userID)
	}
	if activityID != 0 {
		t.Errorf("Got %d, expected 0", activityID)
	}
	if lapID != 0 {
		t.Errorf("Got %d, expected 0", lapID)
	}
}
