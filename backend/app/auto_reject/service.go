package auto_reject

import (
	"context"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

type AutoRejectService struct {
	leaveRepo    interfaces.LeaveRequestRepository
	expenseRepo  interfaces.ExpenseRequestRepository
	discountRepo interfaces.DiscountRequestRepository
	holidayRepo  interfaces.HolidayRepository
	db           interfaces.DB
}

func NewAutoRejectService(
	ctx context.Context,
	leaveRepo interfaces.LeaveRequestRepository,
	expenseRepo interfaces.ExpenseRequestRepository,
	discountRepo interfaces.DiscountRequestRepository,
	holidayRepo interfaces.HolidayRepository,
	db interfaces.DB,
) interfaces.AutoRejectService {
	return &AutoRejectService{
		leaveRepo:    leaveRepo,
		expenseRepo:  expenseRepo,
		discountRepo: discountRepo,
		holidayRepo:  holidayRepo,
		db:           db,
	}
}

func (s *AutoRejectService) AutoRejectExpiredRequests(ctx context.Context) error {
	if err := s.AutoRejectLeaveRequests(ctx); err != nil {
		return err
	}
	if err := s.AutoRejectExpenseRequests(ctx); err != nil {
		return err
	}
	if err := s.AutoRejectDiscountRequests(ctx); err != nil {
		return err
	}
	return nil
}

func (s *AutoRejectService) AutoRejectLeaveRequests(ctx context.Context) error {
	requests, err := s.leaveRepo.GetPendingRequests(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	for _, req := range requests {
		id := req.ID
		createdAt := req.CreatedAt

		workingDays := utils.CountWorkingDays(createdAt, now, func(d time.Time) bool {
			isHoliday, _ := s.holidayRepo.IsHoliday(ctx, d)
			return isHoliday
		})

		if workingDays >= 7 {
			tx, err := s.db.Begin(ctx)
			if err != nil {
				return err
			}
			err = s.leaveRepo.UpdateStatus(ctx, tx, id, "AUTO_REJECTED", 0, "Auto rejected after 7 working days")
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
			tx.Commit(ctx)
		}
	}

	return nil
}

func (s *AutoRejectService) AutoRejectExpenseRequests(ctx context.Context) error {
	requests, err := s.expenseRepo.GetPendingRequests(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	for _, req := range requests {
		id := req.ID
		createdAt := req.CreatedAt

		workingDays := utils.CountWorkingDays(createdAt, now, func(d time.Time) bool {
			isHoliday, _ := s.holidayRepo.IsHoliday(ctx, d)
			return isHoliday
		})

		if workingDays >= 7 {
			tx, err := s.db.Begin(ctx)
			if err != nil {
				return err
			}
			err = s.expenseRepo.UpdateStatus(ctx, tx, id, "AUTO_REJECTED", 0, "Auto rejected after 7 working days")
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
			tx.Commit(ctx)
		}
	}

	return nil
}

func (s *AutoRejectService) AutoRejectDiscountRequests(ctx context.Context) error {
	requests, err := s.discountRepo.GetPendingRequests(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	for _, req := range requests {
		id := req.ID
		createdAt := req.CreatedAt

		workingDays := utils.CountWorkingDays(createdAt, now, func(d time.Time) bool {
			isHoliday, _ := s.holidayRepo.IsHoliday(ctx, d)
			return isHoliday
		})

		if workingDays >= 7 {
			tx, err := s.db.Begin(ctx)
			if err != nil {
				return err
			}
			err = s.discountRepo.UpdateStatus(ctx, tx, id, "AUTO_REJECTED", 0, "Auto rejected after 7 working days")
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
			tx.Commit(ctx)
		}
	}

	return nil
}
