package auth

import (
	"context"
	"net/http"
	"strings"

	commonconstants "github.com/gibson7780/go-project/common/constants"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

type Service interface {
	Signup(ctx context.Context, req *SignupRequest) (*AuthResponse, error)
	Signin(ctx context.Context, req *SigninRequest) (*AuthResponse, error)
	Signout(ctx context.Context, token string) error
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Signup(c *gin.Context) {
	var req *SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error with parsing payload as JSON."})
		return
	}

	result, err := h.service.Signup(c, req)
	if err == commonconstants.ErrDuplicateResource {
		c.JSON(http.StatusConflict, gin.H{"statusCode": http.StatusConflict, "message": "Email already registered."})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Signup error.", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully signed up.", "result": result})
}

func (h *Handler) Signin(c *gin.Context) {
	var req *SigninRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error with parsing payload as JSON."})
		return
	}

	result, err := h.service.Signin(c, req)
	if err == commonconstants.ErrUnauthorized {
		c.JSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Invalid credentials."})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Signin error.", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully signed in.", "result": result})
}

func (h *Handler) Signout(c *gin.Context) {
	token := extractBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Missing token."})
		return
	}

	if err := h.service.Signout(c, token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Signout error.", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully signed out."})
}

func extractBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
