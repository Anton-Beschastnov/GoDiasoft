package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage struct {
	mu            sync.RWMutex
	events        map[string]storage.Event
	notifications *notifications
}

func New() *Storage {
	return &Storage{
		events:        make(map[string]storage.Event),
		notifications: newNotifications(),
	}
}

func (s *Storage) CreateEvent(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID == "" || event.Title == "" {
		return storage.ErrInvalidEvent
	}

	if _, exists := s.events[event.ID]; exists {
		return storage.ErrDateBusy
	}

	if s.isTimeBusy(event) {
		return storage.ErrDateBusy
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; !exists {
		return storage.ErrEventNotFound
	}

	for id, e := range s.events {
		if id != event.ID && e.UserID == event.UserID && s.eventsOverlap(e, event) {
			return storage.ErrDateBusy
		}
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *Storage) GetEvent(_ context.Context, id string) (storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return storage.Event{}, storage.ErrEventNotFound
	}

	return event, nil
}

func (s *Storage) ListEventsForDay(_ context.Context, userID string, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.filterEvents(userID, startOfDay, endOfDay), nil
}

func (s *Storage) ListEventsForWeek(_ context.Context, userID string, startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	startOfWeek := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return s.filterEvents(userID, startOfWeek, endOfWeek), nil
}

func (s *Storage) ListEventsForMonth(_ context.Context, userID string, startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	startOfMonth := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return s.filterEvents(userID, startOfMonth, endOfMonth), nil
}

func (s *Storage) GetEventsForNotification(_ context.Context, now time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	for _, event := range s.events {
		if event.NotifyBefore > 0 &&
			!event.StartTime.Add(-event.NotifyBefore).Before(now) &&
			event.StartTime.After(now) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (s *Storage) DeleteOldEvents(_ context.Context, cutoff time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, event := range s.events {
		if event.StartTime.Before(cutoff) {
			delete(s.events, id)
		}
	}
	return nil
}

type notifications struct {
	mu            sync.RWMutex
	notifications map[string]*storage.Notification
}

func newNotifications() *notifications {
	return &notifications{
		notifications: make(map[string]*storage.Notification),
	}
}

func (n *notifications) Save(_ context.Context, notification *storage.Notification) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.notifications[notification.ID] = notification
	return nil
}

func (n *notifications) Get(_ context.Context, id string) (*storage.Notification, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	notification, exists := n.notifications[id]
	if !exists {
		return nil, storage.ErrNotificationNotFound
	}

	return notification, nil
}

func (s *Storage) SaveNotification(ctx context.Context, notification *storage.Notification) error {
	return s.notifications.Save(ctx, notification)
}

func (s *Storage) GetNotificationByID(ctx context.Context, id string) (*storage.Notification, error) {
	return s.notifications.Get(ctx, id)
}

func (s *Storage) filterEvents(userID string, start, end time.Time) []storage.Event {
	var result []storage.Event
	for _, event := range s.events {
		if event.UserID == userID && !event.StartTime.Before(start) && event.StartTime.Before(end) {
			result = append(result, event)
		}
	}
	return result
}

func (s *Storage) isTimeBusy(newEvent storage.Event) bool {
	for _, event := range s.events {
		if event.UserID == newEvent.UserID && s.eventsOverlap(event, newEvent) {
			return true
		}
	}
	return false
}

func (s *Storage) eventsOverlap(e1, e2 storage.Event) bool {
	return e1.StartTime.Before(e2.EndTime) && e2.StartTime.Before(e1.EndTime)
}
