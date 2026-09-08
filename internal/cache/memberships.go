package cache

// Determining whether a user is a member of a room is a very common query hence why the result
// will be cached

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const membershipTTL = 5 * time.Minute

func membershipKey(roomID int64, userID int64) string {
	return fmt.Sprintf("room:%d:member:%d", roomID, userID)
}

func (r *Redis) GetMembership(
	ctx context.Context,
	roomID int64,
	userID int64,
) (bool, bool, error) {
	key := membershipKey(roomID, userID)

	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, false, nil
		}

		return false, false, fmt.Errorf("Error get membership cache: %w", err)
	}

	return value == "1", true, nil
}

func (r *Redis) SetMembership(
	ctx context.Context,
	roomID int64,
	userID int64,
	isMember bool,
) error {
	key := membershipKey(roomID, userID)

	value := "0"
	if isMember {
		value = "1"
	}

	if err := r.client.Set(
		ctx,
		key,
		value,
		membershipTTL,
	).Err(); err != nil {
		return fmt.Errorf("Error set membership cache: %w", err)
	}

	return nil
}

func (r *Redis) DeleteMembership(
	ctx context.Context,
	roomID int64,
	userID int64,
) error {
	key := membershipKey(roomID, userID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("Error delete membership cache: %w", err)
	}

	return nil
}
