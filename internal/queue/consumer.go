package queue

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"order_service/internal/config"
	"order_service/internal/domain"
	"order_service/internal/service"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func StartKafkaConsumer(svc service.OrderService, cfg *config.Config) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Adress,
		"group.id":          cfg.Kafka.GroupId,
		"auto.offset.reset": cfg.Kafka.OffsetReset,
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}
	defer consumer.Close()

	topic := cfg.Kafka.Topic
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe: %s", err)
	}

	log.Printf("Successfully subscribe to kafka topic:%s", topic)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating", sig)
			return
		default:
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				var order domain.Order
				if err := json.Unmarshal(e.Value, &order); err != nil {
					log.Printf("Failed to unmarshal order: %s", err)
					// DLQ если нужно
					continue
				}

				// Вызов сервиса для сохранения
				ctx := context.Background() // Или с таймаутом
				if err := svc.CreateOrder(ctx, &order); err != nil {
					log.Printf("Failed to process order %s: %s", order.OrderUID, err)
					// DLQ или retry
					continue
				}

				log.Printf("Successfully processed order: %s", order.OrderUID)
				consumer.CommitMessage(e)
			case kafka.Error:
				log.Printf("Kafka error: %v", e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			}
		}
	}
}
