package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"order_service/internal/domain"
	"order_service/internal/service"
)

// OrderHandler определяет интерфейс для HTTP-хендлеров заказов.
type OrderHandler interface {
	CreateOrder(w http.ResponseWriter, r *http.Request)
	GetOrderByID(w http.ResponseWriter, r *http.Request)
}

// orderHandler — реализация OrderHandler.
type orderHandler struct {
	service service.OrderService
}

// NewOrderHandler создает новый экземпляр orderHandler.
func NewOrderHandler(service service.OrderService) OrderHandler {
	return &orderHandler{service: service}
}

// CreateOrder обрабатывает POST /order/ для создания заказа.
func (h *orderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Декодируем JSON из тела запроса
	var order domain.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаем контекст из запроса
	ctx := r.Context()

	// Вызываем сервис для создания заказа
	if err := h.service.CreateOrder(ctx, &order); err != nil {
		http.Error(w, "Failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Возвращаем созданный заказ
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetOrderByID обрабатывает GET /order/{orderUID} для получения заказа.
func (h *orderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем orderUID из URL (например, /order/b563feb7b2b84b6test)
	orderUID := r.URL.Path[len("/order/"):]
	if orderUID == "" {
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	// Создаем контекст из запроса
	ctx := r.Context()

	// Вызываем сервис для получения заказа
	order, err := h.service.GetOrderByID(ctx, orderUID)
	if err != nil {
		if err.Error() == "order_uid cannot be empty" {
			http.Error(w, "Invalid order_uid", http.StatusBadRequest)
			return
		}
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Возвращаем заказ
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
