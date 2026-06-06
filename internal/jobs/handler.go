package jobs

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

type Service interface {
	Create(ctx context.Context, key string, payload *JobsCreateRequest) (*JobsCreateResponse, error)
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	key := c.GetHeader("Idempotency-Key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Idempotency-Key is required"})
		return
	}

	var payload *JobsCreateRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error with parsing payload as JSON."})
		return
	}

	result, err := h.service.Create(c, key, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Create error.", "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully created user.", "result": result})
}
