package storage

import (
	"context"
	"time"
)

type Storage interface {
	CreateEvent(ctx context.Context, event Event) error
	UpdateEvent(ctx context.Context, event Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEvent(ctx context.Context, id string) (Event, error)
	ListEventsForDay(ctx context.Context, userID string, date time.Time) ([]Event, error)
	ListEventsForWeek(ctx context.Context, userID string, startDate time.Time) ([]Event, error)
	ListEventsForMonth(ctx context.Context, userID string, startDate time.Time) ([]Event, error)
}
