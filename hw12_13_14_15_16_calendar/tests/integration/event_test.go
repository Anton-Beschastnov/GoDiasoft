package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
)

type Event struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	NotifyBefore time.Duration `json:"notify_before"`
}

type Notification struct {
	ID        int
	EventID   int
	Message   string
	CreatedAt time.Time
}

func getCalendarAPIURL() string {
	url := os.Getenv("CALENDAR_API_URL")
	if url == "" {
		return "http://localhost:8889"
	}
	return url
}

func getDBConnectionString() string {
	connStr := os.Getenv("TEST_DB_STRING")
	if connStr == "" {
		return "postgres://calendar:calendar@postgres:5432/calendar?sslmode=disable"
	}
	return connStr
}

func TestEventHappyPath(t *testing.T) {
	apiURL := getCalendarAPIURL()

	startTime := time.Now().Add(1 * time.Hour).UTC().Truncate(time.Second)
	endTime := startTime.Add(1 * time.Hour).UTC().Truncate(time.Second)

	event := Event{
		Title:       "Important Meeting",
		Description: "A very important meeting.",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	bodyBytes, err := json.Marshal(event)
	require.NoError(t, err)

	resp, err := http.Post(apiURL+"/events", "application/json", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdEvent Event
	err = json.NewDecoder(resp.Body).Decode(&createdEvent)
	require.NoError(t, err)
	resp.Body.Close()

	require.NotZero(t, createdEvent.ID)
	require.Equal(t, event.Title, createdEvent.Title)
	require.Equal(t, event.StartTime, createdEvent.StartTime)

	resp, err = http.Get(fmt.Sprintf("%s/events/day?time=%s", apiURL, startTime.Format(time.RFC3339)))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var events []Event
	err = json.NewDecoder(resp.Body).Decode(&events)
	require.NoError(t, err)
	resp.Body.Close()

	require.GreaterOrEqual(t, len(events), 1)
	var found bool
	for _, e := range events {
		if e.ID == createdEvent.ID {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestAPIErrorHandling(t *testing.T) {
	apiURL := getCalendarAPIURL()

	testCases := []struct {
		name          string
		event         Event
		expectedCode  int
	}{
		{
			name: "End time before start time",
			event: Event{
				Title:     "Invalid Event",
				StartTime: time.Now().Add(2 * time.Hour),
				EndTime:   time.Now().Add(1 * time.Hour),
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing title",
			event: Event{
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tc.event)
			require.NoError(t, err)

			resp, err := http.Post(apiURL+"/events", "application/json", bytes.NewBuffer(bodyBytes))
			require.NoError(t, err)
			resp.Body.Close()
			require.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

func TestEndToEndNotificationFlow(t *testing.T) {
	ctx := context.Background()
	apiURL := getCalendarAPIURL()
	dbURL := getDBConnectionString()

	pool, err := pgxpool.Connect(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	err = pool.Ping(ctx)
	require.NoError(t, err, "failed to connect to database")

	startTime := time.Now().Add(5 * time.Second).UTC().Truncate(time.Second)
	event := Event{
		Title:       "E2E Test Event",
		Description: "Test for scheduler and storer",
		StartTime:   startTime,
		EndTime:     startTime.Add(30 * time.Minute),
		NotifyBefore: 5 * time.Second,
	}

	bodyBytes, err := json.Marshal(event)
	require.NoError(t, err)

	resp, err := http.Post(apiURL+"/events", "application/json", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdEvent Event
	err = json.NewDecoder(resp.Body).Decode(&createdEvent)
	require.NoError(t, err)
	resp.Body.Close()

	require.NotZero(t, createdEvent.ID)

	var notification Notification
	require.Eventually(t, func() bool {
		query := `SELECT id, event_id, message, created_at FROM notifications WHERE event_id = $1`
		row := pool.QueryRow(ctx, query, createdEvent.ID)
		err := row.Scan(&notification.ID, &notification.EventID, &notification.Message, &notification.CreatedAt)
		return err == nil
	}, 90*time.Second, 5*time.Second, "notification did not appear in the database")

	require.Equal(t, createdEvent.ID, notification.EventID)
	require.Contains(t, notification.Message, createdEvent.Title)
}
