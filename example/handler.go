package example

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

type Service interface {
	CreateExample(ctx context.Context, req *CreateExampleRequest) (*Example, error)
	GetExample(ctx context.Context, id uuid.UUID) (*Example, error)
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateExample(c *gin.Context) {
	var example *CreateExampleRequest
	if err := c.ShouldBindJSON(&example); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error with parsing payload as JSON.")})
		return
	}
	result, err := h.service.CreateExample(c, example)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error with parsing payload as JSON.")})
	}

	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully created user.", "result": result})
}

func (h *Handler) GetExample(c *gin.Context) {
	resId := c.Param("id")
	// var example Example
	// if err := c.ShouldBindJSON(&example); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error with parsing payload as JSON.")})
	// 	return
	// }

	id, err := uuid.Parse(resId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error with parsing payload as JSON.")})
		return
	}

	result, err := h.service.GetExample(c, id)

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully get user.", "result": result})
}
