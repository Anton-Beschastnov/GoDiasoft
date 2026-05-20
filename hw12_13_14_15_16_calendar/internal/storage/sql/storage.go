package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib" // blank import to register pgx driver
)

type Storage struct {
	db  *sql.DB
	dsn string
}

func New(dsn string) *Storage {
	return &Storage{dsn: dsn}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.db = db
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	query := `
		INSERT INTO events (id, title, start_time, end_time, description, user_id, notify_before)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.Title,
		event.StartTime,
		event.EndTime,
		event.Description,
		event.UserID,
		event.NotifyBefore,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	query := `
		UPDATE events 
		SET title = $2, start_time = $3, end_time = $4, description = $5, notify_before = $6
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.Title,
		event.StartTime,
		event.EndTime,
		event.Description,
		event.NotifyBefore,
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	query := `
		SELECT id, title, start_time, end_time, description, user_id, notify_before
		FROM events WHERE id = $1
	`

	var event storage.Event
	var notifyBefore int64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Title,
		&event.StartTime,
		&event.EndTime,
		&event.Description,
		&event.UserID,
		&notifyBefore,
	)
	if err == sql.ErrNoRows {
		return storage.Event{}, storage.ErrEventNotFound
	}
	if err != nil {
		return storage.Event{}, fmt.Errorf("failed to get event: %w", err)
	}

	event.NotifyBefore = time.Duration(notifyBefore)
	return event, nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, userID string, date time.Time) ([]storage.Event, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.listEvents(ctx, userID, startOfDay, endOfDay)
}

func (s *Storage) ListEventsForWeek(ctx context.Context, userID string, startDate time.Time) ([]storage.Event, error) {
	startOfWeek := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return s.listEvents(ctx, userID, startOfWeek, endOfWeek)
}

func (s *Storage) ListEventsForMonth(ctx context.Context, userID string, startDate time.Time) ([]storage.Event, error) {
	startOfMonth := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return s.listEvents(ctx, userID, startOfMonth, endOfMonth)
}

func (s *Storage) listEvents(ctx context.Context, userID string, start, end time.Time) ([]storage.Event, error) {
	query := `
		SELECT id, title, start_time, end_time, description, user_id, notify_before
		FROM events 
		WHERE user_id = $1 AND start_time >= $2 AND start_time < $3
		ORDER BY start_time
	`

	rows, err := s.db.QueryContext(ctx, query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var event storage.Event
		var notifyBefore int64

		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.StartTime,
			&event.EndTime,
			&event.Description,
			&event.UserID,
			&notifyBefore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		event.NotifyBefore = time.Duration(notifyBefore)
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}
