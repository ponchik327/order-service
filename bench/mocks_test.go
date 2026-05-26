package bench

import (
	"context"
	"sync"

	"order_service/internal/cache"
	"order_service/internal/domain"
)

// mockRepo — простой in-memory репозиторий, без I/O.
type mockRepo struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

func newMockRepo() *mockRepo {
	return &mockRepo{orders: make(map[string]*domain.Order)}
}

func (m *mockRepo) Create(_ context.Context, o *domain.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.orders[o.OrderUID]; ok {
		return domain.ErrOrderUIDNotUnique
	}
	cp := *o
	m.orders[o.OrderUID] = &cp
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, uid string) (*domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.orders[uid]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	cp := *o
	return &cp, nil
}

// mockCache — in-memory кэш с переключаемым режимом hit/miss.
type mockCache struct {
	mu     sync.RWMutex
	data   map[string]*domain.Order
	always Mode
}

type Mode int

const (
	ModeDefault Mode = iota // ведёт себя как обычный кэш
	ModeMiss                // всегда возвращает cache miss
	ModeHit                 // всегда возвращает заранее положенный заказ
)

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]*domain.Order)}
}

func (c *mockCache) SetOrder(_ context.Context, uid string, o *domain.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := *o
	c.data[uid] = &cp
	return nil
}

func (c *mockCache) GetOrder(_ context.Context, uid string) (*domain.Order, error) {
	switch c.always {
	case ModeMiss:
		return nil, cache.ErrNotFound
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	o, ok := c.data[uid]
	if !ok {
		return nil, cache.ErrNotFound
	}
	cp := *o
	return &cp, nil
}

func (c *mockCache) Ping() error  { return nil }
func (c *mockCache) Close() error { return nil }
