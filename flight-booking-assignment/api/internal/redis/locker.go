package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const seatLockTTL = 15 * time.Minute

type SeatLocker interface {
	LockSeat(ctx context.Context, flightID, seat string) (bool, error)
	UnlockSeat(ctx context.Context, flightID, seat string) error
}

type RedisLocker struct {
	client *goredis.Client
}

func NewRedisLocker(addr string) *RedisLocker {
	c := goredis.NewClient(&goredis.Options{Addr: addr})
	return &RedisLocker{client: c}
}

func seatKey(flightID, seat string) string {
	return fmt.Sprintf("seat:%s:%s", flightID, seat)
}

// LockSeat attempts to set a seat lock using SET NX. Returns true if the lock
// was acquired, false if the seat was already locked.
func (r *RedisLocker) LockSeat(ctx context.Context, flightID, seat string) (bool, error) {
	ok, err := r.client.SetNX(ctx, seatKey(flightID, seat), "locked", seatLockTTL).Result()
	if err != nil {
		return false, fmt.Errorf("redis set nx: %w", err)
	}
	return ok, nil
}

func (r *RedisLocker) UnlockSeat(ctx context.Context, flightID, seat string) error {
	return r.client.Del(ctx, seatKey(flightID, seat)).Err()
}
