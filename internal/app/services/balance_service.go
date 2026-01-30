package services

import (
	"context"
	"rule-based-approval-engine/internal/app/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceService struct {
	balanceRepo repositories.BalanceRepository
	db          *pgxpool.Pool
}

func NewBalanceService(balanceRepo repositories.BalanceRepository, db *pgxpool.Pool) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		db:          db,
	}
}

func (s *BalanceService) GetMyBalances(ctx context.Context, userID int64) (map[string]interface{}, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var leaveTotal, leaveRemaining int
	var expenseTotal, expenseRemaining float64
	var discountTotal, discountRemaining float64

	leaveRemaining, err = s.balanceRepo.GetLeaveBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	expenseRemaining, err = s.balanceRepo.GetExpenseBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	// Fetch totals and discount as well
	err = tx.QueryRow(ctx, `SELECT total_allocated FROM leaves WHERE user_id=$1`, userID).Scan(&leaveTotal)
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(ctx, `SELECT total_amount FROM expense WHERE user_id=$1`, userID).Scan(&expenseTotal)
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(ctx, `SELECT total_discount, remaining_discount FROM discount WHERE user_id=$1`, userID).Scan(&discountTotal, &discountRemaining)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"leave": map[string]interface{}{
			"total":     leaveTotal,
			"remaining": leaveRemaining,
		},
		"expense": map[string]interface{}{
			"total":     expenseTotal,
			"remaining": expenseRemaining,
		},
		"discount": map[string]interface{}{
			"total":     discountTotal,
			"remaining": discountRemaining,
		},
	}, nil
}
