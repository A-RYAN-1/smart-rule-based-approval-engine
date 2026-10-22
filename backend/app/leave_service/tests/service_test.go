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

func day(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

type leaveFixture struct {
	leaveRepo   *mocks.LeaveRequestRepository
	balanceRepo *mocks.BalanceRepository
	ruleService *mocks.RuleService
	userRepo    *mocks.UserRepository
	holidayRepo *mocks.HolidayRepository
	db          *mocks.DB
	tx          *mocks.Tx
	service     *leave_service.LeaveService
}

func newLeaveFixture(t *testing.T) *leaveFixture {
	f := &leaveFixture{
		leaveRepo:   mocks.NewLeaveRequestRepository(t),
		balanceRepo: mocks.NewBalanceRepository(t),
		ruleService: mocks.NewRuleService(t),
		userRepo:    mocks.NewUserRepository(t),
		holidayRepo: mocks.NewHolidayRepository(t),
		db:          mocks.NewDB(t),
		tx:          mocks.NewTx(t),
	}
	f.service = leave_service.NewLeaveService(
		context.Background(), f.leaveRepo, f.balanceRepo, f.ruleService, f.userRepo, f.holidayRepo, f.db,
	).(*leave_service.LeaveService)
	return f
}

// beginTx makes the database hand out the fixture's transaction, which is
// committed and then rolled back by the deferred rollback, as in production.
func (f *leaveFixture) beginTx(ctx context.Context) {
	f.db.EXPECT().Begin(ctx).Return(f.tx, nil)
	f.tx.EXPECT().Commit(ctx).Return(nil)
	f.tx.EXPECT().Rollback(ctx).Return(nil)
}

func TestCountLeaveDays(t *testing.T) {
	ctx := context.Background()
	friday, monday := day("2031-03-07"), day("2031-03-10")

	t.Run("Friday to Monday counts two days", func(t *testing.T) {
		f := newLeaveFixture(t)
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, friday, monday).Return(nil, nil)

		days, err := f.service.CountLeaveDays(ctx, friday, monday)

		assert.NoError(t, err)
		assert.Equal(t, 2, days)
	})

	t.Run("a holiday inside the range is not counted", func(t *testing.T) {
		f := newLeaveFixture(t)
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, friday, monday).Return([]time.Time{monday}, nil)

		days, err := f.service.CountLeaveDays(ctx, friday, monday)

		assert.NoError(t, err)
		assert.Equal(t, 1, days)
	})

	t.Run("a holiday on a weekend does not count twice", func(t *testing.T) {
		f := newLeaveFixture(t)
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, friday, monday).Return([]time.Time{day("2031-03-08")}, nil)

		days, err := f.service.CountLeaveDays(ctx, friday, monday)

		assert.NoError(t, err)
		assert.Equal(t, 2, days)
	})

	t.Run("a week with one holiday", func(t *testing.T) {
		f := newLeaveFixture(t)
		from, to := day("2031-03-10"), day("2031-03-14")
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, from, to).Return([]time.Time{day("2031-03-12")}, nil)

		days, err := f.service.CountLeaveDays(ctx, from, to)

		assert.NoError(t, err)
		assert.Equal(t, 4, days)
	})

	t.Run("from after to is refused without a database call", func(t *testing.T) {
		f := newLeaveFixture(t)

		_, err := f.service.CountLeaveDays(ctx, monday, friday)

		assert.ErrorIs(t, err, apperrors.ErrInvalidDateRange)
	})

	t.Run("a holiday lookup failure is returned", func(t *testing.T) {
		f := newLeaveFixture(t)
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, friday, monday).Return(nil, assert.AnError)

		_, err := f.service.CountLeaveDays(ctx, friday, monday)

		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestApplyLeave_WorkingDays(t *testing.T) {
	ctx := context.Background()
	userID := int64(7)

	t.Run("a weekend-only request is refused", func(t *testing.T) {
		f := newLeaveFixture(t)
		saturday, sunday := day("2031-03-08"), day("2031-03-09")
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, saturday, sunday).Return(nil, nil)

		_, _, err := f.service.ApplyLeave(ctx, userID, saturday, sunday, "CASUAL", "rest")

		assert.ErrorIs(t, err, apperrors.ErrInvalidLeaveDays)
	})

	t.Run("a range of holidays only is refused", func(t *testing.T) {
		f := newLeaveFixture(t)
		monday := day("2031-03-10")
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, monday, monday).Return([]time.Time{monday}, nil)

		_, _, err := f.service.ApplyLeave(ctx, userID, monday, monday, "CASUAL", "rest")

		assert.ErrorIs(t, err, apperrors.ErrInvalidLeaveDays)
	})

	t.Run("Friday to Monday is checked against the balance and stored as two days", func(t *testing.T) {
		f := newLeaveFixture(t)
		friday, monday := day("2031-03-07"), day("2031-03-10")
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, friday, monday).Return(nil, nil)
		f.leaveRepo.EXPECT().CheckOverlap(ctx, userID, friday, monday).Return(false, nil)
		f.beginTx(ctx)
		// Two days left: enough for the two working days, not for four calendar days.
		f.balanceRepo.EXPECT().GetLeaveBalance(ctx, f.tx, userID).Return(2, nil)
		f.userRepo.EXPECT().GetGrade(ctx, f.tx, userID).Return(int64(1), nil)
		f.ruleService.EXPECT().GetRule(ctx, "LEAVE", int64(1)).
			Return(&models.Rule{ID: 3, Condition: map[string]interface{}{"max_days": float64(3)}}, nil)
		f.leaveRepo.EXPECT().Create(ctx, f.tx, mock.MatchedBy(func(r *models.LeaveRequest) bool {
			return r.Days == 2 && r.Status == constants.StatusAutoApproved
		})).Return(nil)
		f.balanceRepo.EXPECT().DeductLeaveBalance(ctx, f.tx, userID, 2).Return(nil)

		_, status, err := f.service.ApplyLeave(ctx, userID, friday, monday, "CASUAL", "long weekend")

		assert.NoError(t, err)
		assert.Equal(t, constants.StatusAutoApproved, status)
	})

	t.Run("more working days than the balance is refused", func(t *testing.T) {
		f := newLeaveFixture(t)
		from, to := day("2031-03-10"), day("2031-03-14")
		f.holidayRepo.EXPECT().GetHolidayDatesBetween(ctx, from, to).Return(nil, nil)
		f.leaveRepo.EXPECT().CheckOverlap(ctx, userID, from, to).Return(false, nil)
		f.db.EXPECT().Begin(ctx).Return(f.tx, nil)
		f.tx.EXPECT().Rollback(ctx).Return(nil)
		f.balanceRepo.EXPECT().GetLeaveBalance(ctx, f.tx, userID).Return(4, nil)

		_, _, err := f.service.ApplyLeave(ctx, userID, from, to, "CASUAL", "week off")

		assert.ErrorIs(t, err, apperrors.ErrLeaveBalanceExceeded)
	})
}

