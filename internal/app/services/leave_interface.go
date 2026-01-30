package services

import (
	"context"
	"time"
)

// LeaveServiceInterface defines the interface for LeaveService to allow mocking
type LeaveServiceInterface interface {
	ApplyLeave(ctx context.Context, userID int64, from time.Time, to time.Time, days int, leaveType string, reason string) (string, string, error)
	CancelLeave(ctx context.Context, userID, requestID int64) error
}
