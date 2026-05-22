package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer реализация ProducerInterface
type Producer struct {
	writer *kafka.Writer
	config Config
}

// NewProducer создает новый producer
func NewProducer(cfg Config) (*Producer, error) {
	// В kafka-go, если ты указываешь Topic при создании Writer,
	// то при отправке сообщения (Message) поле Topic должно быть пустым.
	// И наоборот: если Topic не указан в Writer, его нужно указывать в Message.
	// Мы убираем Topic отсюда, так как передаем его динамически в методе Send.
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

// Send отправляет уведомление в топик
func (p *Producer) Send(ctx context.Context, topic string, notification *Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Указываем топик здесь
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

// Close закрывает producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
