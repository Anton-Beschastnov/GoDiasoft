package storage

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("date is already busy with another event")
	ErrInvalidEvent  = errors.New("invalid event data")
)

type Event struct {
	ID           string
	Title        string
	StartTime    time.Time
	EndTime      time.Time
	Description  string
	UserID       string
	NotifyBefore time.Duration
}
