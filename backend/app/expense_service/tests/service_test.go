package tests

import (
	"context"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/expense_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/expense_service/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExpenseService_ApplyExpense(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        int64
		amount        float64
		category      string
		reason        string
		mockSetup     func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx)
		expectedError error
	}{
		{
			name:     "Success - Auto Approved",
			userID:   1,
			amount:   50.0,
			category: "TRAVEL",
			reason:   "Client meeting",
			mockSetup: func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				b.EXPECT().GetExpenseBalance(ctx, tx, int64(1)).Return(1000.0, nil)
				u.EXPECT().GetGrade(ctx, tx, int64(1)).Return(int64(1), nil)
				r.EXPECT().GetRule(ctx, "EXPENSE", int64(1)).Return(&models.Rule{
					ID:        1,
					Condition: map[string]interface{}{"max_amount": 100.0},
				}, nil)
				e.EXPECT().Create(ctx, tx, mock.AnythingOfType("*models.ExpenseRequest")).Return(nil)
				b.EXPECT().DeductExpenseBalance(ctx, tx, int64(1), 50.0).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Maybe()
			},
			expectedError: nil,
		},
		{
			name:     "Expense Limit Exceeded",
			userID:   1,
			amount:   2000.0,
			category: "EQUIPMENT",
			mockSetup: func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				b.EXPECT().GetExpenseBalance(ctx, tx, int64(1)).Return(1000.0, nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Once()
			},
			expectedError: apperrors.ErrExpenseLimitExceeded,
		},
		{
			name:     "Invalid Amount",
			userID:   1,
			amount:   0,
			category: "FOOD",
			mockSetup: func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
			},
			expectedError: apperrors.ErrInvalidExpenseAmount,
		},
		{
			name:     "Success - Manual Approval Pending",
			userID:   1,
			amount:   500.0,
			category: "TRAVEL",
			mockSetup: func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				b.EXPECT().GetExpenseBalance(ctx, tx, int64(1)).Return(1000.0, nil)
				u.EXPECT().GetGrade(ctx, tx, int64(1)).Return(int64(1), nil)
				r.EXPECT().GetRule(ctx, "EXPENSE", int64(1)).Return(&models.Rule{
					ID:        1,
					Condition: map[string]interface{}{"max_amount": 100.0}, // max_amount < amount
				}, nil)
				e.EXPECT().Create(ctx, tx, mock.Anything).Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Maybe()
			},
			expectedError: nil,
		},
		{
			name:     "DB Begin Error",
			userID:   1,
			amount:   50.0,
			category: "TRAVEL",
			mockSetup: func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, r *mocks.RuleService, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(nil, apperrors.ErrTransactionBegin)
			},
			expectedError: apperrors.ErrTransactionBegin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockE := mocks.NewExpenseRequestRepository(t)
			mockB := mocks.NewBalanceRepository(t)
			mockR := mocks.NewRuleService(t)
			mockU := mocks.NewUserRepository(t)
			mockDB := mocks.NewDB(t)
			mockTx := mocks.NewTx(t)

			tt.mockSetup(mockE, mockB, mockR, mockU, mockDB, mockTx)

			service := expense_service.NewExpenseService(ctx, mockE, mockB, mockR, mockU, mockDB)
			_, _, err := service.ApplyExpense(ctx, tt.userID, tt.amount, tt.category, tt.reason)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExpenseApprovalService_ApproveExpense(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		role          string
		approverID    int64
		requestID     int64
		comment       string
		mockSetup     func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx)
		expectedError error
	}{
		{
			name:       "Success - Admin Approves",
			role:       constants.RoleAdmin,
			approverID: 3,
			requestID:  10,
			comment:    "Valid expense",
			mockSetup: func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				e.EXPECT().GetByID(ctx, tx, int64(10)).Return(&models.ExpenseRequest{
					ID:         10,
					EmployeeID: 1,
					Amount:     50.0,
					Status:     constants.StatusPending,
				}, nil)
				u.EXPECT().GetRole(ctx, tx, int64(1)).Return(constants.RoleEmployee, nil)
				b.EXPECT().DeductExpenseBalance(ctx, tx, int64(1), 50.0).Return(nil)
				e.EXPECT().UpdateStatus(ctx, tx, int64(10), "APPROVED", int64(3), "Valid expense").Return(nil)
				tx.EXPECT().Commit(ctx).Return(nil)
				tx.EXPECT().Rollback(ctx).Return(nil).Maybe()
			},
			expectedError: nil,
		},
		{
			name:       "Unauthorized Role",
			role:       constants.RoleEmployee,
			approverID: 1,
			requestID:  10,
			comment:    "Try to approve",
			mockSetup: func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
			},
			expectedError: apperrors.ErrEmployeeCannotApprove,
		},
		{
			name:       "Repository Error - GetByID",
			role:       constants.RoleAdmin,
			approverID: 3,
			requestID:  10,
			comment:    "OK",
			mockSetup: func(e *mocks.ExpenseRequestRepository, b *mocks.BalanceRepository, u *mocks.UserRepository, db *mocks.DB, tx *mocks.Tx) {
				db.EXPECT().Begin(ctx).Return(tx, nil)
				e.EXPECT().GetByID(ctx, tx, int64(10)).Return(nil, apperrors.ErrDatabase)
				tx.EXPECT().Rollback(ctx).Return(nil).Once()
			},
			expectedError: apperrors.ErrDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockE := mocks.NewExpenseRequestRepository(t)
			mockB := mocks.NewBalanceRepository(t)
			mockU := mocks.NewUserRepository(t)
			mockDB := mocks.NewDB(t)
			mockTx := mocks.NewTx(t)

			tt.mockSetup(mockE, mockB, mockU, mockDB, mockTx)

			service := expense_service.NewExpenseApprovalService(ctx, mockE, mockB, mockU, mockDB)
			err := service.ApproveExpense(ctx, tt.role, tt.approverID, tt.requestID, tt.comment)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExpenseApprovalService_RejectExpense(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockE := mocks.NewExpenseRequestRepository(t)
		mockU := mocks.NewUserRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockE.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(&models.ExpenseRequest{
			ID:         10,
			EmployeeID: 1,
			Status:     constants.StatusPending,
		}, nil)
		mockU.EXPECT().GetRole(ctx, mockTx, int64(1)).Return(constants.RoleEmployee, nil)
		mockE.EXPECT().UpdateStatus(ctx, mockTx, int64(10), "REJECTED", int64(2), "Too high").Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil)
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := expense_service.NewExpenseApprovalService(ctx, mockE, nil, mockU, mockDB)
		err := service.RejectExpense(ctx, constants.RoleManager, 2, 10, "Too high")

		assert.NoError(t, err)
	})

	t.Run("Unauthorized Role", func(t *testing.T) {
		service := expense_service.NewExpenseApprovalService(ctx, nil, nil, nil, nil)
		err := service.RejectExpense(ctx, constants.RoleEmployee, 1, 10, "No")

		assert.ErrorIs(t, err, apperrors.ErrEmployeeCannotApprove)
	})
}

