package gin_handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linsfelipec/url-analytics/internal/handler"
	"github.com/linsfelipec/url-analytics/internal/url"
)

type URLService interface {
	Create(ctx context.Context, originalURL string) (*url.URL, error)
	GetByCode(ctx context.Context, code string) (*url.URL, error)
}

type Handler struct {
	service URLService
}

func NewHandler(s URLService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Create(ctx *gin.Context) {
	var request handler.CreateURLRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON request",
		})
		return
	}

	createdURL, err := h.service.Create(ctx.Request.Context(), request.URL)
	if err != nil {
		if errors.Is(err, url.ErrInvalidURL) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid URL",
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal error",
			})
		}
		return
	}

	response := handler.CreateURLResponse{
		Code:        createdURL.Code,
		ShortURL:    "http://localhost:8080/" + createdURL.Code,
		OriginalURL: createdURL.OriginalURL,
	}
	ctx.JSON(http.StatusCreated, response)
}

func (h *Handler) GetByCode(ctx *gin.Context) {
	code := ctx.Param("code")

	foundURL, err := h.service.GetByCode(ctx.Request.Context(), code)
	if err != nil {
		if errors.Is(err, url.ErrURLNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "URL not found",
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal error",
			})
		}
		return
	}

	response := handler.GetURLResponse{
		Code:        foundURL.Code,
		OriginalURL: foundURL.OriginalURL,
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) Redirect(ctx *gin.Context) {
	code := ctx.Param("code")
	foundURL, err := h.service.GetByCode(ctx.Request.Context(), code)
	if err != nil {
		if errors.Is(err, url.ErrURLNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "URL not found",
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal error",
			})
		}
		return
	}

	ctx.Redirect(http.StatusFound, foundURL.OriginalURL)
}
