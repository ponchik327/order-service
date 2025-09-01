package queue

import (
	"order_service/internal/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func StartKafkaProducer(cfg *config.Config) (*kafka.Producer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Adress,
	}
	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}
