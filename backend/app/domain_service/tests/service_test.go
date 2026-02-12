package tests

import (
	"context"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/domain_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/domain_service/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDiscountService_ApplyDiscount(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success - Auto Approved", func(t *testing.T) {
		mockDiscountRepo := mocks.NewDiscountRequestRepository(t)
		mockBalanceRepo := mocks.NewBalanceRepository(t)
		mockRuleService := mocks.NewRuleService(t)
		mockUserRepo := mocks.NewUserRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockBalanceRepo.EXPECT().GetDiscountBalance(ctx, mockTx, userID).Return(50.0, nil)
		mockUserRepo.EXPECT().GetGrade(ctx, mockTx, userID).Return(int64(1), nil)
		mockRuleService.EXPECT().GetRule(ctx, "DISCOUNT", int64(1)).Return(&models.Rule{
			ID:        1,
			Condition: map[string]interface{}{"max_percent": 10.0},
		}, nil)
		mockDiscountRepo.EXPECT().Create(ctx, mockTx, mock.Anything).Return(nil)
		mockBalanceRepo.EXPECT().DeductDiscountBalance(ctx, mockTx, userID, 5.0).Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil).Once()
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := domain_service.NewDiscountService(ctx, mockDiscountRepo, mockBalanceRepo, mockRuleService, mockUserRepo, mockDB)
		msg, status, err := service.ApplyDiscount(ctx, userID, 5.0, "Reward")

		assert.NoError(t, err)
		assert.Equal(t, "AUTO_APPROVED", status)
		assert.Contains(t, msg, "approved by system")
	})

	t.Run("Fail - Limit Exceeded", func(t *testing.T) {
		mockDiscountRepo := mocks.NewDiscountRequestRepository(t)
		mockBalanceRepo := mocks.NewBalanceRepository(t)
		mockRuleService := mocks.NewRuleService(t)
		mockUserRepo := mocks.NewUserRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockBalanceRepo.EXPECT().GetDiscountBalance(ctx, mockTx, userID).Return(2.0, nil)
		mockTx.EXPECT().Rollback(ctx).Return(nil).Once()

		service := domain_service.NewDiscountService(ctx, mockDiscountRepo, mockBalanceRepo, mockRuleService, mockUserRepo, mockDB)
		_, _, err := service.ApplyDiscount(ctx, userID, 5.0, "Too much")

		assert.ErrorIs(t, err, apperrors.ErrDiscountLimitExceeded)
	})

	t.Run("Success - Manual Approval Pending", func(t *testing.T) {
		mockDiscountRepo := mocks.NewDiscountRequestRepository(t)
		mockBalanceRepo := mocks.NewBalanceRepository(t)
		mockRuleService := mocks.NewRuleService(t)
		mockUserRepo := mocks.NewUserRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockBalanceRepo.EXPECT().GetDiscountBalance(ctx, mockTx, userID).Return(50.0, nil)
		mockUserRepo.EXPECT().GetGrade(ctx, mockTx, userID).Return(int64(1), nil)
		mockRuleService.EXPECT().GetRule(ctx, "DISCOUNT", int64(1)).Return(&models.Rule{
			ID:        1,
			Condition: map[string]interface{}{"max_percent": 1.0}, // max < percent
		}, nil)
		mockDiscountRepo.EXPECT().Create(ctx, mockTx, mock.Anything).Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil).Once()
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := domain_service.NewDiscountService(ctx, mockDiscountRepo, mockBalanceRepo, mockRuleService, mockUserRepo, mockDB)
		_, status, err := service.ApplyDiscount(ctx, userID, 5.0, "Reward")

		assert.NoError(t, err)
		assert.Equal(t, "PENDING", status)
	})

	t.Run("DB Begin Error", func(t *testing.T) {
		mockDB := mocks.NewDB(t)
		mockDB.EXPECT().Begin(ctx).Return(nil, apperrors.ErrTransactionBegin)

		service := domain_service.NewDiscountService(ctx, nil, nil, nil, nil, mockDB)
		_, _, err := service.ApplyDiscount(ctx, userID, 5.0, "Fail")
		assert.ErrorIs(t, err, apperrors.ErrTransactionBegin)
	})
}

