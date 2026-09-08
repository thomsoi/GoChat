package cache

// Retrieving messages by room is a common query and can be quite expensive
// (if no limit or a large limit is provided) hence why it is cached

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"gochat/internal/message"
)

const (
	messageCacheTTL   = 30 * time.Second
	messageCacheLimit = 100 // corresponds with limit=100 in messages/messages.go to retrieve the number of messages per room
)

func messagesKey(roomID int64) string {
	return fmt.Sprintf("room:%d:messages", roomID)
}

// Aattempts to retrieve the recent messages for a room from Redis
// The second return value tells the caller whether the cache contained
// a value (false is cache miss and true is cache hit)
func (r *Redis) GetMessagesByRoom(ctx context.Context, roomID int64) ([]message.Message, bool, error) {
	key := messagesKey(roomID)

	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// The key does not exist
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("Error when getting cached messages: %w", err)
	}

	var messages []message.Message

	if err := json.Unmarshal([]byte(value), &messages); err != nil {
		return nil, false, fmt.Errorf("Error unmarshal cached messages: %w", err)
	}

	return messages, true, nil
}

// Stores the recent messages for a room in Redis
func (r *Redis) SetMessagesByRoom(ctx context.Context, roomID int64, messages []message.Message) error {
	key := messagesKey(roomID)

	value, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("Error marshal messages for cache: %w", err)
	}

	if err := r.client.Set(
		ctx,
		key,
		value,
		messageCacheTTL,
	).Err(); err != nil {
		return fmt.Errorf("Error set cached messages: %w", err)
	}

	return nil
}

// Invalidates the cached message list for a room
func (r *Redis) DeleteMessagesByRoom(ctx context.Context, roomID int64) error {
	key := messagesKey(roomID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("Error delete cached messages: %w", err)
	}

	return nil
}
