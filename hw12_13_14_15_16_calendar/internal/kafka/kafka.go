package kafka

import (
	"context"
	"time"
)

type Notification struct {
	EventID   string    `json:"eventId"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"startTime"`
	UserID    string    `json:"userId"`
}

type ProducerInterface interface {
	Send(ctx context.Context, topic string, notification *Notification) error
	Close() error
}

type ConsumerInterface interface {
	Consume(ctx context.Context, topic string, handler func(*Notification) error) error
	Close() error
}

type Config struct {
	BootstrapServers []string `yaml:"bootstrapServers"`
	Topic            string   `yaml:"topic"`
}
