package services

import (
	"context"
)

type ExpenseApprovalServiceInterface interface {
	GetPendingExpenseRequests(ctx context.Context, role string, approverID int64) ([]map[string]interface{}, error)
	ApproveExpense(ctx context.Context, role string, approverID, requestID int64, comment string) error
	RejectExpense(ctx context.Context, role string, approverID, requestID int64, comment string) error
}
