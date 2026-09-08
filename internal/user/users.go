package user

/*
This file represents a User. It contains the type how it is represented
in the Go application (it is a domain object). The user service has functions
which can be called, and it uses an underlying `store`. This has been abstracted
so this does not know the specific details, just the methods.
*/

import (
	"fmt"
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Go's application representation of a User
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// For registering a user, they supply a plaintext password
// but the hash is stored in the database
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Service is a struct which has some value which has the
// following methods defined for it
type Service struct {
	store Store
}

// The underlying store must have these methods. (internal/store/users)
type Store interface {
	CreateUser(
		ctx context.Context,
		user *User,
	) (*User, error)

	GetUserByID(
		ctx context.Context,
		id int64,
	) (*User, error)

	GetUserByUsername(
		ctx context.Context,
		username string,
	) (*User, error)
}

// Create a Service, given a Store. This decouples Store and Service
func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// TODO: Authorisation and business logic checks will go here

// Creates a new user
func (s *Service) CreateUser(ctx context.Context, i CreateUserRequest) (*User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(i.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("Error. Hash password failed: %w", err);
	}

	u := &User{
		Username:     i.Username,
		Email:	      i.Email,
		PasswordHash: string(passwordHash),
	}

	return s.store.CreateUser(ctx, u)
}

// Obtains a user by their ID
func (s *Service) GetUserByID(ctx context.Context, id int64) (*User, error) {
	return s.store.GetUserByID(ctx, id)
}

// Obtains a user by their username
func (s *Service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.store.GetUserByUsername(ctx, username)
}