func TestBalanceService_GetBalances(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockBalanceRepo := mocks.NewBalanceRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockBalanceRepo.EXPECT().GetLeaveBalance(ctx, mockTx, userID).Return(10, nil)
		mockBalanceRepo.EXPECT().GetLeaveFullBalance(ctx, mockTx, userID).Return(30, 20, nil)
		mockBalanceRepo.EXPECT().GetExpenseFullBalance(ctx, mockTx, userID).Return(1000.0, 500.0, nil)
		mockBalanceRepo.EXPECT().GetDiscountFullBalance(ctx, mockTx, userID).Return(20.0, 15.0, nil)
		mockTx.EXPECT().Commit(ctx).Return(nil).Once()
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := domain_service.NewBalanceService(ctx, mockBalanceRepo, mockDB)
		result, err := service.GetBalances(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, 20.0, result["discount"].(map[string]interface{})["total"])
		assert.Equal(t, 20, result["leave"].(map[string]interface{})["remaining"])
	})
}

func TestDiscountApprovalService(t *testing.T) {
	ctx := context.Background()

	t.Run("ApproveDiscount - Success", func(t *testing.T) {
		mockDiscountRepo := mocks.NewDiscountRequestRepository(t)
		mockBalanceRepo := mocks.NewBalanceRepository(t)
		mockUserRepo := mocks.NewUserRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockDiscountRepo.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(&models.DiscountRequest{
			ID:                 10,
			EmployeeID:         1,
			DiscountPercentage: 5.0,
			Status:             "PENDING",
		}, nil)
		mockUserRepo.EXPECT().GetRole(ctx, mockTx, int64(1)).Return("EMPLOYEE", nil)
		mockDiscountRepo.EXPECT().UpdateStatus(ctx, mockTx, int64(10), "APPROVED", int64(2), "Good").Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil).Once()
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := domain_service.NewDiscountApprovalService(ctx, mockDiscountRepo, mockBalanceRepo, mockUserRepo, mockDB)
		err := service.ApproveDiscount(ctx, "MANAGER", 2, 10, "Good")

		assert.NoError(t, err)
	})

	t.Run("ApproveDiscount - Repository Error", func(t *testing.T) {
		mockDiscountRepo := mocks.NewDiscountRequestRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockDiscountRepo.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(nil, apperrors.ErrDatabase)
		mockTx.EXPECT().Rollback(ctx).Return(nil).Once()

		service := domain_service.NewDiscountApprovalService(ctx, mockDiscountRepo, nil, nil, mockDB)
		err := service.ApproveDiscount(ctx, "ADMIN", 1, 10, "OK")

		assert.Error(t, err)
	})
}

func TestDiscountApprovalService_RejectDiscount(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockDiscountRepo := mocks.NewDiscountRequestRepository(t)
		mockUserRepo := mocks.NewUserRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockDiscountRepo.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(&models.DiscountRequest{
			ID:         10,
			EmployeeID: 1,
			Status:     "PENDING",
		}, nil)
		mockUserRepo.EXPECT().GetRole(ctx, mockTx, int64(1)).Return("EMPLOYEE", nil)
		mockDiscountRepo.EXPECT().UpdateStatus(ctx, mockTx, int64(10), "REJECTED", int64(2), "No").Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil).Once()
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := domain_service.NewDiscountApprovalService(ctx, mockDiscountRepo, nil, mockUserRepo, mockDB)
		err := service.RejectDiscount(ctx, "MANAGER", 2, 10, "No")

		assert.NoError(t, err)
	})

	t.Run("Unauthorized Role", func(t *testing.T) {
		service := domain_service.NewDiscountApprovalService(ctx, nil, nil, nil, nil)
		err := service.RejectDiscount(ctx, "EMPLOYEE", 1, 10, "No")

		assert.ErrorIs(t, err, apperrors.ErrEmployeeCannotApprove)
	})
}

func TestDiscountService_CancelDiscount(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockDiscountRepo := mocks.NewDiscountRequestRepository(t)
		mockDB := mocks.NewDB(t)
		mockTx := mocks.NewTx(t)

		mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)
		mockDiscountRepo.EXPECT().GetByID(ctx, mockTx, int64(10)).Return(&models.DiscountRequest{
			ID:         10,
			EmployeeID: 1,
			Status:     "PENDING",
		}, nil)
		mockDiscountRepo.EXPECT().Cancel(ctx, mockTx, int64(10)).Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil).Once()
		mockTx.EXPECT().Rollback(ctx).Return(nil).Maybe()

		service := domain_service.NewDiscountService(ctx, mockDiscountRepo, nil, nil, nil, mockDB)
		err := service.CancelDiscount(ctx, 1, 10)

		assert.NoError(t, err)
	})
}
