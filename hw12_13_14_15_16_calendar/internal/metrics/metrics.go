package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	EventsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "events_created_total",
			Help: "Total number of created events.",
		},
	)

	EventsUpdated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "events_updated_total",
			Help: "Total number of updated events.",
		},
	)

	EventsDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "events_deleted_total",
			Help: "Total number of deleted events.",
		},
	)

	NotificationsSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_sent_total",
			Help: "Total number of sent notifications.",
		},
		[]string{"status"}, // "success" or "error"
	)

	SchedulerRuns = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_runs_total",
			Help: "Total number of scheduler runs.",
		},
		[]string{"status"}, // "success" or "error"
	)

	NotificationsSaved = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_saved_total",
			Help: "Total number of saved notifications in storer.",
		},
		[]string{"status"}, // "success" or "error"
	)
)
