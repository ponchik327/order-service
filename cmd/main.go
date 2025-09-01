package main

import (
	"context"
	"log"
	"net/http"
	"order_service/internal/config"
	"order_service/internal/database"
	"order_service/internal/handler"
	"order_service/internal/middleware"
	"order_service/internal/queue"
	"order_service/internal/repository"
	"order_service/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

// main — точка входа приложения.
func main() {
	// Загрузка конфига
	cfg := config.MustLoad()

	// Инициализация подключения к базе данных
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Инициализация слоев
	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo)
	h := handler.NewOrderHandler(svc, cfg)

	// Настройка маршрутизатора chi
	r := chi.NewRouter()

	// Подключаем логгер
	r.Use(middleware.RequestLogger)

	// Работа с заказами
	r.Route("/order", func(r chi.Router) {
		r.Get("/{orderID}", h.GetOrderByID)  // GET /order/{id} -> получить заказ по ID
		r.Get("/generate", h.GenerateOrders) // GET /order/generate?count=N -> сгенерировать N заказов
		r.Post("/", h.SendOrderToKafka)      // POST /order -> отправить заказ в Kafka
	})

	// Раздача статических файлов (фронтенд)
	r.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir("static"))))

	// Обработка 404
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	// Запуск Kafka Consumer в горутине с сервисом
	go queue.StartKafkaConsumer(svc, cfg)

	srv := &http.Server{
		Addr:         cfg.HttpServer.Adress,
		Handler:      r,
		WriteTimeout: cfg.HttpServer.Timeout,
		ReadTimeout:  cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	// Запуск HTTP-сервера
	go func() {
		log.Printf("Starting server on :%s", cfg.HttpServer.Adress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Создаем канал для перехвата сигналов
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнал
	<-sigchan

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Закрываем соединение с БД
	log.Println("Closing database connection...")
	if err := db.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	} else {
		log.Println("Database connection closed successfully")
	}

	// Выполняем graceful shutdown
	log.Println("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped gracefully")
}
