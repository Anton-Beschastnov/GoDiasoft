package memorystorage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage"
)

func TestCreateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "1",
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}

	err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := s.GetEvent(ctx, "1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ID != event.ID || got.Title != event.Title {
		t.Errorf("expected event %+v, got %+v", event, got)
	}
}

func TestCreateEventInvalidData(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:    "",
		Title: "Test",
	}

	err := s.CreateEvent(ctx, event)
	if err != storage.ErrInvalidEvent {
		t.Errorf("expected ErrInvalidEvent, got %v", err)
	}

	event = storage.Event{
		ID:    "1",
		Title: "",
	}

	err = s.CreateEvent(ctx, event)
	if err != storage.ErrInvalidEvent {
		t.Errorf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestCreateEventDuplicateID(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "1",
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}

	_ = s.CreateEvent(ctx, event)

	err := s.CreateEvent(ctx, event)
	if err != storage.ErrDateBusy {
		t.Errorf("expected ErrDateBusy, got %v", err)
	}
}

func TestCreateEventDateBusy(t *testing.T) {
	s := New()
	ctx := context.Background()

	start := time.Now()
	event1 := storage.Event{
		ID:        "1",
		Title:     "Event 1",
		StartTime: start,
		EndTime:   start.Add(2 * time.Hour),
		UserID:    "user1",
	}

	_ = s.CreateEvent(ctx, event1)

	event2 := storage.Event{
		ID:        "2",
		Title:     "Event 2",
		StartTime: start.Add(time.Hour),
		EndTime:   start.Add(3 * time.Hour),
		UserID:    "user1",
	}

	err := s.CreateEvent(ctx, event2)
	if err != storage.ErrDateBusy {
		t.Errorf("expected ErrDateBusy, got %v", err)
	}

	event3 := storage.Event{
		ID:        "3",
		Title:     "Event 3",
		StartTime: start.Add(time.Hour),
		EndTime:   start.Add(3 * time.Hour),
		UserID:    "user2",
	}

	err = s.CreateEvent(ctx, event3)
	if err != nil {
		t.Errorf("expected no error for different user, got %v", err)
	}
}

func TestUpdateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "1",
		Title:     "Original Title",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}

	_ = s.CreateEvent(ctx, event)

	event.Title = "Updated Title"
	err := s.UpdateEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, _ := s.GetEvent(ctx, "1")
	if got.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", got.Title)
	}
}

func TestUpdateEventNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:    "nonexistent",
		Title: "Test",
	}

	err := s.UpdateEvent(ctx, event)
	if err != storage.ErrEventNotFound {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "1",
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}

	_ = s.CreateEvent(ctx, event)

	err := s.DeleteEvent(ctx, "1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = s.GetEvent(ctx, "1")
	if err != storage.ErrEventNotFound {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEventNotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	err := s.DeleteEvent(ctx, "nonexistent")
	if err != storage.ErrEventNotFound {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestListEventsForDay(t *testing.T) {
	s := New()
	ctx := context.Background()

	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	event1 := storage.Event{
		ID:        "1",
		Title:     "Event 1",
		StartTime: date.Add(10 * time.Hour),
		EndTime:   date.Add(11 * time.Hour),
		UserID:    "user1",
	}

	event2 := storage.Event{
		ID:        "2",
		Title:     "Event 2",
		StartTime: date.Add(14 * time.Hour),
		EndTime:   date.Add(15 * time.Hour),
		UserID:    "user1",
	}

	event3 := storage.Event{
		ID:        "3",
		Title:     "Event 3",
		StartTime: date.AddDate(0, 0, 1).Add(10 * time.Hour),
		EndTime:   date.AddDate(0, 0, 1).Add(11 * time.Hour),
		UserID:    "user1",
	}

	_ = s.CreateEvent(ctx, event1)
	_ = s.CreateEvent(ctx, event2)
	_ = s.CreateEvent(ctx, event3)

	events, err := s.ListEventsForDay(ctx, "user1", date)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestListEventsForWeek(t *testing.T) {
	s := New()
	ctx := context.Background()

	startOfWeek := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 10; i++ {
		event := storage.Event{
			ID:        string(rune('a' + i)),
			Title:     "Event",
			StartTime: startOfWeek.AddDate(0, 0, i),
			EndTime:   startOfWeek.AddDate(0, 0, i).Add(time.Hour),
			UserID:    "user1",
		}
		_ = s.CreateEvent(ctx, event)
	}

	events, err := s.ListEventsForWeek(ctx, "user1", startOfWeek)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 7 {
		t.Errorf("expected 7 events, got %d", len(events))
	}
}

func TestListEventsForMonth(t *testing.T) {
	s := New()
	ctx := context.Background()

	startOfMonth := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 35; i++ {
		event := storage.Event{
			ID:        string(rune(i + 100)),
			Title:     "Event",
			StartTime: startOfMonth.AddDate(0, 0, i),
			EndTime:   startOfMonth.AddDate(0, 0, i).Add(time.Hour),
			UserID:    "user1",
		}
		_ = s.CreateEvent(ctx, event)
	}

	events, err := s.ListEventsForMonth(ctx, "user1", startOfMonth)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 31 {
		t.Errorf("expected 31 events (January), got %d", len(events))
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := storage.Event{
				ID:        string(rune(id + 1000)),
				Title:     "Concurrent Event",
				StartTime: time.Now().Add(time.Duration(id) * time.Hour * 24),
				EndTime:   time.Now().Add(time.Duration(id)*time.Hour*24 + time.Hour),
				UserID:    "user1",
			}

			_ = s.CreateEvent(ctx, event)
		}(i)
	}

	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = s.GetEvent(ctx, string(rune(id+1000)))
		}(i)
	}

	wg.Wait()
}

func TestConcurrentReadWrite(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "concurrent-test",
		Title:     "Test",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}
	_ = s.CreateEvent(ctx, event)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			_, _ = s.GetEvent(ctx, "concurrent-test")
		}()

		go func(i int) {
			defer wg.Done()
			event := storage.Event{
				ID:        "concurrent-test",
				Title:     "Updated",
				StartTime: time.Now().Add(time.Duration(i) * time.Minute),
				EndTime:   time.Now().Add(time.Duration(i)*time.Minute + time.Hour),
				UserID:    "user1",
			}
			_ = s.UpdateEvent(ctx, event)
		}(i)
	}

	wg.Wait()
}
