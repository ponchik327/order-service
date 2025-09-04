package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"order_service/internal/config"
	"order_service/internal/domain"
	"order_service/internal/queue"
	"order_service/internal/service"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// OrderHandler определяет интерфейс для HTTP-хендлеров заказов.
type OrderHandler interface {
	GetOrderByID(w http.ResponseWriter, r *http.Request)
	GenerateOrders(w http.ResponseWriter, r *http.Request)
	SendOrderToKafka(w http.ResponseWriter, r *http.Request)
}

// orderHandler — реализация OrderHandler.
type orderHandler struct {
	service service.OrderService
	config  *config.Config
}

// NewOrderHandler создает новый экземпляр orderHandler.
func NewOrderHandler(service service.OrderService, config *config.Config) OrderHandler {
	return &orderHandler{
		service: service,
		config:  config,
	}
}

// GetOrderByID обрабатывает GET /order/{orderUID} для получения заказа.
func (h *orderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем orderUID из URL
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
		if errors.Is(err, domain.ErrOrderUIDNotUnique) || errors.Is(err, domain.ErrOrderUIDEmpty) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if errors.Is(err, domain.ErrOrderNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		if errors.Is(err, domain.ErrInternal) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

func (h *orderHandler) GenerateOrders(w http.ResponseWriter, r *http.Request) {
	// Получаем параметр count из query
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 {
		http.Error(w, `{"error": "Invalid count parameter"}`, http.StatusBadRequest)
		return
	}

	// Ограничиваем максимальное количество заказов
	if count > 1000 {
		http.Error(w, `{"error": "Count exceeds maximum limit of 1000"}`, http.StatusBadRequest)
		return
	}

	// Генерируем заказы
	orders := make([]domain.Order, count)
	for i := 0; i < count; i++ {
		orders[i] = domain.GenerateRandomOrder()
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Сериализуем и отправляем ответ
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *orderHandler) SendOrderToKafka(w http.ResponseWriter, r *http.Request) {
	var order domain.Order

	// Декодируем JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, `{"error": "Invalid order JSON"}`, http.StatusBadRequest)
		return
	}

	// Сериализация в JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		http.Error(w, `{"error": "Failed to marshal order"}`, http.StatusInternalServerError)
		return
	}

	// Инициализация Kafka producer
	producer, err := queue.StartKafkaProducer(h.config)
	if err != nil {
		http.Error(w, `{"error": "Failed to connect to Kafka"}`, http.StatusInternalServerError)
		return
	}
	defer producer.Close()

	// Отправляем сообщение в Kafka
	topic := "orders"
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          orderJSON,
	}, nil)

	if err != nil {
		http.Error(w, `{"error": "Failed to send message to Kafka"}`, http.StatusInternalServerError)
		return
	}

	// Ждём подтверждения (или таймаута)
	producer.Flush(1000)

	// Ответ клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Order sent to Kafka successfully"})
}
