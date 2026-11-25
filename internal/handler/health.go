package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

func Healthz(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	healthzStatus := StatusResponse{Status: "ok"}
	if err := json.NewEncoder(w).Encode(healthzStatus); err != nil {
		log.Printf("error encoding health response: %v", err)
	}
}
