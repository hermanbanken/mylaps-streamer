package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func tryFlush(w http.ResponseWriter) {
	if flusher, hasFlush := w.(http.Flusher); hasFlush {
		flusher.Flush()
	}
}

// @see: https://play.golang.org/p/BNIz1iY5kq
func tryClose(w http.ResponseWriter) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return // The rw can't be hijacked, return early.
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		log.Error(err)
	}
	// Close the hijacked raw tcp connection.
	if err := conn.Close(); err != nil {
		log.Error(err)
	}
}

func getCookieOrElse(r *http.Request, key string, orElse string) string {
	if c, err := r.Cookie(key); r != nil && err == nil {
		return c.Value
	}
	return orElse
}

func getAuth(w http.ResponseWriter, r *http.Request) (accessToken string, userID string, hasToken bool) {
	accessToken = getCookieOrElse(r, "accesstoken", r.URL.Query().Get("accesstoken"))
	userID = getCookieOrElse(r, "userid", r.URL.Query().Get("userid"))
	if accessToken == "" {
		http.Redirect(w, r, "/", 307)
		hasToken = false
	} else {
		hasToken = true
	}
	return
}

func reverse(a []Activity) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}

func parseEventID(id string) (userID string, activityID uint64, lap uint) {
	subs := strings.Split(id, ":")
	if len(subs) == 0 {
		return "", 0, 0
	}
	userID = subs[0]
	if len(subs) > 1 {
		activityID, _ = strconv.ParseUint(subs[1], 10, 64)
	}
	if len(subs) > 2 {
		var lap64 uint64
		lap64, _ = strconv.ParseUint(subs[2], 10, 64)
		lap = uint(lap64)
	}
	return
}

func parseDuration(durationText string) (time.Duration, error) {
	asDuration := strings.Replace(strings.Replace(durationText, ":", "m", 1), ".", "s", 1) + "ms"
	return time.ParseDuration(asDuration)
}

func digits(durationText string) string {
	var digits = "___"
	if duration, err := parseDuration(durationText); err == nil {
		if duration > 20*time.Second && duration < 50*time.Second {
			dotPos := strings.Index(durationText, ".")
			digits = durationText[dotPos-2:dotPos] + durationText[dotPos+1:dotPos+2]
		} else {
			log.Infof("Non 20-50s lap duration: %s", durationText)
		}
	} else {
		log.Infof("Non parsable lap duration: %s", durationText)
	}
	return digits
}

func durationToDigits(duration time.Duration) string {
	var digits = "___"
	if duration > 20*time.Second && duration < 50*time.Second {
		s := strconv.Itoa(int(duration.Seconds()))
		ms := strconv.Itoa(int(duration.Seconds()*1000) % 1000)
		return s + leftPad(ms, "0", 3-len(ms))
	} else {
		log.Infof("Non 20-50s lap duration: %s", duration)
	}
	return digits
}

func leftPad(s string, padStr string, pLen int) string {
	return strings.Repeat(padStr, pLen) + s
}
