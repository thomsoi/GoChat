package store

import (
	"gochat/internal/cache"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Responsible for database operations, throughout the other
// files in `internal/store`, this Store struct will have all
// of the methods
// It is essentially a database access object that owns the shared
// connection pool
// Even though there is one Store, each service defines an interface
// for the store to have, and it will only contain methods that it needs to use
type Store struct {
	db *pgxpool.Pool
	cache *cache.Redis
}

// Database infrastructure is already defined in
// internal/database/database.go
// so if we wish to change it then this stays the same
func New(db *pgxpool.Pool, cache *cache.Redis) *Store {
	return &Store{db: db, cache: cache}
}
