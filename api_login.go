package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetToken refreshes the OAuth token
func (api *API) GetToken(username, password string) (*LoginJSON, error) {
	data := url.Values{}
	data.Add("grant_type", "password")
	data.Add("username", username)
	data.Add("password", password)

	url := fmt.Sprintf("https://%s%s", api.Endpoints.Users, "/oauth/token")
	log.Infof("[GetToken] POST %s", url)
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Add("User-Agent", "mylaps-streamer/1.0.0")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.SetBasicAuth(api.AppKey, api.AppSecret)

	res, err := netClient.Do(req)
	if res == nil {
		return nil, err
	}
	log.Debugf("[GetToken] HTTP%d", res.StatusCode)

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var body LoginJSON
		err = json.NewDecoder(res.Body).Decode(&body)
		if err != nil {
			log.Errorf("[GetToken] Unmarshal incorrect: %s", err)
			return nil, err
		}
		log.Infof("[GetToken] Result: length=%d expires_in=%d", len(body.AccessToken), body.ExpiresIn)
		return &body, nil
	}

	// Other statuscodes
	log.Debugf("curl -X POST %s -f -H 'Authorization: Basic %s' -d '%s'", url, basicAuth(username, password), data.Encode())
	return nil, errors.New(string(fmt.Sprintf("%s", res.Status)))
}

// GetClaims fetches the basic profile info
func (api *API) GetClaims(accessToken string) (*ProfileJSON, error) {
	url := fmt.Sprintf("https://%s%s", api.Endpoints.Users, "/auth/claims")
	req, err := http.NewRequest("POST", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("User-Agent", "mylaps-streamer/1.0.0")

	res, err := netClient.Do(req)
	if res == nil {
		return nil, err
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var body ClaimsJSON
		err = json.NewDecoder(res.Body).Decode(&body)
		defer res.Body.Close()
		if err != nil {
			log.Errorf("[GetClaims] Unmarshal incorrect: %s", err)
			return nil, err
		}
		var profile ProfileJSON
		for _, claim := range body.Claims {
			if strings.HasSuffix(claim.Type, "nameidentifier") {
				profile.UserID = claim.Value
			}
			if strings.HasSuffix(claim.Type, "name") {
				profile.NickName = claim.Value
			}
		}
		if profile.UserID == "" {
			return nil, fmt.Errorf("No usable claims found")
		}
		return &profile, nil
	}

	// Other statuscodes
	log.Debugf("curl -X POST %s -f -H 'Authorization: Bearer %s'", url, accessToken)
	return nil, errors.New(string(fmt.Sprintf("%s", res.Status)))
}

// GetProfile fetches the basic profile info
func (api *API) GetProfile(accessToken, email string) (*ProfileJSON, error) {
	data := url.Values{}
	data.Add("username", email)
	url := fmt.Sprintf("https://%s%s", api.Endpoints.Users, "/api/v2/accounts/email")
	req, err := http.NewRequest("PUT", url, strings.NewReader(data.Encode()))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("User-Agent", "mylaps-streamer/1.0.0")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := netClient.Do(req)
	if res == nil {
		return nil, err
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var body ProfileJSON
		err = json.NewDecoder(res.Body).Decode(&body)
		if err != nil {
			log.Errorf("[GetToken] Unmarshal incorrect: %s", err)
			return nil, err
		}
		return &body, nil
	}

	// Other statuscodes
	log.Debugf("curl -X PUT %s -f -H 'Authorization: Bearer %s' -d '%s'", url, accessToken, data.Encode())
	return nil, errors.New(string(fmt.Sprintf("%s", res.Status)))
}
