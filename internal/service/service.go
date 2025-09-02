package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
}

// NewOrderService создает новый экземпляр orderService.
func NewOrderService(repo repository.OrderRepository, config *config.Config) OrderService {
	return &orderService{
		repo:   repo,
		config: config,
	}
}

// CreateOrder создает заказ с базовой валидацией.
func (s *orderService) CreateOrder(ctx context.Context, order *domain.Order) error {
	if order.OrderUID == "" {
		return errors.New("order_uid cannot be empty")
	}
	return s.repo.Create(ctx, order)
}

// GetOrderByID получает заказ по order_uid.
func (s *orderService) GetOrderByID(ctx context.Context, orderUID string) (*domain.Order, error) {
	if orderUID == "" {
		return nil, errors.New("order_uid cannot be empty")
	}
	return s.repo.GetByID(ctx, orderUID)
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
