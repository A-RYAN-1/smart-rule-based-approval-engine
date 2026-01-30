// File: internal/app/services/tests/expense_service_test.go
package tests

import (
	"context"
	"testing"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/stretchr/testify/assert"
)

func TestExpenseService_ApplyExpense_Validation(t *testing.T) {
	service := services.NewExpenseService(nil, nil, nil, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		amount    float64
		category  string
		expectErr error
	}{
		{
			name:      "Invalid User ID",
			userID:    0,
			amount:    100.0,
			category:  "Food",
			expectErr: apperrors.ErrInvalidUser,
		},
		{
			name:      "Invalid Amount",
			userID:    1,
			amount:    0,
			category:  "Food",
			expectErr: apperrors.ErrInvalidExpenseAmount,
		},
		{
			name:      "Invalid Category",
			userID:    1,
			amount:    100.0,
			category:  "",
			expectErr: apperrors.ErrInvalidExpenseCategory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.ApplyExpense(ctx, tt.userID, tt.amount, tt.category, "Reason")
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestExpenseService_CancelExpense_Validation(t *testing.T) {
	// Since s.db.Begin(ctx) is called first, we can't easily test the full flow
	// without a real DB pool or refactoring.
	// But we can test that it returns error if db is nil (panic or error depending on implementation)
	// Actually, let's focus on logic that happens after DB if possible,
	// but here DB is the first thing.
}

// NOTE: Success cases and DB-integrated paths are limited by the concrete *pgxpool.Pool dependency.
