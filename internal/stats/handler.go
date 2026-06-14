package stats

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	service Service
}

type Service interface {
	CreateStat(ctx context.Context, tx *sqlx.Tx, req *CreateStatRequest) error
	GetStat(ctx context.Context, id uuid.UUID) (*GetStatResponse, error)
	BatchStats(data map[string]int64) error
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// func (h *Handler) CreateStat(c *gin.Context) {
// 	var example *CreateStatRequest
// 	if err := c.ShouldBindJSON(&example); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error with parsing payload as JSON."})
// 		return
// 	}
// 	err := h.service.CreateStat(c, example)
// 	if err != nil {

// 		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error with parsing payload as JSON."})
// 	}

// 	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully created user."})
// }

func (h *Handler) GetStat(c *gin.Context) {
	resId := c.Param("code")
	// var example Example
	// if err := c.ShouldBindJSON(&example); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error with parsing payload as JSON.")})
	// 	return
	// }

	id, err := uuid.Parse(resId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error with parsing payload as JSON."})
		return
	}

	result, err := h.service.GetStat(c, id)

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully get user.", "result": result})
}
