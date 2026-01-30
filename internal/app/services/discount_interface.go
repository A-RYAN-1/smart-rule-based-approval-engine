package services

import (
	"context"
)

type DiscountServiceInterface interface {
	ApplyDiscount(ctx context.Context, userID int64, percent float64, reason string) (string, string, error)
	CancelDiscount(ctx context.Context, userID, requestID int64) error
	GetPendingDiscountRequests(ctx context.Context, role string, approverID int64) ([]map[string]interface{}, error)
	ApproveDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error
	RejectDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error
}
