package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func httpError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("error: %v", err)
	httpError(w, http.StatusInternalServerError)
}

func jsonResponse(w http.ResponseWriter, value any) {
	jsonResponseStatus(w, http.StatusOK, value)
}

func jsonResponseStatus(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		serverError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
