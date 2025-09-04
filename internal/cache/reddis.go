package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"order_service/internal/config"
	"order_service/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("cache miss")

type Cache interface {
	SetOrder(ctx context.Context, orderUID string, order *domain.Order) error
	GetOrder(ctx context.Context, orderUID string) (*domain.Order, error)
	Ping() error
	Close() error
}

type cache struct {
	rc  *redis.Client
	ttl time.Duration
}

func NewCache(config *config.Config) Cache {
	return &cache{
		rc: redis.NewClient(&redis.Options{
			Addr: config.Cache.Adress,
		}),
		ttl: config.Cache.Ttl,
	}
}

func (c *cache) SetOrder(ctx context.Context, orderUID string, order *domain.Order) error {
	orderBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("ошибка сериализации заказа orderUID=%s, err=%v", orderUID, err)
	}

	return c.rc.Set(ctx, orderUID, orderBytes, c.ttl).Err()
}

func (c *cache) GetOrder(ctx context.Context, orderUID string) (*domain.Order, error) {
	orderJSON, err := c.rc.Get(ctx, orderUID).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var order domain.Order
	err = json.Unmarshal([]byte(orderJSON), &order)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации заказа orderUID=%s, err=%v", orderUID, err)
	}

	return &order, nil
}

func (c *cache) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := c.rc.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println("Successfully connected to Redis")
	return nil
}

func (c *cache) Close() error {
	if err := c.rc.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}
	log.Println("Redis client closed")
	return nil
}
