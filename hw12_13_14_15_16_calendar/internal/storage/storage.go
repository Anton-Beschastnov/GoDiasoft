package storage

import (
	"context"
	"errors"
	"time"
)

var ErrNotificationNotFound = errors.New("notification not found")

type Storage interface {
	CreateEvent(ctx context.Context, event Event) error
	UpdateEvent(ctx context.Context, event Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEvent(ctx context.Context, id string) (Event, error)
	ListEventsForDay(ctx context.Context, userID string, date time.Time) ([]Event, error)
	ListEventsForWeek(ctx context.Context, userID string, startDate time.Time) ([]Event, error)
	ListEventsForMonth(ctx context.Context, userID string, startDate time.Time) ([]Event, error)
	// Методы для scheduler.
	// GetEventsForNotification возвращает события, для которых нужно отправить уведомление.
	GetEventsForNotification(ctx context.Context, now time.Time) ([]Event, error)
	// DeleteOldEvents удаляет события старше 1 года.
	DeleteOldEvents(ctx context.Context, cutoff time.Time) error
	// Методы для storer.
	// SaveNotification сохраняет уведомление.
	SaveNotification(ctx context.Context, notification *Notification) error
	// GetNotificationByID возвращает уведомление по ID.
	GetNotificationByID(ctx context.Context, id string) (*Notification, error)
}

// Notification представляет уведомление о событии.
type Notification struct {
	ID          string     `json:"id"`
	EventID     string     `json:"eventId"`
	Title       string     `json:"title"`
	StartTime   time.Time  `json:"startTime"`
	UserID      string     `json:"userId"`
	CreatedAt   time.Time  `json:"createdAt"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
}
