package services

import "context"

type DiscountApprovalServiceInterface interface {
	GetPendingDiscountRequests(ctx context.Context, role string, approverID int64) ([]map[string]interface{}, error)
	ApproveDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error
	RejectDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error
}
