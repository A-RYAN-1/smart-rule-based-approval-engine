// File: internal/app/services/tests/expense_service_test.go
//
//	test cases for apply expense service
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
