package app

import (
	"context"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage"
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
	GetEventsForNotification(ctx context.Context, now time.Time) ([]storage.Event, error)
	DeleteOldEvents(ctx context.Context, cutoff time.Time) error
	SaveNotification(ctx context.Context, notification *storage.Notification) error
	GetNotificationByID(ctx context.Context, id string) (*storage.Notification, error)
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

func (a *App) GetEventsForNotification(ctx context.Context, now time.Time) ([]storage.Event, error) {
	return a.storage.GetEventsForNotification(ctx, now)
}

func (a *App) DeleteOldEvents(ctx context.Context, cutoff time.Time) error {
	return a.storage.DeleteOldEvents(ctx, cutoff)
}

func (a *App) SaveNotification(ctx context.Context, notification *storage.Notification) error {
	return a.storage.SaveNotification(ctx, notification)
}

func (a *App) GetNotificationByID(ctx context.Context, id string) (*storage.Notification, error) {
	return a.storage.GetNotificationByID(ctx, id)
}
