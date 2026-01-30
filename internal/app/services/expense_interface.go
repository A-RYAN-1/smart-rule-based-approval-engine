package services

import (
	"context"
)

type ExpenseServiceInterface interface {
	ApplyExpense(ctx context.Context, userID int64, amount float64, category string, reason string) (string, string, error)
	CancelExpense(ctx context.Context, userID, requestID int64) error
}
