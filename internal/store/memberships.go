package store

import (
	"context"
	"fmt"
	"log"
)

func (s *Store) AddMember(
	ctx context.Context,
	roomID int64,
	userID int64,
	role string,
) error {
	const query = `
		INSERT INTO room_members (
			room_id,
			user_id,
			role
		)
		VALUES ($1, $2, $3)
	`

	_, err := s.db.Exec(
		ctx,
		query,
		roomID,
		userID,
		role,
	)

	if err != nil {
		return fmt.Errorf("Error. Add room member: %w", err)
	}

	// Invalidate any old cached "not a member" result
    if err := s.cache.DeleteMembership(ctx, roomID, userID); err != nil {
        log.Printf("Error delete membership cache: %v", err)
    }

	return nil
}

func (s *Store) RemoveMember(
	ctx context.Context,
	roomID int64,
	userID int64,
) error {
	const query = `
		DELETE FROM room_members
		WHERE room_id = $1
		  AND user_id = $2
	`

	_, err := s.db.Exec(
		ctx,
		query,
		roomID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("Error. Remove room member: %w", err)
	}

	// The database changed, so invalidate the cached membership result
    if err := s.cache.DeleteMembership(ctx, roomID, userID); err != nil {
        log.Printf("Error delete membership cache: %v", err)
    }

	return nil
}

func (s *Store) IsMember(
	ctx context.Context,
	roomID int64,
	userID int64,
) (bool, error) {
	
	// Try Redis first to see if it is cached
	isMember, hit, err := s.cache.GetMembership(ctx, roomID, userID)
	if err != nil {
		// Cache failure, try Postgres instead later
		log.Printf("Cache failure")
	} else if hit {
		// Cache hit, return the value
		return isMember, nil
	}

	// Cache miss, try Postgres
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM room_members
			WHERE room_id = $1
			  AND user_id = $2
		)
	`

	var exists bool

	err2 := s.db.QueryRow(
		ctx,
		query,
		roomID,
		userID,
	).Scan(&exists)

	if err2 != nil {
		return false, fmt.Errorf("Error. Check room membership: %w", err2)
	}

	// Before we return the Postgres result, populate the redis cache for next time
	if err := s.cache.SetMembership(ctx, roomID, userID, exists); err != nil {
		// Cache failed
	}

	// Postgres result
	return exists, nil
}