func TestExpenseService_CancelExpense(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockE := mocks.NewExpenseRequestRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockE.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(&models.ExpenseRequest{
			ID:         10,
			EmployeeID: 1,
			Status:     constants.StatusPending,
		}, nil)
		mockE.EXPECT().Cancel(ctx, mockTx, int64(10)).Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil)
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := expense_service.NewExpenseService(ctx, mockE, nil, nil, nil, mockDB)
		err := service.CancelExpense(ctx, 1, 10)

		assert.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockE := mocks.NewExpenseRequestRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockE.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(nil, apperrors.ErrExpenseRequestNotFound)
		mockTx.EXPECT().Rollback(ctx).Return(nil).Once()

		service := expense_service.NewExpenseService(ctx, mockE, nil, nil, nil, mockDB)
		err := service.CancelExpense(ctx, 1, 10)

		assert.ErrorIs(t, err, apperrors.ErrExpenseRequestNotFound)
	})
}

func TestExpenseApprovalService_GetPendingExpenseRequests(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Manager", func(t *testing.T) {
		mockE := mocks.NewExpenseRequestRepository(t)
		mockE.EXPECT().GetPendingForManager(ctx, int64(2)).Return([]map[string]interface{}{}, nil)

		service := expense_service.NewExpenseApprovalService(ctx, mockE, nil, nil, nil)
		_, err := service.GetPendingExpenseRequests(ctx, constants.RoleManager, 2)

		assert.NoError(t, err)
	})
}
