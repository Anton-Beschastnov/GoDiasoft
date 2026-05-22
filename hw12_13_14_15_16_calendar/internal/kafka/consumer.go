package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	config Config
}

func NewConsumer(cfg Config) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.BootstrapServers,
		Topic:       cfg.Topic,
		GroupID:     "calendar-storer",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		MaxWait:     10 * time.Second,
	})

	return &Consumer{
		reader: reader,
		config: cfg,
	}, nil
}

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
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return fmt.Errorf("failed to commit message: %w", err)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
