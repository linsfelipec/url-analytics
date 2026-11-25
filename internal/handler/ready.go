package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ReadyRepository interface {
	PingContext(context.Context) error
}

func Readyz(repo ReadyRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repoCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		response := StatusResponse{Status: "ok"}
		responseStatus := http.StatusOK
		if err := repo.PingContext(repoCtx); err != nil {
			response.Status = "unavailable"
			responseStatus = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatus)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("error encoding ready response: %v", err)
		}
	}
}