func TestCancelLeave_RestoresStoredDays(t *testing.T) {
	ctx := context.Background()
	f := newLeaveFixture(t)
	userID, requestID := int64(7), int64(41)
	f.beginTx(ctx)
	// Friday to Monday: four calendar days, two of them leave days.
	f.leaveRepo.EXPECT().GetByID(ctx, f.tx, requestID).Return(&models.LeaveRequest{
		ID: requestID, EmployeeID: userID, Status: constants.StatusAutoApproved,
		FromDate: day("2031-03-07"), ToDate: day("2031-03-10"), Days: 2,
	}, nil)
	f.leaveRepo.EXPECT().Cancel(ctx, f.tx, requestID).Return(nil)
	f.balanceRepo.EXPECT().RestoreLeaveBalance(ctx, f.tx, userID, 2).Return(nil)

	assert.NoError(t, f.service.CancelLeave(ctx, userID, requestID))
}

func TestApproveLeave_DeductsStoredDays(t *testing.T) {
	ctx := context.Background()
	leaveReq := mocks.NewLeaveRequestRepository(t)
	balanceRepo := mocks.NewBalanceRepository(t)
	userRepo := mocks.NewUserRepository(t)
	db := mocks.NewDB(t)
	tx := mocks.NewTx(t)
	service := leave_service.NewLeaveApprovalService(ctx, leaveReq, balanceRepo, userRepo, db)

	approverID, employeeID, requestID := int64(2), int64(7), int64(41)
	db.EXPECT().Begin(ctx).Return(tx, nil)
	tx.EXPECT().Commit(ctx).Return(nil)
	tx.EXPECT().Rollback(ctx).Return(nil)
	leaveReq.EXPECT().GetByID(ctx, tx, requestID).Return(&models.LeaveRequest{
		ID: requestID, EmployeeID: employeeID, Status: constants.StatusPending,
		FromDate: day("2031-03-07"), ToDate: day("2031-03-10"), Days: 2,
	}, nil)
	userRepo.EXPECT().GetRole(ctx, tx, employeeID).Return(constants.RoleEmployee, nil)
	balanceRepo.EXPECT().DeductLeaveBalance(ctx, tx, employeeID, 2).Return(nil)
	leaveReq.EXPECT().UpdateStatus(ctx, tx, requestID, "APPROVED", approverID, "ok").Return(nil)

	assert.NoError(t, service.ApproveLeave(ctx, constants.RoleManager, approverID, requestID, "ok"))
}
