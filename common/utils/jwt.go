package commonhelpers

import (
	"time"

	commonconstants "github.com/gibson7780/go-project/common/constants"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID uuid.UUID                 `json:"user_id"`
	Email  string                    `json:"email"`
	Type   commonconstants.TokenType `json:"type"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, email string, tokenType commonconstants.TokenType, secret string, ttl time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString string, secret string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, commonconstants.ErrUnauthorized
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, commonconstants.ErrUnauthorized
	}

	return claims, nil
}
