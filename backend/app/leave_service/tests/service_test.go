package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/app/leave_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/leave_service/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLeaveService_ApplyLeave(t *testing.T) {
	ctx := context.Background()
	tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
	dayAfter := tomorrow.AddDate(0, 0, 1)

	tests := []struct {
		name          string
		userID        int64
		from          time.Time
		to            time.Time
		days          int
		leaveType     string
		reason        string
		mockSetup     func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx)
		expectedError error
	}{
		{
			name:      "Success - Auto Approved",
			userID:    1,
			from:      tomorrow,
			to:        dayAfter,
			days:      2,
			leaveType: "SICK",
			reason:    "Feeling unwell",
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				l.EXPECT().CheckOverlap(ctx, int64(1), tomorrow, dayAfter).Return(false, nil)
				db.EXPECT().Begin(ctx).Return(tx, nil)
				b.EXPECT().GetLeaveBalance(ctx, tx, int64(1)).Return(10, nil)
				u.EXPECT().GetGrade(ctx, tx, int64(1)).Return(int64(1), nil)
				r.EXPECT().GetRule(ctx, "LEAVE", int64(1)).Return(&models.Rule{
					ID:        1,
					Condition: map[string]interface{}{"max_days": 5.0},
				}, nil)
				l.EXPECT().Create(ctx, tx, mock.AnythingOfType("*models.LeaveRequest")).Return(nil)
				b.EXPECT().DeductLeaveBalance(ctx, tx, int64(1), 2).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Maybe()
			},
			expectedError: nil,
		},
		{
			name:      "Leave Balance Exceeded",
			userID:    1,
			from:      tomorrow,
			to:        dayAfter,
			days:      20,
			leaveType: "SICK",
			reason:    "Long trip",
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				l.EXPECT().CheckOverlap(ctx, int64(1), tomorrow, dayAfter).Return(false, nil)
				db.EXPECT().Begin(ctx).Return(tx, nil)
				b.EXPECT().GetLeaveBalance(ctx, tx, int64(1)).Return(10, nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Once()
			},
			expectedError: apperrors.ErrLeaveBalanceExceeded,
		},
		{
			name:   "Invalid User",
			userID: 0,
			from:   tomorrow,
			to:     dayAfter,
			days:   2,
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
			},
			expectedError: apperrors.ErrInvalidUser,
		},
		{
			name:      "Overlap Detected",
			userID:    1,
			from:      tomorrow,
			to:        dayAfter,
			days:      2,
			leaveType: "SICK",
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				l.EXPECT().CheckOverlap(ctx, int64(1), tomorrow, dayAfter).Return(true, nil)
			},
			expectedError: apperrors.ErrLeaveOverlap,
		},
		{
			name:      "Success - Manual Approval Pending",
			userID:    1,
			from:      tomorrow,
			to:        dayAfter,
			days:      2,
			leaveType: "SICK",
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				l.EXPECT().CheckOverlap(ctx, int64(1), tomorrow, dayAfter).Return(false, nil)
				db.EXPECT().Begin(ctx).Return(tx, nil)
				b.EXPECT().GetLeaveBalance(ctx, tx, int64(1)).Return(10, nil)
				u.EXPECT().GetGrade(ctx, tx, int64(1)).Return(int64(1), nil)
				r.EXPECT().GetRule(ctx, "LEAVE", int64(1)).Return(&models.Rule{
					ID:        1,
					Condition: map[string]interface{}{"max_days": 1.0}, // max_days < days
				}, nil)
				l.EXPECT().Create(ctx, tx, mock.Anything).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Maybe()
			},
			expectedError: nil,
		},
		{
			name:      "DB Begin Error",
			userID:    1,
			from:      tomorrow,
			to:        dayAfter,
			days:      2,
			leaveType: "SICK",
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				l.EXPECT().CheckOverlap(ctx, int64(1), tomorrow, dayAfter).Return(false, nil)
				db.EXPECT().Begin(ctx).Return(nil, apperrors.ErrTransactionBegin)
			},
			expectedError: apperrors.ErrTransactionBegin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockL := mocks.NewLeaveRequestRepository(t)
			mockB := mocks.NewBalanceRepository(t)
			mockR := mocks.NewRuleService(t)
			mockU := mocks.NewUserRepository(t)
			mockDB := mocks.NewDB(t)
			mockTx := mocks.NewTx(t)

			tt.mockSetup(mockL, mockB, mockR, mockU, mockDB, mockTx)

			service := leave_service.NewLeaveService(ctx, mockL, mockB, mockR, mockU, mockDB)
			_, _, err := service.ApplyLeave(ctx, tt.userID, tt.from, tt.to, tt.days, tt.leaveType, tt.reason)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLeaveApprovalService_ApproveLeave(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		role            string
		approverID      int64
		requestID       int64
		approvalComment string
		mockSetup       func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx)
		expectedError   error
	}{
		{
			name:            "Success - Manager Approves",
			role:            constants.RoleManager,
			approverID:      2,
			requestID:       10,
			approvalComment: "Approved by manager",
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				l.EXPECT().GetByID(ctx, tx, int64(10)).Return(&models.LeaveRequest{
					ID:         10,
					EmployeeID: 1,
					Status:     constants.StatusPending,
					FromDate:   time.Now(),
					ToDate:     time.Now().AddDate(0, 0, 1),
				}, nil)
				u.EXPECT().GetRole(ctx, tx, int64(1)).Return(constants.RoleEmployee, nil)
				b.EXPECT().DeductLeaveBalance(ctx, tx, int64(1), 2).Return(nil)
				l.EXPECT().UpdateStatus(ctx, tx, int64(10), "APPROVED", int64(2), "Approved by manager").Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Maybe()
			},
			expectedError: nil,
		},
		{
			name:            "Self Approval Not Allowed",
			role:            constants.RoleManager,
			approverID:      1,
			requestID:       10,
			approvalComment: "Self approve",
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				l.EXPECT().GetByID(ctx, tx, int64(10)).Return(&models.LeaveRequest{
					ID:         10,
					EmployeeID: 1,
					Status:     constants.StatusPending,
				}, nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Once()
			},
			expectedError: apperrors.ErrSelfApprovalNotAllowed,
		},
		{
			name:            "Repository Error - GetByID",
			role:            constants.RoleManager,
			approverID:      2,
			requestID:       10,
			approvalComment: "OK",
			mockSetup: func(l *mocks.LeaveRequestRepository, b *mocks.BalanceRepository, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				l.EXPECT().GetByID(ctx, tx, int64(10)).Return(nil, apperrors.ErrDatabase)
				tx.EXPECT().Rollback(ctx).Return(nil).Once()
			},
			expectedError: apperrors.ErrDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockL := mocks.NewLeaveRequestRepository(t)
			mockB := mocks.NewBalanceRepository(t)
			mockU := mocks.NewUserRepository(t)
			mockDB := mocks.NewDB(t)
			mockTx := mocks.NewTx(t)

			tt.mockSetup(mockL, mockB, mockU, mockDB, mockTx)

			service := leave_service.NewLeaveApprovalService(ctx, mockL, mockB, mockU, mockDB)
			err := service.ApproveLeave(ctx, tt.role, tt.approverID, tt.requestID, tt.approvalComment)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLeaveApprovalService_RejectLeave(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockL := mocks.NewLeaveRequestRepository(t)
		mockU := mocks.NewUserRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockL.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(&models.LeaveRequest{
			ID:         10,
			EmployeeID: 1,
			Status:     constants.StatusPending,
		}, nil)
		mockU.EXPECT().GetRole(ctx, mockTx, int64(1)).Return(constants.RoleEmployee, nil)
		mockL.EXPECT().UpdateStatus(ctx, mockTx, int64(10), "REJECTED", int64(2), "No").Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil)
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := leave_service.NewLeaveApprovalService(ctx, mockL, nil, mockU, mockDB)
		err := service.RejectLeave(ctx, constants.RoleManager, 2, 10, "No")

		assert.NoError(t, err)
	})

	t.Run("Unauthorized Role", func(t *testing.T) {
		service := leave_service.NewLeaveApprovalService(ctx, nil, nil, nil, nil)
		err := service.RejectLeave(ctx, constants.RoleEmployee, 1, 10, "No")

		assert.ErrorIs(t, err, apperrors.ErrEmployeeCannotApprove)
	})
}

