package database

/*
This file is responsible for connecting to the PostgreSQL database
Without a driver/library, Go cannot communicate with PostgreSQL

`pgx` handles things such as establishing PostgreSQL connections,
executing queries, scanning results, etc.
*/

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Creates a new pool (a collection of database collections) to handle requests
Once a request is received the pool gives it a connection.
*/
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		return nil, fmt.Errorf("Create Postgres pool error: %w", err)
	}

	// Check if PostgreSQL is accepting connections,
	// ctx is here if a cancellation or deadline is necessary
	// The caller of this function decides what this cancellation time is
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("Ping Postgres error: %w", err)
	}

	return pool, nil
}
