package main

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func start(httpPort uint) {
	// Start server
	log.Infof("Listening on port %d", httpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil))
}
