// +build !darwin

package main

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

import "google.golang.org/appengine/datastore"
import "google.golang.org/appengine"

// https://cloud.google.com/appengine/docs/standard/go/building-app/storing-data
func getUsers(r *http.Request) (map[string]SavedUser, error) {
	ctx := appengine.NewContext(r)
	q := datastore.NewQuery("User").Limit(20)
	var users []SavedUser
	if _, err := q.GetAll(ctx, &users); err != nil {
		log.Errorf("Getting users: %v", err)
		return nil, err
	}
	var result = map[string]SavedUser{}
	for _, user := range users {
		result[user.UserID] = user
	}
	return result, nil
}

func storeUser(user SavedUser, r *http.Request) error {
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "User", user.UserID, 0, nil)
	if _, err := datastore.Put(ctx, key, &user); err != nil {
		log.Errorf("datastore.Put: %v", err)
		return fmt.Errorf("Couldn't add new user. %v", err)
	}
	return nil
}
