package services

import (
	"context"
	"rule-based-approval-engine/internal/app/repositories"
	"rule-based-approval-engine/internal/pkg/utils"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AutoRejectService struct {
	leaveRepo    repositories.LeaveRequestRepository
	expenseRepo  repositories.ExpenseRequestRepository
	discountRepo repositories.DiscountRequestRepository
	holidayRepo  repositories.HolidayRepository
	db           *pgxpool.Pool
}

func NewAutoRejectService(
	leaveRepo repositories.LeaveRequestRepository,
	expenseRepo repositories.ExpenseRequestRepository,
	discountRepo repositories.DiscountRequestRepository,
	holidayRepo repositories.HolidayRepository,
	db *pgxpool.Pool,
) *AutoRejectService {
	return &AutoRejectService{
		leaveRepo:    leaveRepo,
		expenseRepo:  expenseRepo,
		discountRepo: discountRepo,
		holidayRepo:  holidayRepo,
		db:           db,
	}
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
