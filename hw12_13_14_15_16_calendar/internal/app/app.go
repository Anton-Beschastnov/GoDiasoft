package app

import (
	"context"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/internal/storage"
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEvent(ctx context.Context, id string) (storage.Event, error)
	ListEventsForDay(ctx context.Context, userID string, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, userID string, startDate time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, userID string, startDate time.Time) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	a.logger.Debug("creating event", "id", event.ID, "title", event.Title)
	return a.storage.CreateEvent(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, event storage.Event) error {
	a.logger.Debug("updating event", "id", event.ID)
	return a.storage.UpdateEvent(ctx, event)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	a.logger.Debug("deleting event", "id", id)
	return a.storage.DeleteEvent(ctx, id)
}

func (a *App) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	return a.storage.GetEvent(ctx, id)
}

func (a *App) ListEventsForDay(ctx context.Context, userID string, date time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForDay(ctx, userID, date)
}

func (a *App) ListEventsForWeek(ctx context.Context, userID string, startDate time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForWeek(ctx, userID, startDate)
}

func (a *App) ListEventsForMonth(ctx context.Context, userID string, startDate time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForMonth(ctx, userID, startDate)
}
