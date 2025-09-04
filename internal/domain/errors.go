package domain

import "errors"

// Ошибки бизнес-логики
var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidOrder      = errors.New("invalid order data")
	ErrOrderUIDNotUnique = errors.New("order UID is not unique")
	ErrOrderUIDEmpty     = errors.New("order UID cannot be empty")
	ErrInternal          = errors.New("internal server error")
)
