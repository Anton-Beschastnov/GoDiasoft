package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/api"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage/memory"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// helperUUID creates UUID from string for tests.
func helperUUID(s string) openapi_types.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return openapi_types.UUID(u) //nolint:unconvert
}

func newTestLogger() logger.Iface {
	return logger.New("info", nil)
}

func newTestStorage() storage.Storage {
	return memorystorage.New()
}

func newTestHandler() *CalendarHandler {
	return NewCalendarHandler(newTestLogger(), newTestStorage())
}

func TestListEvents(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	// Create test events
	event1 := storage.Event{
		ID:        helperUUID("00000000-0000-0000-0000-000000000001").String(),
		Title:     "Event 1",
		StartTime: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		UserID:    "user1",
	}
	event2 := storage.Event{
		ID:        helperUUID("00000000-0000-0000-0000-000000000002").String(),
		Title:     "Event 2",
		StartTime: time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC),
		UserID:    "user1",
	}
	event3 := storage.Event{
		ID:        helperUUID("00000000-0000-0000-0000-000000000003").String(),
		Title:     "Event 3",
		StartTime: time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
		UserID:    "user1",
	}

	_ = h.storage.CreateEvent(ctx, event1)
	_ = h.storage.CreateEvent(ctx, event2)
	_ = h.storage.CreateEvent(ctx, event3)

	// Test missing user_id
	req := httptest.NewRequest(http.MethodGet, "/events?start_date=2024-01-15", nil)
	w := httptest.NewRecorder()
	h.ListEvents(w, req, api.ListEventsParams{})
	if resp := w.Result(); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	// Test missing start_date
	req = httptest.NewRequest(http.MethodGet, "/events?user_id=user1", nil)
	w := httptest.NewRecorder()
	h.ListEvents(w, req, api.ListEventsParams{UserId: "user1"})
	if resp := w.Result(); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	// Test day period
	day := api.Day
	req = httptest.NewRequest(http.MethodGet, "/events?user_id=user1&start_date=2024-01-15", nil)
	w = httptest.NewRecorder()
	h.ListEvents(w, req, api.ListEventsParams{
		UserId:    "user1",
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Period:    &day,
	})
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var events []api.Event
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	// Test week period
	week := api.Week
	req = httptest.NewRequest(http.MethodGet, "/events?user_id=user1&start_date=2024-01-15", nil)
	w = httptest.NewRecorder()
	h.ListEvents(w, req, api.ListEventsParams{
		UserId:    "user1",
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Period:    &week,
	})
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events for week, got %d", len(events))
	}

	// Test month period
	month := api.Month
	req = httptest.NewRequest(http.MethodGet, "/events?user_id=user1&start_date=2024-01-01", nil)
	w = httptest.NewRecorder()
	h.ListEvents(w, req, api.ListEventsParams{
		UserId:    "user1",
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Period:    &month,
	})
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events for month, got %d", len(events))
	}
}

