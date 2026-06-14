package jobs

import (
	"context"
	"net/http"

	"github.com/gibson7780/go-project/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, key string, payload *JobsCreateRequest) (*JobsCreateResponse, error)
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	key := c.GetHeader("Idempotency-Key")
	v, exists := c.Get(auth.UserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Unauthorized."})
		return
	}
	userID, ok := v.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "Invalid user context."})
		return
	}

	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Idempotency-Key is required"})
		return
	}

	var payload *JobsCreateRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error with parsing payload as JSON."})
		return
	}

	result, err := h.service.Create(c, userID, key, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Create error.", "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully created user.", "result": result})
}
