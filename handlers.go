package main

import (
	"fmt"
	"log"
	"net/http"
)

//getVersionHanlder will return the version of the service
func getVersionHanlder(w http.ResponseWriter, r *http.Request) {
	log.Println("[/version:GET] Requested api version. " + appVersion)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, fmt.Sprintf(`{"service": "%s", "version": "%s"}`, appName, appVersion))
}

//getStatusHanlder will supply a general service status
func getStatusHanlder(w http.ResponseWriter, r *http.Request) {
	log.Println("[/:GET] Requested service status.")

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, fmt.Sprintf(`{"service": "%s", "status":"running", version": "%s"}`, appName, appVersion))
}
