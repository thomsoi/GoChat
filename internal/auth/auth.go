package auth

import (
	"context"
	"errors"
	"fmt"

	"gochat/internal/user"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("Invalid credentials")

// AuthService does not need to know the entire user store, so only expose
// what is needed
type UserStore interface {
	GetUserByUsername(ctx context.Context, username string) (*user.User, error)
}

// DTOs
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// The authentication server must be able to call GetUserByUsername
type Service struct {
	userStore UserStore
	jwt       *JWTManager
}

func NewService(users UserStore, jwtManager *JWTManager) *Service {
	return &Service{userStore: users, jwt: jwtManager}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	u, err := s.userStore.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(u.ID)
	if err != nil {
		return nil, fmt.Errorf("Error when generating token: %w", err)
	}

	return &LoginResponse{Token: token}, nil
}
