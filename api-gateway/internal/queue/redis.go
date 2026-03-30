package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Key: sorted set where score = scheduled unix timestamp, member = email ID
const scheduledKey = "outreach:email:scheduled"

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(redisURL string) (*RedisQueue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisQueue{client: client}, nil
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}

// Enqueue adds an email to the schedule at the given send time
func (q *RedisQueue) Enqueue(ctx context.Context, emailID string, sendAt time.Time) error {
	return q.client.ZAdd(ctx, scheduledKey, redis.Z{
		Score:  float64(sendAt.Unix()),
		Member: emailID,
	}).Err()
}

// Cancel removes an email from the schedule. Returns true if it was actually removed.
func (q *RedisQueue) Cancel(ctx context.Context, emailID string) (bool, error) {
	removed, err := q.client.ZRem(ctx, scheduledKey, emailID).Result()
	if err != nil {
		return false, err
	}
	return removed > 0, nil
}

// ClaimDue atomically fetches and removes one email whose scheduled time <= now.
// Returns ("", nil) if nothing is due.
func (q *RedisQueue) ClaimDue(ctx context.Context) (string, error) {
	nowScore := fmt.Sprintf("%d", time.Now().Unix())

	// ZPOPMIN-style: get the lowest-scored entry, but only if score <= now
	results, err := q.client.ZRangeByScore(ctx, scheduledKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    nowScore,
		Offset: 0,
		Count:  1,
	}).Result()
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	emailID := results[0]

	// Atomically remove — if ZREM returns 1, we claimed it (no other worker got it)
	removed, err := q.client.ZRem(ctx, scheduledKey, emailID).Result()
	if err != nil {
		return "", err
	}
	if removed == 0 {
		return "", nil
	}

	return emailID, nil
}
