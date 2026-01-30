package services

import (
	"context"
)

type LeaveApprovalServiceInterface interface {
	GetPendingLeaveRequests(ctx context.Context, role string, approverID int64) ([]map[string]interface{}, error)
	ApproveLeave(ctx context.Context, role string, approverID, requestID int64, approvalComment string) error
	RejectLeave(ctx context.Context, role string, approverID, requestID int64, rejectionComment string) error
}
