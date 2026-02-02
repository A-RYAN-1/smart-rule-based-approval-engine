// test cases for leave service
package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"rule-based-approval-engine/internal/app/repositories/mocks"
	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/stretchr/testify/assert"
)

func TestLeaveService_ApplyLeave_Validation(t *testing.T) {
	service := services.NewLeaveService(nil, nil, nil, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		from      time.Time
		to        time.Time
		days      int
		expectErr error
	}{
		{
			name:      "Invalid User ID",
			userID:    0,
			from:      time.Now().Add(24 * time.Hour),
			to:        time.Now().Add(48 * time.Hour),
			days:      1,
			expectErr: apperrors.ErrInvalidUser,
		},
		{
			name:      "Invalid Days",
			userID:    1,
			from:      time.Now().Add(24 * time.Hour),
			to:        time.Now().Add(48 * time.Hour),
			days:      0,
			expectErr: apperrors.ErrInvalidLeaveDays,
		},
		{
			name:      "Invalid Date Range",
			userID:    1,
			from:      time.Now().Add(48 * time.Hour),
			to:        time.Now().Add(24 * time.Hour),
			days:      1,
			expectErr: apperrors.ErrInvalidDateRange,
		},
		{
			name:      "Past Date",
			userID:    1,
			from:      time.Now().Add(-24 * time.Hour),
			to:        time.Now().Add(24 * time.Hour),
			days:      1,
			expectErr: apperrors.ErrPastDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.ApplyLeave(ctx, tt.userID, tt.from, tt.to, tt.days, "SICK", "Reason")
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestLeaveService_ApplyLeave_Overlap(t *testing.T) {
	mockLeaveRepo := mocks.NewLeaveRequestRepository(t)
	service := services.NewLeaveService(mockLeaveRepo, nil, nil, nil, nil)
	ctx := context.Background()
	userID := int64(1)
	from := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
	to := from.Add(24 * time.Hour)

	t.Run("Existing Overlap", func(t *testing.T) {
		mockLeaveRepo.On("CheckOverlap", ctx, userID, from, to).Return(true, nil).Once()
		_, _, err := service.ApplyLeave(ctx, userID, from, to, 1, "SICK", "Reason")
		assert.Equal(t, apperrors.ErrLeaveOverlap, err)
	})

	t.Run("Overlap Check Error", func(t *testing.T) {
		mockLeaveRepo.On("CheckOverlap", ctx, userID, from, to).Return(false, errors.New("db error")).Once()
		_, _, err := service.ApplyLeave(ctx, userID, from, to, 1, "SICK", "Reason")
		assert.Equal(t, apperrors.ErrLeaveVerificationFailed, err)
	})
}
