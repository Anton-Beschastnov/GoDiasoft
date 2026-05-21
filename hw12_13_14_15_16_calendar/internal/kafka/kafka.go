package kafka

import (
	"context"
	"time"
)

// Notification представляет уведомление о событии
type Notification struct {
	EventID   string    `json:"event_id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	UserID    string    `json:"user_id"`
}

// ProducerInterface интерфейс для отправки сообщений в Kafka
type ProducerInterface interface {
	// Send отправляет уведомление в топик
	Send(ctx context.Context, topic string, notification *Notification) error
	// Close закрывает producer
	Close() error
}

// ConsumerInterface интерфейс для чтения сообщений из Kafka
type ConsumerInterface interface {
	// Consume читает сообщения из топика
	Consume(ctx context.Context, topic string, handler func(*Notification) error) error
	// Close закрывает consumer
	Close() error
}

// Config конфигурация подключения к Kafka
type Config struct {
	BootstrapServers []string `yaml:"bootstrap_servers"`
	Topic            string   `yaml:"topic"`
}
