package std_handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/linsfelipec/url-analytics/internal/handler"
	"github.com/linsfelipec/url-analytics/internal/url"
)

type URLService interface {
	Create(ctx context.Context, originalURL string) (*url.URL, error)
}

type Handler struct {
	service URLService
}

func NewHandler(s URLService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request handler.CreateURLRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON request", http.StatusBadRequest)
		return
	}

	createdURL, err := h.service.Create(r.Context(), request.URL)
	if err != nil {
		if errors.Is(err, url.ErrInvalidURL) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := handler.CreateURLResponse{
		Code:        createdURL.Code,
		ShortURL:    "http://localhost:8080/" + createdURL.Code,
		OriginalURL: createdURL.OriginalURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
