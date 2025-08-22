package main

import (
	"log"
	"net/http"
	"order_service/internal/database"
	"order_service/internal/handler"
	"order_service/internal/repository"
	"order_service/internal/service"
	"os"

	"github.com/go-chi/chi/v5"
)

// main — точка входа приложения.
func main() {
	// Инициализация подключения к базе данных
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация слоев
	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo)
	h := handler.NewOrderHandler(svc)

	// Настройка маршрутизатора chi
	r := chi.NewRouter()

	// Эндпоинты
	r.Post("/order", h.CreateOrder)
	r.Get("/order/{orderID}", h.GetOrderByID)

	// Получение порта из переменной окружения или использование значения по умолчанию
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Запуск HTTP-сервера
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
