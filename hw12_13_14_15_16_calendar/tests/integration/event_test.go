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

// Event mirrors api.Event from the server.
type Event struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	UserID       string    `json:"user_id"`
	NotifyBefore int64     `json:"notify_before"` // nanoseconds
}

// CreateEventRequest mirrors api.CreateEventRequest.
type CreateEventRequest struct {
	Title        string    `json:"title"`
	Description  *string   `json:"description,omitempty"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	UserID       string    `json:"user_id"`
	NotifyBefore *int64    `json:"notify_before,omitempty"`
}

// Notification represents the row stored in the notifications table by the storer.
type Notification struct {
	ID        string
	EventID   string
	Title     string
	StartTime time.Time
	UserID    string
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

// waitForCalendar blocks until the calendar HTTP API responds or times out.
func waitForCalendar(t *testing.T, apiURL string) {
	t.Helper()
	require.Eventually(t, func() bool {
		targetURL := apiURL + "/events?user_id=health&start_date=" + time.Now().Format(time.RFC3339)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, targetURL, nil)
		if err != nil {
			return false
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}
		resp.Body.Close()
		// Any HTTP response (including 400 for bad params) means server is up.
		return true
	}, 60*time.Second, 2*time.Second, "calendar API did not become ready")
}

// TestEventHappyPath tests creating an event and fetching it by day/week/month listing.
func TestEventHappyPath(t *testing.T) {
	apiURL := getCalendarAPIURL()
	waitForCalendar(t, apiURL)

	const testUserID = "integration-test-user-1"

	startTime := time.Now().Add(1 * time.Hour).UTC().Truncate(time.Second)
	endTime := startTime.Add(1 * time.Hour)

	reqBody := CreateEventRequest{
		Title:     "Important Meeting",
		StartTime: startTime,
		EndTime:   endTime,
		UserID:    testUserID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, apiURL+"/events", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdEvent Event
	err = json.NewDecoder(resp.Body).Decode(&createdEvent)
	require.NoError(t, err)
	resp.Body.Close()

	require.NotEmpty(t, createdEvent.ID)
	require.Equal(t, reqBody.Title, createdEvent.Title)
	require.Equal(t, startTime, createdEvent.StartTime)

	t.Run("list by day", func(t *testing.T) {
		url := fmt.Sprintf("%s/events?user_id=%s&start_date=%s&period=day",
			apiURL, testUserID, startTime.Format(time.RFC3339))
		getReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		require.NoError(t, err)
		getResp, err := http.DefaultClient.Do(getReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, getResp.StatusCode)

		var events []Event
		err = json.NewDecoder(getResp.Body).Decode(&events)
		require.NoError(t, err)
		getResp.Body.Close()

		require.GreaterOrEqual(t, len(events), 1)
		found := containsEvent(events, createdEvent.ID)
		require.True(t, found, "created event not found in day listing")
	})

	t.Run("list by week", func(t *testing.T) {
		url := fmt.Sprintf("%s/events?user_id=%s&start_date=%s&period=week",
			apiURL, testUserID, startTime.Format(time.RFC3339))

		getReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		require.NoError(t, err)

		getResp, err := http.DefaultClient.Do(getReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, getResp.StatusCode)

		var events []Event
		err = json.NewDecoder(getResp.Body).Decode(&events)
		require.NoError(t, err)
		getResp.Body.Close()

		require.GreaterOrEqual(t, len(events), 1)
		require.True(t, containsEvent(events, createdEvent.ID), "created event not found in week listing")
	})

	t.Run("list by month", func(t *testing.T) {
		url := fmt.Sprintf("%s/events?user_id=%s&start_date=%s&period=month",
			apiURL, testUserID, startTime.Format(time.RFC3339))
		getReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		require.NoError(t, err)

		getResp, err := http.DefaultClient.Do(getReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, getResp.StatusCode)

		var events []Event
		err = json.NewDecoder(getResp.Body).Decode(&events)
		require.NoError(t, err)
		getResp.Body.Close()

		require.GreaterOrEqual(t, len(events), 1)
		require.True(t, containsEvent(events, createdEvent.ID), "created event not found in month listing")
	})
}

// containsEvent returns true if the event with the given ID is in the list.
func containsEvent(events []Event, id string) bool {
	for _, e := range events {
		if e.ID == id {
			return true
		}
	}
	return false
}

// TestAPIErrorHandling tests that the API returns proper error codes for invalid input.
func TestAPIErrorHandling(t *testing.T) {
	apiURL := getCalendarAPIURL()
	waitForCalendar(t, apiURL)

	now := time.Now().UTC()

	testCases := []struct {
		name         string
		req          CreateEventRequest
		expectedCode int
	}{
		{
			name: "End time before start time",
			req: CreateEventRequest{
				Title:     "Invalid Event",
				UserID:    "test-user-err",
				StartTime: now.Add(2 * time.Hour),
				EndTime:   now.Add(1 * time.Hour),
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing title",
			req: CreateEventRequest{
				UserID:    "test-user-err",
				StartTime: now.Add(1 * time.Hour),
				EndTime:   now.Add(2 * time.Hour),
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing user_id",
			req: CreateEventRequest{
				Title:     "Event Without User",
				StartTime: now.Add(1 * time.Hour),
				EndTime:   now.Add(2 * time.Hour),
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tc.req)
			require.NoError(t, err)

			postReq, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				apiURL+"/events",
				bytes.NewBuffer(bodyBytes),
			)
			require.NoError(t, err)
			postReq.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(postReq)
			require.NoError(t, err)
			resp.Body.Close()
			require.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

// TestEndToEndNotificationFlow creates an event with a short notify_before,
// waits for the scheduler to send a notification to Kafka, and then checks
// that the storer has persisted it to the notifications table.
func TestEndToEndNotificationFlow(t *testing.T) {
	ctx := context.Background()
	apiURL := getCalendarAPIURL()
	waitForCalendar(t, apiURL)

	dbURL := getDBConnectionString()
	pool, err := pgxpool.Connect(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	err = pool.Ping(ctx)
	require.NoError(t, err, "failed to connect to database")

	const testUserID = "integration-test-user-e2e"

	// Планировщик в интеграционных тестах сканирует каждые 10 секунд (scheduler_config_integration.yaml).
	// Создаём событие с notify_before=30s и start_time=1 минута от текущего момента,
	// чтобы планировщик гарантированно его поймал.
	startTime := time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	notifyBefore := int64(30 * time.Second) // 30 секунд в наносекундах

	reqBody := CreateEventRequest{
		Title:        "E2E Test Event",
		UserID:       testUserID,
		StartTime:    startTime,
		EndTime:      startTime.Add(30 * time.Minute),
		NotifyBefore: &notifyBefore,
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	postReq, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		apiURL+"/events",
		bytes.NewBuffer(bodyBytes),
	)
	require.NoError(t, err)
	postReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(postReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdEvent Event
	err = json.NewDecoder(resp.Body).Decode(&createdEvent)
	require.NoError(t, err)
	resp.Body.Close()

	require.NotEmpty(t, createdEvent.ID)
	t.Logf("Created event with ID: %s, waiting for notification to appear in DB...", createdEvent.ID)

	// Ждём максимум 90 секунд — планировщик сканирует каждые 10 секунд.
	var notification Notification
	require.Eventually(t, func() bool {
		// The notifications table schema: id (UUID), event_id, title, start_time, user_id, created_at
		query := `SELECT id, event_id, title, start_time, user_id, created_at
		          FROM notifications WHERE event_id = $1`
		row := pool.QueryRow(ctx, query, createdEvent.ID)
		err := row.Scan(
			&notification.ID,
			&notification.EventID,
			&notification.Title,
			&notification.StartTime,
			&notification.UserID,
			&notification.CreatedAt,
		)
		if err != nil {
			t.Logf("notification not yet in DB: %v", err)
			return false
		}
		return true
	}, 90*time.Second, 5*time.Second, "notification did not appear in the database within 90 seconds")

	require.Equal(t, createdEvent.ID, notification.EventID)
	require.Equal(t, createdEvent.Title, notification.Title)
	require.Equal(t, testUserID, notification.UserID)
}
