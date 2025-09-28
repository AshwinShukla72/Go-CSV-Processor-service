package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/AshwinShukla72/csv-processor/services"
	"github.com/redis/go-redis/v9"
)

type (
	JobQueue interface {
		Queue(id string) error
		Worker(store Store, parser services.Parser)
		State(id string) (string, error) // Managing job States
	}
	redisQueue struct {
		client *redis.Client
		key    string
	}
)

func NewRedisQueue(addr string) JobQueue {
	client := redis.NewClient(&redis.Options{Addr: addr})
	return &redisQueue{client: client, key: "jobs"}
}

func (r *redisQueue) Queue(id string) error {
	ctx := context.Background()
	r.client.Set(ctx, id+":state", "processing", 0)
	return r.client.LPush(ctx, r.key, id).Err()
}
func (r *redisQueue) State(id string) (string, error) {
	ctx := context.Background()
	state, err := r.client.Get(ctx, id+":state").Result()
	if err == redis.Nil {
		return "", errors.New("not found")
	}
	return state, err
}

func (r *redisQueue) Worker(store Store, parser services.Parser) {
	ctx := context.Background()
	for {
		id, err := r.client.RPop(ctx, r.key).Result()
		if err == redis.Nil {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		// Load original file using store interface
		raw, err := store.LoadOriginal(id)
		if err != nil {
			// Log error and mark job as failed
			r.client.Set(ctx, id+":state", "failed", 0)
			continue
		}

		// Process the CSV data
		processed, err := parser.Parse(raw)
		if err != nil {
			// Mark job as failed if parsing fails
			r.client.Set(ctx, id+":state", "failed", 0)
			continue
		}

		// Save processed data
		err = store.SaveProcessed(id, processed)
		if err != nil {
			// Mark job as failed if saving fails
			r.client.Set(ctx, id+":state", "failed", 0)
			continue
		}

		// Mark job as completed
		r.client.Set(ctx, id+":state", "done", 0)
	}
}
