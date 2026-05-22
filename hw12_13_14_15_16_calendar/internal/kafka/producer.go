package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	config Config
}

func NewProducer(cfg Config) (*Producer, error) {

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.BootstrapServers...),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 100 * time.Millisecond,
		Compression:  kafka.Snappy,
	}

	return &Producer{
		writer: writer,
		config: cfg,
	}, nil
}

func (p *Producer) Send(ctx context.Context, topic string, notification *Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
