package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultRedisChannel = "opus:ws:broadcast"

// RedisAdapter implements Adapter using Redis pub/sub with an in-memory fallback.
type RedisAdapter struct {
	mu          sync.RWMutex
	client      *redis.Client
	pubsub      *redis.PubSub
	fallback    *InMemoryAdapter
	logger      *slog.Logger
	channel     string
	useFallback bool
	closed      bool
}

// NewRedisAdapter creates a Redis pub/sub adapter with automatic in-memory fallback.
func NewRedisAdapter(redisURL string, logger *slog.Logger) *RedisAdapter {
	if logger == nil {
		logger = slog.Default()
	}
	adapter := &RedisAdapter{
		fallback:    NewInMemoryAdapter(),
		logger:      logger,
		channel:     defaultRedisChannel,
		useFallback: true,
	}
	if redisURL == "" {
		return adapter
	}
	client, err := initRedisClient(redisURL, logger)
	if err != nil {
		return adapter
	}
	adapter.client = client
	adapter.useFallback = false
	return adapter
}

func initRedisClient(redisURL string, logger *slog.Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Warn("invalid redis url, using in-memory fallback", slog.Any("error", err))
		return nil, err
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("redis ping failed, using in-memory fallback", slog.Any("error", err))
		client.Close()
		return nil, err
	}
	return client, nil
}

// NewAdapter constructs a broadcast adapter based on configuration.
func NewAdapter(redisURL string, logger *slog.Logger) Adapter {
	if redisURL == "" {
		return NewInMemoryAdapter()
	}
	return NewRedisAdapter(redisURL, logger)
}

// Publish dispatches an envelope through Redis, or in-memory fallback on failure.
func (a *RedisAdapter) Publish(ctx context.Context, env Envelope) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.closed {
		return errors.New("adapter is closed")
	}
	if a.useFallback || a.client == nil {
		return a.fallback.Publish(ctx, env)
	}
	return a.publishRedis(ctx, env)
}

func (a *RedisAdapter) publishRedis(ctx context.Context, env Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if err := a.client.Publish(ctx, a.channel, data).Err(); err != nil {
		a.logger.Warn("redis publish failed, using in-memory fallback", slog.Any("error", err))
		return a.fallback.Publish(ctx, env)
	}
	return nil
}

// Subscribe listens for broadcast envelopes from Redis or the in-memory fallback.
func (a *RedisAdapter) Subscribe(ctx context.Context, handler Handler) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return errors.New("adapter is closed")
	}
	if err := a.fallback.Subscribe(ctx, handler); err != nil {
		return err
	}
	if a.useFallback || a.client == nil {
		return nil
	}
	pubsub := a.client.Subscribe(ctx, a.channel)
	a.pubsub = pubsub
	go a.listen(pubsub, handler)
	return nil
}

func (a *RedisAdapter) listen(pubsub *redis.PubSub, handler Handler) {
	ch := pubsub.Channel()
	for msg := range ch {
		var env Envelope
		if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
			a.logger.Warn("failed to decode redis broadcast message", slog.Any("error", err))
			continue
		}
		handler(env)
	}
}

// Close terminates Redis connections and releases resources.
func (a *RedisAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closed = true
	var errs []error
	if a.pubsub != nil {
		errs = append(errs, a.pubsub.Close())
	}
	if a.client != nil {
		errs = append(errs, a.client.Close())
	}
	if a.fallback != nil {
		errs = append(errs, a.fallback.Close())
	}
	return errors.Join(errs...)
}

// UsingFallback reports whether the adapter is currently using in-memory fallback.
func (a *RedisAdapter) UsingFallback() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.useFallback
}
