// File: internal/app/services/tests/discount_service_test.go
package tests

import (
	"context"
	"testing"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/stretchr/testify/assert"
)

func TestDiscountService_ApplyDiscount_Validation(t *testing.T) {
	service := services.NewDiscountService(nil, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		percent   float64
		expectErr error
	}{
		{
			name:      "Invalid User ID",
			userID:    0,
			percent:   10.0,
			expectErr: apperrors.ErrInvalidUser,
		},
		{
			name:      "Invalid Percent",
			userID:    1,
			percent:   0,
			expectErr: apperrors.ErrInvalidDiscountPercent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.ApplyDiscount(ctx, tt.userID, tt.percent, "Reason")
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestDiscountService_ApproveDiscount_Validation(t *testing.T) {
	service := services.NewDiscountService(nil, nil, nil)
	ctx := context.Background()

	t.Run("Employee Cannot Approve", func(t *testing.T) {
		err := service.ApproveDiscount(ctx, "EMPLOYEE", 1, 1, "Comment")
		assert.Equal(t, apperrors.ErrEmployeeCannotApprove, err)
	})

	t.Run("Comment Required", func(t *testing.T) {
		err := service.ApproveDiscount(ctx, "ADMIN", 1, 1, "")
		assert.Equal(t, apperrors.ErrCommentRequired, err)
	})
}
