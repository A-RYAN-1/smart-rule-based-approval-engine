package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/app/auto_reject"
	"github.com/ankita-advitot/rule_based_approval_engine/app/auto_reject/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAutoRejectService_AutoRejectLeaveRequests(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		mockSetup     func(l *mocks.LeaveRequestRepository, e *mocks.ExpenseRequestRepository, d *mocks.DiscountRequestRepository, h *mocks.HolidayRepository, db *mocks.DB, tx *mocks.Tx)
		expectedError error
	}{
		{
			name: "Success - Reject Expired Requests in all repos",
			mockSetup: func(l *mocks.LeaveRequestRepository, e *mocks.ExpenseRequestRepository, d *mocks.DiscountRequestRepository, h *mocks.HolidayRepository, db *mocks.DB, tx *mocks.Tx) {
				pastDate := time.Now().AddDate(0, 0, -10)

				// Leave Repo
				l.EXPECT().GetPendingRequests(ctx).Return([]struct {
					ID        int64
					CreatedAt time.Time
				}{{ID: 1, CreatedAt: pastDate}}, nil)

				// Expense Repo
				e.EXPECT().GetPendingRequests(ctx).Return([]struct {
					ID        int64
					CreatedAt time.Time
				}{{ID: 10, CreatedAt: pastDate}}, nil)

				// Discount Repo
				d.EXPECT().GetPendingRequests(ctx).Return([]struct {
					ID        int64
					CreatedAt time.Time
				}{{ID: 20, CreatedAt: pastDate}}, nil)

				h.EXPECT().IsHoliday(ctx, mock.Anything).Return(false, nil).Maybe()

				// Expectations for all three rejections
				db.EXPECT().Begin(ctx).Return(tx, nil).Times(3)
				l.EXPECT().UpdateStatus(ctx, tx, int64(1), "AUTO_REJECTED", int64(0), mock.Anything).Return(nil)
				e.EXPECT().UpdateStatus(ctx, tx, int64(10), "AUTO_REJECTED", int64(0), mock.Anything).Return(nil)
				d.EXPECT().UpdateStatus(ctx, tx, int64(20), "AUTO_REJECTED", int64(0), mock.Anything).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil).Times(3)
			},
		},
		{
			name: "Fail - Leave Repo Fetch Error",
			mockSetup: func(l *mocks.LeaveRequestRepository, e *mocks.ExpenseRequestRepository, d *mocks.DiscountRequestRepository, h *mocks.HolidayRepository, db *mocks.DB, tx *mocks.Tx) {
				l.EXPECT().GetPendingRequests(ctx).Return(nil, assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name: "Success - No Expired Requests",
			mockSetup: func(l *mocks.LeaveRequestRepository, e *mocks.ExpenseRequestRepository, d *mocks.DiscountRequestRepository, h *mocks.HolidayRepository, db *mocks.DB, tx *mocks.Tx) {
				now := time.Now()
				l.EXPECT().GetPendingRequests(ctx).Return([]struct {
					ID        int64
					CreatedAt time.Time
				}{{ID: 1, CreatedAt: now}}, nil)
				e.EXPECT().GetPendingRequests(ctx).Return(nil, nil)
				d.EXPECT().GetPendingRequests(ctx).Return(nil, nil)
				h.EXPECT().IsHoliday(ctx, mock.Anything).Return(false, nil).Maybe()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLeaveRepo := mocks.NewLeaveRequestRepository(t)
			mockExpenseRepo := mocks.NewExpenseRequestRepository(t)
			mockDiscountRepo := mocks.NewDiscountRequestRepository(t)
			mockHolidayRepo := mocks.NewHolidayRepository(t)
			mockDB := mocks.NewDB(t)
			mockTx := mocks.NewTx(t)

			tt.mockSetup(mockLeaveRepo, mockExpenseRepo, mockDiscountRepo, mockHolidayRepo, mockDB, mockTx)

			service := auto_reject.NewAutoRejectService(ctx, mockLeaveRepo, mockExpenseRepo, mockDiscountRepo, mockHolidayRepo, mockDB)
			err := service.AutoRejectExpiredRequests(ctx)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
