package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HelloWorldHander(w http.ResponseWriter, req *http.Request) {
	log.Debug("Entering in dummyHander")
	w.Write([]byte("Hello world\n"))
}
