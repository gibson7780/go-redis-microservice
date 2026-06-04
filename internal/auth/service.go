package auth

import (
	"context"
	"fmt"
	"time"

	commonconstants "github.com/gibson7780/go-project/common/constants"
	commonhelpers "github.com/gibson7780/go-project/common/utils"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
)

type service struct {
	repo        Repository
	redisClient redis.UniversalClient
	jwtSecret   string
}

type Repository interface {
	CreateUser(payload *User) (*User, error)
	GetUserByEmail(email string) (*User, error)
}

func NewService(repo Repository, redisClient redis.UniversalClient, jwtSecret string) *service {
	return &service{
		repo:        repo,
		redisClient: redisClient,
		jwtSecret:   jwtSecret,
	}
}

func (s *service) Signup(ctx context.Context, req *SignupRequest) (*AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(&User{
		Email:    req.Email,
		Password: string(hashed),
	})
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user)
}

func (s *service) Signin(ctx context.Context, req *SigninRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(req.Email)
	if err == commonconstants.ErrNotFound {
		return nil, commonconstants.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, commonconstants.ErrUnauthorized
	}

	return s.issueTokens(ctx, user)
}

func (s *service) Signout(ctx context.Context, tokenString string) error {
	claims, err := commonhelpers.ParseToken(tokenString, s.jwtSecret)
	if err != nil {
		return commonconstants.ErrUnauthorized
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	key := blacklistKey(tokenString)
	return s.redisClient.Set(ctx, key, "1", ttl).Err()
}

func (s *service) VerifyToken(ctx context.Context, tokenString string) (*commonhelpers.JWTClaims, error) {
	claims, err := commonhelpers.ParseToken(tokenString, s.jwtSecret)
	if err != nil {
		return nil, commonconstants.ErrUnauthorized
	}

	// only access tokens may be used to authenticate API requests
	if claims.Type != commonconstants.Access {
		return nil, commonconstants.ErrUnauthorized
	}

	return claims, nil
}

func (s *service) issueTokens(ctx context.Context, user *User) (*AuthResponse, error) {
	access, err := commonhelpers.GenerateToken(user.ID, user.Email, commonconstants.Access, s.jwtSecret, accessTTL)
	if err != nil {
		return nil, err
	}

	refresh, err := commonhelpers.GenerateToken(user.ID, user.Email, commonconstants.Refresh, s.jwtSecret, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		ID:           user.ID,
		Email:        user.Email,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func blacklistKey(token string) string {
	return fmt.Sprintf("auth:blacklist:%s", token)
}
