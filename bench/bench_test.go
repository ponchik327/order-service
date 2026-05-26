// Package bench содержит HTTP-уровневые бенчмарки для горячих эндпоинтов.
//
// Бенчмарки поднимают chi-роутер с реальными handler/service и in-memory
// моками для repository и cache. Это даёт стабильные, повторяемые цифры
// без зависимости от Postgres/Redis/Kafka.
//
// Запуск:
//
//	go test -bench=. -benchmem ./bench/...
//
// Сбор профилей:
//
//	go test -bench=BenchmarkGetOrderByID_CacheHit -benchmem \
//	    -cpuprofile=profiles/cpu.out -memprofile=profiles/mem.out ./bench
//	go tool pprof -top profiles/cpu.out
package bench

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"order_service/internal/config"
	"order_service/internal/domain"
	"order_service/internal/handler"
	"order_service/internal/service"

	"github.com/go-chi/chi/v5"
)

func newTestRouter(svc service.OrderService, cfg *config.Config) http.Handler {
	h := handler.NewOrderHandler(svc, cfg)
	r := chi.NewRouter()
	r.Route("/order", func(r chi.Router) {
		r.Get("/{orderID}", h.GetOrderByID)
		r.Get("/generate", h.GenerateOrders)
	})
	return r
}

func testConfig() *config.Config {
	return &config.Config{
		HttpServer: config.HttpServer{
			Adress:      "127.0.0.1:0",
			Timeout:     10 * time.Second,
			IdleTimeout: 60 * time.Second,
		},
	}
}

// BenchmarkGenerateOrders бьёт по GET /order/generate?count=100.
// Чистый CPU/память: генератор + JSON-encoding, никакой БД.
func BenchmarkGenerateOrders(b *testing.B) {
	cfg := testConfig()
	svc := service.NewOrderService(newMockRepo(), cfg, newMockCache())
	srv := httptest.NewServer(newTestRouter(svc, cfg))
	defer srv.Close()

	url := srv.URL + "/order/generate?count=100"
	client := srv.Client()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(url)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkGetOrderByID_CacheHit бьёт по GET /order/{uid} с гарантированным
// попаданием в кэш. Тестирует горячий путь cache hit.
func BenchmarkGetOrderByID_CacheHit(b *testing.B) {
	cfg := testConfig()
	mc := newMockCache()
	mr := newMockRepo()
	svc := service.NewOrderService(mr, cfg, mc)

	order := domain.GenerateRandomOrder()
	_ = mc.SetOrder(context.Background(), order.OrderUID, &order)

	srv := httptest.NewServer(newTestRouter(svc, cfg))
	defer srv.Close()

	url := srv.URL + "/order/" + order.OrderUID
	client := srv.Client()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(url)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkGetOrderByID_DBPath — кэш всегда miss, заказ берётся из mockRepo.
// Тестирует фолбэк-путь через репозиторий и запись результата в кэш.
func BenchmarkGetOrderByID_DBPath(b *testing.B) {
	cfg := testConfig()
	mc := newMockCache()
	mc.always = ModeMiss
	mr := newMockRepo()
	svc := service.NewOrderService(mr, cfg, mc)

	order := domain.GenerateRandomOrder()
	_ = mr.Create(context.Background(), &order)

	srv := httptest.NewServer(newTestRouter(svc, cfg))
	defer srv.Close()

	url := srv.URL + "/order/" + order.OrderUID
	client := srv.Client()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(url)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkGenerateRandomOrder — чистый бенчмарк функции-генератора заказа,
// без HTTP и JSON. Полезен, чтобы изолировать стоимость самой генерации.
func BenchmarkGenerateRandomOrder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = domain.GenerateRandomOrder()
	}
}
