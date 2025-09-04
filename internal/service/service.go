package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"order_service/internal/cache"
	"order_service/internal/config"
	"order_service/internal/domain"
	"order_service/internal/repository"
)

// OrderService определяет интерфейс для бизнес-логики заказов.
type OrderService interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByID(ctx context.Context, orderUID string) (*domain.Order, error)
	HandleOrder(ctx context.Context, message []byte) error
}

// orderService — реализация OrderService.
type orderService struct {
	repo   repository.OrderRepository
	config *config.Config
	cache  cache.Cache
}

// NewOrderService создает новый экземпляр orderService.
func NewOrderService(repo repository.OrderRepository, config *config.Config, cache cache.Cache) OrderService {
	return &orderService{
		repo:   repo,
		config: config,
		cache:  cache,
	}
}

// CreateOrder создает заказ с базовой валидацией.
func (s *orderService) CreateOrder(ctx context.Context, order *domain.Order) error {
	if order.OrderUID == "" {
		return domain.ErrOrderUIDEmpty
	}
	return s.repo.Create(ctx, order)
}

// GetOrderByID получает заказ по order_uid.
func (s *orderService) GetOrderByID(ctx context.Context, orderUID string) (*domain.Order, error) {
	if orderUID == "" {
		return nil, domain.ErrOrderUIDEmpty
	}

	// Сходили в кэш
	order, err := s.cache.GetOrder(ctx, orderUID)
	if err == nil {
		return order, nil
	}
	cacheErr := err

	// Кэш вернул ошибку, идём в БД
	order, err = s.repo.GetByID(ctx, orderUID)
	if err != nil {
		return nil, err
	}

	// Кладём заказ в кэш
	if cacheErr == cache.ErrNotFound {
		err = s.cache.SetOrder(ctx, orderUID, order)
		if err != nil {
			log.Printf("Ошибка сохранения в кэш: %v\n", err)
		}
	} else { // Ошибка кэша не связана с отсутствием заказа
		log.Printf("Ошибка кэша: %v, fallback на БД\n", err)
	}

	return order, nil
}

// HandleOrder парсит и сохраняет заказ
func (s *orderService) HandleOrder(ctx context.Context, message []byte) error {
	var order domain.Order
	if err := json.Unmarshal(message, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	if err := s.CreateOrder(ctx, &order); err != nil {
		return fmt.Errorf("failed create order %s: %w", order.OrderUID, err)
	}

	log.Printf("Successfully processed order: %s", order.OrderUID)
	return nil
}
