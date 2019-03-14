package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

var userIndex = map[string]SavedUser{}
var userFile = "users.json"

func check(err error) {
	if err != nil {
		log.Error(err)
		panic(err)
	}
}

func getUsers(r *http.Request) (map[string]SavedUser, error) {
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		log.Debugf("Users file '%s' does not exist yet.\n", userFile)
		return nil, nil
	}
	b, err := ioutil.ReadFile(userFile)
	check(err)
	var usersList []SavedUser
	err = json.Unmarshal(b, &usersList)
	check(err)
	log.Debugf("Loaded users from file '%s'.\n", userFile)

	// From list to map
	for _, user := range usersList {
		userIndex[user.UserID] = user
	}
	return userIndex, nil
}

func storeUser(user SavedUser, r *http.Request) error {
	userIndex[user.UserID] = user

	// To list from map
	var usersList []SavedUser
	for _, user := range userIndex {
		usersList = append(usersList, user)
	}

	data, err := json.MarshalIndent(usersList, "", "  ")
	check(err)
	jsonFile, err := os.Create(userFile)
	defer jsonFile.Close()
	jsonFile.Write(data)
	jsonFile.Sync()
	return nil
}
