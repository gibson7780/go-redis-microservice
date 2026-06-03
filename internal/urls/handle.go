package urls

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

type Service interface {
	CreateUrl(ctx context.Context, req *CreateUrlRequest, idemKey string) (*CreateUrlResponse, error)
	GetUrl(ctx context.Context, code string) (*GetUrlResponse, error)
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateUrl(c *gin.Context) {
	idemKey := c.GetHeader("Idempotency-Key")
	// fmt.Println("idemKey", idemKey)
	var UrlRequest *CreateUrlRequest
	if err := c.ShouldBindJSON(&UrlRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error with parsing payload as JSON."})
		return
	}
	result, err := h.service.CreateUrl(c, UrlRequest, idemKey)
	slog.Error("error:", "err", err)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Create error.", "error": err})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully created user.", "result": result})
}

func (h *Handler) GetUrl(c *gin.Context) {
	code := c.Param("code")
	slog.Info("code", "code", code)
	result, err := h.service.GetUrl(c, code)
	slog.Info("success", "result", result)
	slog.Info("success", "result", err)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": err.Error()})
		return
	}

	// c.JSON(http.StatusPermanentRedirect, gin.H{"statusCode": http.StatusOK, "message": "Successfully get url.", "result": result})
	c.Redirect(http.StatusFound, result.OriginUrl)
}
