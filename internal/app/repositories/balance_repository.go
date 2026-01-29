package repositories

import (
	"context"

	"rule-based-approval-engine/internal/app/services/helpers"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type balanceRepository struct {
	db *pgxpool.Pool
}

// NewBalanceRepository creates a new instance of BalanceRepository
func NewBalanceRepository(db *pgxpool.Pool) BalanceRepository {
	return &balanceRepository{db: db}
}

func (r *balanceRepository) GetLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64) (int, error) {
	var remaining int

	err := tx.QueryRow(
		ctx,
		`SELECT remaining_count FROM leaves WHERE user_id=$1`,
		userID,
	).Scan(&remaining)

	if err == pgx.ErrNoRows {
		return 0, apperrors.ErrLeaveBalanceMissing
	}
	if err != nil {
		return 0, apperrors.ErrBalanceFetchFailed
	}

	return remaining, nil
}

func (r *balanceRepository) GetExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64) (float64, error) {
	var remaining float64

	err := tx.QueryRow(
		ctx,
		`SELECT remaining_amount FROM expense WHERE user_id=$1`,
		userID,
	).Scan(&remaining)

	if err == pgx.ErrNoRows {
		return 0, apperrors.ErrExpenseBalanceMissing
	}
	if err != nil {
		return 0, apperrors.ErrBalanceFetchFailed
	}

	return remaining, nil
}

func (r *balanceRepository) DeductLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64, days int) error {
	_, err := tx.Exec(
		ctx,
		`UPDATE leaves
		 SET remaining_count = remaining_count - $1
		 WHERE user_id=$2`,
		days, userID,
	)

	if err != nil {
		return helpers.MapPgError(err)
	}

	return nil
}

func (r *balanceRepository) DeductExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64, amount float64) error {
	_, err := tx.Exec(
		ctx,
		`UPDATE expense
		 SET remaining_amount = remaining_amount - $1
		 WHERE user_id=$2`,
		amount, userID,
	)

	if err != nil {
		return helpers.MapPgError(err)
	}

	return nil
}

func (r *balanceRepository) RestoreLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64, days int) error {
	_, err := tx.Exec(
		ctx,
		`UPDATE leaves 
		 SET remaining_count = remaining_count + $1
		 WHERE user_id=$2`,
		days, userID,
	)

	return err
}

func (r *balanceRepository) RestoreExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64, amount float64) error {
	_, err := tx.Exec(
		ctx,
		`UPDATE expense
		 SET remaining_amount = remaining_amount + $1
		 WHERE user_id=$2`,
		amount, userID,
	)

	return err
}

func (r *balanceRepository) InitializeBalances(ctx context.Context, tx pgx.Tx, userID int64, gradeID int64) error {
	var leaveLimit int
	var expenseLimit float64
	var discountLimit float64

	err := tx.QueryRow(
		ctx,
		`SELECT annual_leave_limit, annual_expense_limit, discount_limit_percent
		 FROM grades WHERE id=$1`,
		gradeID,
	).Scan(&leaveLimit, &expenseLimit, &discountLimit)

	if err != nil {
		if err == pgx.ErrNoRows {
			return apperrors.ErrQueryFailed
		}
		return helpers.MapPgError(err)
	}

	// Leave wallet
	_, err = tx.Exec(
		ctx,
		`INSERT INTO leaves (user_id, total_allocated, remaining_count)
		 VALUES ($1,$2,$2)
		 ON CONFLICT (user_id) DO NOTHING`,
		userID, leaveLimit,
	)
	if err != nil {
		return helpers.MapPgError(err)
	}

	// Expense wallet
	_, err = tx.Exec(
		ctx,
		`INSERT INTO expense (user_id, total_amount, remaining_amount)
		 VALUES ($1,$2,$2)
		 ON CONFLICT (user_id) DO NOTHING`,
		userID, expenseLimit,
	)
	if err != nil {
		return helpers.MapPgError(err)
	}

	// Discount wallet
	_, err = tx.Exec(
		ctx,
		`INSERT INTO discount (user_id, total_discount, remaining_discount)
		 VALUES ($1,$2,$2)
		 ON CONFLICT (user_id) DO NOTHING`,
		userID, discountLimit,
	)
	if err != nil {
		return helpers.MapPgError(err)
	}

	return nil
}
