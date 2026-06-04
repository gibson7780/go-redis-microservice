package auth

import (
	"context"
	"net/http"

	commonhelpers "github.com/gibson7780/go-project/common/utils"
	"github.com/gin-gonic/gin"
)

// context keys for values set by RequireAuth
const (
	ContextUserID = "userID"
	ContextEmail  = "email"
)

type Verifier interface {
	VerifyToken(ctx context.Context, token string) (*commonhelpers.JWTClaims, error)
}

// RequireAuth rejects requests without a valid Bearer access token and stores
// the authenticated user's claims on the gin context for downstream handlers.
func RequireAuth(verifier Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Missing or malformed token."})
			return
		}

		claims, err := verifier.VerifyToken(c, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Invalid or expired token."})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Next()
	}
}
