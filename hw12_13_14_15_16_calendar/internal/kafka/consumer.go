package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer реализация ConsumerInterface
type Consumer struct {
	reader *kafka.Reader
	config Config
}

// NewConsumer создает новый consumer
func NewConsumer(cfg Config) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.BootstrapServers,
		Topic:       cfg.Topic,
		GroupID:     "calendar-storer",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     10 * time.Second,
	})

	return &Consumer{
		reader: reader,
		config: cfg,
	}, nil
}

// Consume читает сообщения из топика и вызывает handler для каждого уведомления
func (c *Consumer) Consume(ctx context.Context, _ string, handler func(*Notification) error) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				continue
			}

			var notification Notification
			if err := json.Unmarshal(msg.Value, &notification); err != nil {
				continue
			}

			if err := handler(&notification); err != nil {
				// Логирование ошибки, но продолжаем обработку
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return fmt.Errorf("failed to commit message: %w", err)
			}
		}
	}
}

// Close закрывает consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
