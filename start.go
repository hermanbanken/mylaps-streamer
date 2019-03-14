// +build !darwin

package main

import "google.golang.org/appengine"

func start(httpPort uint) {
	appengine.Main()
}