func TestCreateEvent(t *testing.T) {
	h := newTestHandler()

	// Test missing title
	reqBody := map[string]interface{}{
		"start_time": time.Now().Format(time.RFC3339),
		"end_time":   time.Now().Add(time.Hour).Format(time.RFC3339),
		"user_id":    "user1",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateEvent(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	// Test missing start_time
	reqBody = map[string]interface{}{
		"title":    "Test Event",
		"end_time": time.Now().Add(time.Hour).Format(time.RFC3339),
		"user_id":  "user1",
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w = httptest.NewRecorder()
	h.CreateEvent(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	// Test missing end_time
	reqBody = map[string]interface{}{
		"title":      "Test Event",
		"start_time": time.Now().Format(time.RFC3339),
		"user_id":    "user1",
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w = httptest.NewRecorder()
	h.CreateEvent(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	// Test missing user_id
	reqBody = map[string]interface{}{
		"title":      "Test Event",
		"start_time": time.Now().Format(time.RFC3339),
		"end_time":   time.Now().Add(time.Hour).Format(time.RFC3339),
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateEvent(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	// Test valid event creation
	reqBody = map[string]interface{}{
		"title":      "Test Event",
		"start_time": time.Now().Format(time.RFC3339),
		"end_time":   time.Now().Add(time.Hour).Format(time.RFC3339),
		"user_id":    "user1",
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateEvent(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var createdEvent api.Event
	if err := json.NewDecoder(w.Body).Decode(&createdEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if createdEvent.Title != "Test Event" {
		t.Errorf("expected title 'Test Event', got %s", createdEvent.Title)
	}
	if createdEvent.UserID != "user1" {
		t.Errorf("expected user_id 'user1', got %s", createdEvent.UserID)
	}

	// Test duplicate event (date busy) - same time slot
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w = httptest.NewRecorder()
	h.CreateEvent(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestGetEvent(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	// Create test event
	event := storage.Event{
		ID:        helperUUID("00000000-0000-0000-0000-000000000001").String(),
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}
	_ = h.storage.CreateEvent(ctx, event)

	// Test get existing event
	req := httptest.NewRequest(http.MethodGet, "/events/"+event.ID, nil)
	w := httptest.NewRecorder()
	h.GetEvent(w, req, helperUUID(event.ID))
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var gotEvent api.Event
	if err := json.NewDecoder(w.Body).Decode(&gotEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if gotEvent.Title != "Test Event" {
		t.Errorf("expected title 'Test Event', got %s", gotEvent.Title)
	}

	// Test get non-existing event
	req = httptest.NewRequest(http.MethodGet, "/events/nonexistent", nil)
	w = httptest.NewRecorder()
	h.GetEvent(w, req, helperUUID("00000000-0000-0000-0000-000000000000"))
	if resp := w.Result(); resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestUpdateEvent(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	// Create test event
	event := storage.Event{
		ID:        helperUUID("00000000-0000-0000-0000-000000000001").String(),
		Title:     "Original Title",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}
	_ = h.storage.CreateEvent(ctx, event)

	// Test update non-existing event
	reqBody := map[string]interface{}{
		"title": "Updated Title",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/events/nonexistent", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateEvent(w, req, helperUUID("00000000-0000-0000-0000-000000000000"))
	if resp := w.Result(); resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}

	// Test update existing event
	reqBody = map[string]interface{}{
		"title": "Updated Title",
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPut, "/events/"+event.ID, bytes.NewReader(body))
	w = httptest.NewRecorder()
	h.UpdateEvent(w, req, helperUUID(event.ID))
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var updatedEvent api.Event
	if err := json.NewDecoder(w.Body).Decode(&updatedEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if updatedEvent.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", updatedEvent.Title)
	}
}

func TestDeleteEvent(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	// Create test event
	event := storage.Event{
		ID:        helperUUID("00000000-0000-0000-0000-000000000001").String(),
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}
	_ = h.storage.CreateEvent(ctx, event)

	// Test delete non-existing event
	req := httptest.NewRequest(http.MethodDelete, "/events/nonexistent", nil)
	w := httptest.NewRecorder()
	h.DeleteEvent(w, req, helperUUID("00000000-0000-0000-0000-000000000000"))
	if resp := w.Result(); resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}

	// Test delete existing event
	req = httptest.NewRequest(http.MethodDelete, "/events/"+event.ID, nil)
	w := httptest.NewRecorder()
	h.DeleteEvent(w, req, helperUUID(event.ID))
	if resp := w.Result(); resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", resp.StatusCode)
	}

	// Verify event is deleted
	req = httptest.NewRequest(http.MethodGet, "/events/"+event.ID, nil)
	w = httptest.NewRecorder()
	h.GetEvent(w, req, helperUUID(event.ID))
	if resp := w.Result(); resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404 after delete, got %d", resp.StatusCode)
	}
}

func TestCreateEventInvalidJSON(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	h.CreateEvent(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateEventInvalidJSON(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPut, "/events/nonexistent", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	h.UpdateEvent(w, req, helperUUID("00000000-0000-0000-0000-000000000000"))
	if resp := w.Result(); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestListEventsDifferentUser(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	// Create events for different users
	event1 := storage.Event{
		ID:        helperUUID("00000000-0000-0000-0000-000000000001").String(),
		Title:     "User1 Event",
		StartTime: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		UserID:    "user1",
	}
	event2 := storage.Event{
		ID:        helperUUID("00000000-0000-0000-0000-000000000002").String(),
		Title:     "User2 Event",
		StartTime: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		UserID:    "user2",
	}

	_ = h.storage.CreateEvent(ctx, event1)
	_ = h.storage.CreateEvent(ctx, event2)

	// List events for user1
	req := httptest.NewRequest(http.MethodGet, "/events?user_id=user1&start_date=2024-01-15", nil)
	w := httptest.NewRecorder()
	day := api.Day
	h.ListEvents(w, req, api.ListEventsParams{
		UserId:    "user1",
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Period:    &day,
	})
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var events []api.Event
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event for user1, got %d", len(events))
	}
	if events[0].Title != "User1 Event" {
		t.Errorf("expected 'User1 Event', got %s", events[0].Title)
	}

	// List events for user2
	req = httptest.NewRequest(http.MethodGet, "/events?user_id=user2&start_date=2024-01-15", nil)
	w = httptest.NewRecorder()
	h.ListEvents(w, req, api.ListEventsParams{
		UserId:    "user2",
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Period:    &day,
	})
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event for user2, got %d", len(events))
	}
	if events[0].Title != "User2 Event" {
		t.Errorf("expected 'User2 Event', got %s", events[0].Title)
	}
}

func TestCreateEventWithOptionalFields(t *testing.T) {
	h := newTestHandler()

	// Test event with all optional fields
	reqBody := map[string]interface{}{
		"title":         "Event with all fields",
		"start_time":    time.Now().Format(time.RFC3339),
		"end_time":      time.Now().Add(time.Hour).Format(time.RFC3339),
		"user_id":       "user1",
		"description":   "Test description",
		"notify_before": int64(30),
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateEvent(w, req)
	if resp := w.Result(); resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var createdEvent api.Event
	if err := json.NewDecoder(w.Body).Decode(&createdEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if createdEvent.Description != "Test description" {
		t.Errorf("expected description 'Test description', got %s", createdEvent.Description)
	}
	if createdEvent.NotifyBefore != 30 {
		t.Errorf("expected notify_before 30, got %d", createdEvent.NotifyBefore)
	}
}

func TestUpdateEventPartialUpdate(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	// Create test event
	event := storage.Event{
		ID:          helperUUID("00000000-0000-0000-0000-000000000001").String(),
		Title:       "Original Title",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		UserID:      "user1",
		Description: "Original Description",
	}
	_ = h.storage.CreateEvent(ctx, event)

	// Test partial update (only title)
	reqBody := map[string]interface{}{
		"title": "Updated Title",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/events/"+event.ID, bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateEvent(w, req, helperUUID(event.ID))
	if resp := w.Result(); resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var updatedEvent api.Event
	if err := json.NewDecoder(w.Body).Decode(&updatedEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if updatedEvent.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", updatedEvent.Title)
	}
	if updatedEvent.Description != "Original Description" {
		t.Errorf("expected description to remain unchanged, got %s", updatedEvent.Description)
	}
}
