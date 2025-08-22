package service

import (
	"context"
	"errors"
	"order_service/internal/domain"
	"order_service/internal/repository"
)

// OrderService определяет интерфейс для бизнес-логики заказов.
type OrderService interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByID(ctx context.Context, orderUID string) (*domain.Order, error)
}

// orderService — реализация OrderService.
type orderService struct {
	repo repository.OrderRepository
}

// NewOrderService создает новый экземпляр orderService.
func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{repo: repo}
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