func TestLeaveService_CancelLeave(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockL := mocks.NewLeaveRequestRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockL.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(&models.LeaveRequest{
			ID:         10,
			EmployeeID: 1,
			Status:     constants.StatusPending,
		}, nil)
		mockL.EXPECT().Cancel(ctx, mockTx, int64(10)).Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil)
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := leave_service.NewLeaveService(ctx, mockL, nil, nil, nil, mockDB)
		err := service.CancelLeave(ctx, 1, 10)

		assert.NoError(t, err)
	})

	t.Run("Not Authorized Owner", func(t *testing.T) {
		mockL := mocks.NewLeaveRequestRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockL.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(&models.LeaveRequest{
			ID:         10,
			EmployeeID: 2, // Different owner
			Status:     constants.StatusPending,
		}, nil)
		mockTx.EXPECT().Rollback(ctx).Return(nil).Once()

		service := leave_service.NewLeaveService(ctx, mockL, nil, nil, nil, mockDB)
		err := service.CancelLeave(ctx, 1, 10)

		assert.ErrorIs(t, err, apperrors.ErrLeaveRequestNotFound)
	})
}

func TestLeaveApprovalService_GetPendingLeaveRequests(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Manager", func(t *testing.T) {
		mockL := mocks.NewLeaveRequestRepository(t)
		mockL.EXPECT().GetPendingForManager(ctx, int64(2)).Return([]map[string]interface{}{}, nil)

		service := leave_service.NewLeaveApprovalService(ctx, mockL, nil, nil, nil)
		_, err := service.GetPendingLeaveRequests(ctx, constants.RoleManager, 2)

		assert.NoError(t, err)
	})

	t.Run("Unauthorized Role", func(t *testing.T) {
		service := leave_service.NewLeaveApprovalService(ctx, nil, nil, nil, nil)
		_, err := service.GetPendingLeaveRequests(ctx, constants.RoleEmployee, 1)

		assert.ErrorIs(t, err, apperrors.ErrUnauthorizedRole)
	})
}
