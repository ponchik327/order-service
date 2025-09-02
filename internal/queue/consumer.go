package queue

import (
	"context"
	"fmt"
	"log"
	"sync"

	"order_service/internal/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type OrderHandler interface {
	HandleOrder(ctx context.Context, message []byte) error
}

type KafkaConsumer interface {
	Start()
	Stop()
}

type kafkaConsumer struct {
	consumer *kafka.Consumer
	config   *config.Config
	handler  OrderHandler
	wg       *sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// Инициализирует новый консюмер
func NewKafkaConsumer(handler OrderHandler, config *config.Config) (KafkaConsumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": config.Kafka.Adress,
		"group.id":          config.Kafka.GroupId,
		"auto.offset.reset": config.Kafka.OffsetReset,
	}

	consumer, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	err = consumer.SubscribeTopics([]string{config.Kafka.Topic}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &kafkaConsumer{
		consumer: consumer,
		config:   config,
		handler:  handler,
		wg:       &sync.WaitGroup{},
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Стартует работу консюмера
func (k *kafkaConsumer) Start() {
	go func() {
		for {
			select {
			case <-k.ctx.Done():
				return
			default:
				event := k.consumer.Poll(100)
				if event == nil {
					continue
				}
				k.consume(event)
			}
		}
	}()
}

// Останавливает работу консюмера
func (k *kafkaConsumer) Stop() {
	k.cancel()
	k.wg.Wait()
	k.consumer.Close()
}

// Обрабатывает событие очереди
func (k *kafkaConsumer) consume(event kafka.Event) {
	k.wg.Add(1)
	defer k.wg.Done()
	switch e := event.(type) {
	case *kafka.Message:
		err := k.handler.HandleOrder(k.ctx, e.Value) // При отмене контекста транзакция бд ролбекнится, кафка не закомитится
		if err != nil {
			log.Printf("Failed handle order: %s", err)
			return
		}
		k.consumer.CommitMessage(e)
	case kafka.Error:
		log.Printf("Kafka error: %v", e)
		return
	}
}
