package store

import (
	"context"
	"fmt"

	"gochat/internal/room"
)

// NOTE: This uses the Store struct that is created in `store/store.go`.
func (s *Store) CreateRoom(ctx context.Context, r *room.Room) (*room.Room, error) {
	const query = `
		INSERT INTO rooms (
			room_name,
			created_by
		)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := s.db.QueryRow(
		ctx,
		query,
		r.Name,
		r.CreatedBy,
	).Scan(
		&r.ID,
		&r.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Error. Create room: %w", err)
	}

	return r, nil
}

func (s *Store) GetRoomByID(ctx context.Context, id int64) (*room.Room, error) {
	const query = `
		SELECT
			id,
			room_name,
			created_by,
			created_at
		FROM rooms
		WHERE id = $1
	`

	var r room.Room

	err := s.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&r.ID,
		&r.Name,
		&r.CreatedBy,
		&r.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Error. Get room by id: %w", err)
	}

	return &r, nil
}

func (s *Store) ListRooms(ctx context.Context) ([]room.Room, error) {
	const query = `
		SELECT
			id,
			room_name,
			created_by,
			created_at
		FROM rooms
		ORDER BY id
	`
	// QueryRow is querying one row, this will return potentially many rows.
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Error. List rooms: %w", err)
	}
	defer rows.Close()

	var rooms []room.Room

	for rows.Next() {
		var r room.Room

		if err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.CreatedBy,
			&r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("Error. Scan room: %w", err)
		}

		rooms = append(rooms, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error. Iterate rooms: %w", err)
	}

	return rooms, nil
}

func (s *Store) ListRoomsByUser(ctx context.Context, userID int64) ([]room.Room, error) {
	const query = `
        SELECT
            r.id,
            r.room_name,
            r.created_by,
            r.created_at
        FROM rooms r
        INNER JOIN room_members rm
            ON rm.room_id = r.id
        WHERE rm.user_id = $1
        ORDER BY r.id
    `

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list rooms by user: %w", err)
	}
	defer rows.Close()

	rooms := make([]room.Room, 0)

	for rows.Next() {
		var r room.Room

		if err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.CreatedBy,
			&r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}

		rooms = append(rooms, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}

	return rooms, nil
}

func (s *Store) ListMembersByRoom(ctx context.Context, roomID int64) ([]room.RoomMember, error) {
	const query = `
		SELECT
			rm.user_id,
			u.username,
			rm.role,
			rm.joined_at
		FROM room_members rm
		INNER JOIN users u
			ON u.id = rm.user_id
		WHERE rm.room_id = $1
		ORDER BY rm.joined_at
	`

	rows, err := s.db.Query(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room members: %w", err)
	}
	defer rows.Close()

	members := make([]room.RoomMember, 0)

	for rows.Next() {
		var member room.RoomMember

		if err := rows.Scan(
			&member.UserID,
			&member.Username,
			&member.Role,
			&member.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room member: %w", err)
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room members: %w", err)
	}

	return members, nil
}

func (s *Store) DeleteRoom(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM rooms
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("Error. Delete room: %w", err)
	}

	return nil
}
