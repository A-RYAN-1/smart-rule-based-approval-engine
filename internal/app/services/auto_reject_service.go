package services

import (
	"context"
	"rule-based-approval-engine/internal/app/repositories"
	"rule-based-approval-engine/internal/pkg/utils"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AutoRejectService struct {
	leaveRepo   repositories.LeaveRequestRepository
	expenseRepo repositories.ExpenseRequestRepository
	holidayRepo repositories.HolidayRepository
	db          *pgxpool.Pool
}

func NewAutoRejectService(
	leaveRepo repositories.LeaveRequestRepository,
	expenseRepo repositories.ExpenseRequestRepository,
	holidayRepo repositories.HolidayRepository,
	db *pgxpool.Pool,
) *AutoRejectService {
	return &AutoRejectService{
		leaveRepo:   leaveRepo,
		expenseRepo: expenseRepo,
		holidayRepo: holidayRepo,
		db:          db,
	}
}

func (s *AutoRejectService) AutoRejectLeaveRequests(ctx context.Context) error {
	rows, err := s.db.Query(
		ctx,
		`SELECT id, created_at 
		 FROM leave_requests 
		 WHERE status='PENDING'`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	now := time.Now()

	for rows.Next() {
		var id int64
		var createdAt time.Time

		if err := rows.Scan(&id, &createdAt); err != nil {
			return err
		}

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

	return rows.Err()
}

func (s *AutoRejectService) AutoRejectExpenseRequests(ctx context.Context) error {
	rows, err := s.db.Query(
		ctx,
		`SELECT id, created_at 
		 FROM expense_requests 
		 WHERE status='PENDING'`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	now := time.Now()

	for rows.Next() {
		var id int64
		var createdAt time.Time

		if err := rows.Scan(&id, &createdAt); err != nil {
			return err
		}

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

	return rows.Err()
}

func (s *AutoRejectService) AutoRejectDiscountRequests(ctx context.Context) error {
	rows, err := s.db.Query(
		ctx,
		`SELECT id, created_at 
		 FROM discount_requests 
		 WHERE status='PENDING'`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	now := time.Now()

	for rows.Next() {
		var id int64
		var createdAt time.Time

		if err := rows.Scan(&id, &createdAt); err != nil {
			return err
		}

		workingDays := utils.CountWorkingDays(createdAt, now, func(d time.Time) bool {
			isHoliday, _ := s.holidayRepo.IsHoliday(ctx, d)
			return isHoliday
		})

		if workingDays >= 7 {
			_, err = s.db.Exec(
				ctx,
				`UPDATE discount_requests
				 SET status='AUTO_REJECTED',
				     approval_comment='Auto rejected after 7 working days'
				 WHERE id=$1`,
				id,
			)
			if err != nil {
				return err
			}
		}
	}

	return rows.Err()
}
