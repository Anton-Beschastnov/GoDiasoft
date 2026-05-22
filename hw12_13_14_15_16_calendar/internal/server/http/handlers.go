package internalhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/api"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type CalendarHandler struct {
	storage storage.Storage
	logger  logger.Iface
}

func NewCalendarHandler(log logger.Iface, st storage.Storage) *CalendarHandler {
	return &CalendarHandler{
		storage: st,
		logger:  log,
	}
}

func (h *CalendarHandler) ListEvents(w http.ResponseWriter, r *http.Request, params api.ListEventsParams) {
	ctx := r.Context()

	userID := params.UserId
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	startDate := params.StartDate
	if startDate.IsZero() {
		http.Error(w, "start_date is required", http.StatusBadRequest)
		return
	}

	period := api.Day
	if params.Period != nil {
		period = *params.Period
	}

	var events []storage.Event
	var err error

	switch period {
	case api.Day:
		events, err = h.storage.ListEventsForDay(ctx, userID, startDate)
	case api.Week:
		events, err = h.storage.ListEventsForWeek(ctx, userID, startDate)
	case api.Month:
		events, err = h.storage.ListEventsForMonth(ctx, userID, startDate)
	default:
		events, err = h.storage.ListEventsForDay(ctx, userID, startDate)
	}

	if err != nil {
		h.logger.Error("failed to list events", "error", err)
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}

	h.logger.Info("events listed", "count", len(events))

	response := make([]api.Event, len(events))
	for i, event := range events {
		response[i] = h.eventToAPI(event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *CalendarHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.StartTime.IsZero() {
		http.Error(w, "start_time is required", http.StatusBadRequest)
		return
	}
	if req.EndTime.IsZero() {
		http.Error(w, "end_time is required", http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	event := storage.Event{
		ID:           uuid.New().String(),
		Title:        req.Title,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Description:  "",
		UserID:       req.UserID,
		NotifyBefore: 0,
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.NotifyBefore != nil {
		event.NotifyBefore = time.Duration(*req.NotifyBefore)
	}

	if err := h.storage.CreateEvent(ctx, event); err != nil {
		if errors.Is(err, storage.ErrDateBusy) {
			http.Error(w, "date is busy", http.StatusConflict)
			return
		}
		if errors.Is(err, storage.ErrInvalidEvent) {
			http.Error(w, "invalid event data", http.StatusBadRequest)
			return
		}
		h.logger.Error("failed to create event", "error", err)
		http.Error(w, "failed to create event", http.StatusInternalServerError)
		return
	}

	h.logger.Info("event created", "id", event.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(h.eventToAPI(event)); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *CalendarHandler) GetEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	event, err := h.storage.GetEvent(ctx, id.String())
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get event", "error", err)
		http.Error(w, "failed to get event", http.StatusInternalServerError)
		return
	}

	h.logger.Info("event retrieved", "id", id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(h.eventToAPI(event)); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *CalendarHandler) UpdateEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	var req api.UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.storage.GetEvent(ctx, id.String())
	if err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get event", "error", err)
		http.Error(w, "failed to get event", http.StatusInternalServerError)
		return
	}

	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.StartTime != nil {
		event.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		event.EndTime = *req.EndTime
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.NotifyBefore != nil {
		event.NotifyBefore = time.Duration(*req.NotifyBefore)
	}

	if err := h.storage.UpdateEvent(ctx, event); err != nil {
		h.logger.Error("failed to update event", "error", err)
		http.Error(w, "failed to update event", http.StatusInternalServerError)
		return
	}

	h.logger.Info("event updated", "id", id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(h.eventToAPI(event)); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *CalendarHandler) DeleteEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()

	if err := h.storage.DeleteEvent(ctx, id.String()); err != nil {
		if errors.Is(err, storage.ErrEventNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to delete event", "error", err)
		http.Error(w, "failed to delete event", http.StatusInternalServerError)
		return
	}

	h.logger.Info("event deleted", "id", id)

	w.WriteHeader(http.StatusNoContent)
}

func (h *CalendarHandler) eventToAPI(event storage.Event) api.Event {
	return api.Event{
		ID:           event.ID,
		Title:        event.Title,
		StartTime:    event.StartTime,
		EndTime:      event.EndTime,
		Description:  event.Description,
		UserID:       event.UserID,
		NotifyBefore: int64(event.NotifyBefore),
	}
}
