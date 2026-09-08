package store

import (
	"context"
	"fmt"

	"gochat/internal/user"
)

/*
Create a user in the database. Takes the context, to allow the database operation
to be cancelled or to respect a timeout, as well as the User which will contain
the information to be persisted to PostgreSQL.
We pass in a pointer so that we can give it an ID and a time created using Scan.
*/
func (s *Store) CreateUser(ctx context.Context, u *user.User) (*user.User, error) {
	const query = `
		INSERT INTO users (
			username,
			email,
			password_hash
		)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	// Returning the ID and created at, so we ened QueryRow and not exec.
	err := s.db.QueryRow(
		ctx,
		query,
		u.Username,
		u.Email,
		u.PasswordHash,
	).Scan(
		&u.ID,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Error. Create user: %w", err)
	}

	// No need to return the new User object since a pointer to it was passed instead.
	return u, nil
}

// Obtain a user from the database and return a pointer to it.
func (s *Store) GetUserByID(ctx context.Context, id int64) (*user.User, error) {
	const query = `
		SELECT
			id,
			username,
			email,
			password_hash,
			created_at
		FROM users
		WHERE id = $1
	`
	var u user.User

	err := s.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Error. Get user by ID: %w", err)
	}

	return &u, nil
}

// Obtain a user by their username
func (s *Store) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	const query = `
		SELECT
			id,
			username,
			email,
			password_hash,
			created_at
		FROM users
		WHERE username = $1
	`
	var u user.User

	err := s.db.QueryRow(
		ctx,
		query,
		username,
	).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Error. Get user by username: %w", err)
	}

	return &u, nil
}

// Obtain a user by their email address
// func (s *Store) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
// 	const query = `
// 		SELECT
// 			id,
// 			username,
// 			email,
// 			password_hash,
// 			created_at
// 		FROM users
// 		WHERE email = $1
// 	`
// 	var u user.User

// 	err := s.db.QueryRow(
// 		ctx,
// 		query,
// 		email,
// 	).Scan(
// 		&u.ID,
// 		&u.Username,
// 		&u.Email,
// 		&u.PasswordHash,
// 		&u.CreatedAt,
// 	)

// 	if err != nil {
// 		return nil, fmt.Errorf("Error. Get user by email: %w", err)
// 	}

// 	return &u, nil
// }
