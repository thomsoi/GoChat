package store

import (
	"context"
	"fmt"
	"log"

	"gochat/internal/message"
)

const messageCacheLimit = 100

func (s *Store) CreateMessage(ctx context.Context, m *message.Message) (*message.Message, error) {
	const query = `
		INSERT INTO messages (
			room_id,
			sender_id,
			content
		)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := s.db.QueryRow(
		ctx,
		query,
		m.RoomID,
		m.SenderID,
		m.Content,
	).Scan(
		&m.ID,
		&m.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Error. Create message: %w", err)
	}

	// Invalidate the old cached messages by room list since a new one has been created
	// only if no return has happened yet / postgres has not failed yet
	if err := s.cache.DeleteMessagesByRoom(ctx, m.RoomID); err != nil {
		log.Printf("Error message cache delete failed for room %d: %v", m.RoomID, err)
	}

	return m, nil
}

func (s *Store) GetMessageByID(ctx context.Context, id int64) (*message.Message, error) {
	const query = `
		SELECT
			id,
			room_id,
			sender_id,
			content,
			created_at,
			edited_at,
			deleted_at
		FROM messages
		WHERE id = $1
	`

	var m message.Message

	err := s.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&m.ID,
		&m.RoomID,
		&m.SenderID,
		&m.Content,
		&m.CreatedAt,
		&m.EditedAt,
		&m.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Error. Get message by id: %w", err)
	}

	return &m, nil
}

// TODO: Check the authenticated user is actually in the room and look into limit / query
func (s *Store) ListMessagesByRoom(ctx context.Context, roomID int64, limit int) ([]message.Message, error) {
	// TODO: This if statement is also true but is useful if client supplies the limit in the future
	if limit == messageCacheLimit {
		cachedMessages, found, err := s.cache.GetMessagesByRoom(ctx, roomID)
		if err != nil {
			// If Redis fails, continue to Postgres
			log.Printf("message cache GET failed for room %d: %v", roomID, err)
		} else if found {
			log.Printf("HELLO INSIDE LIST MESSAGES");
			return cachedMessages, nil
		}
	}
	
	const query = `
		SELECT
			id,
			room_id,
			sender_id,
			content,
			created_at,
			edited_at,
			deleted_at
		FROM messages
		WHERE room_id = $1
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := s.db.Query(ctx, query, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("Error. List messages: %w", err)
	}
	defer rows.Close()

	var messages []message.Message

	for rows.Next() {
		var m message.Message

		if err := rows.Scan(
			&m.ID,
			&m.RoomID,
			&m.SenderID,
			&m.Content,
			&m.CreatedAt,
			&m.EditedAt,
			&m.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("Error. Scan message: %w", err)
		}

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error. Iterate messages: %w", err)
	}

	// Cache the messages
	if limit == messageCacheLimit {
		if err := s.cache.SetMessagesByRoom(ctx, roomID, messages); err != nil {
			log.Printf("Error message cache SET failed for room %d: %v", roomID, err)
		}
	} else {
		log.Printf("MISMATCH");
	}

	return messages, nil
}

func (s *Store) UpdateMessage(ctx context.Context, id int64, content string) (*message.Message, error) {
	const query = `
		UPDATE messages
		SET
			content = $1,
			edited_at = NOW()
		WHERE id = $2
			AND deleted_at IS NULL
		RETURNING
			id,
			room_id,
			sender_id,
			content,
			created_at,
			edited_at,
			deleted_at
	`

	var m message.Message

	err := s.db.QueryRow(
		ctx,
		query,
		content,
		id,
	).Scan(
		&m.ID,
		&m.RoomID,
		&m.SenderID,
		&m.Content,
		&m.CreatedAt,
		&m.EditedAt,
		&m.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("update message: %w", err)
	}

	if err := s.cache.DeleteMessagesByRoom(ctx, m.RoomID); err != nil {
		log.Printf("Error message cache delete failed for room %d: %v", m.RoomID, err)
	}

	return &m, nil
}

// Soft delete so references still have the message identity.
// When getting message history, if deleted_at is null then it will be retrieved.
func (s *Store) DeleteMessage(ctx context.Context, id int64) error {
	const query = `
		UPDATE messages
		SET deleted_at = NOW()
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	_, err := s.db.Exec(
		ctx,
		query,
		id,
	)

	if err != nil {
		return fmt.Errorf("Error. Delete message: %w", err)
	}

	if err := s.cache.DeleteMessagesByRoom(ctx, id); err != nil {
		log.Printf("Error message cache delete failed for room %d: %v", id, err)
	}

	return nil
}
