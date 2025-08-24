package queue

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

func StartKafkaProducer() (*kafka.Producer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:29092",
	}
	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}
