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

	leaveTotal, leaveRemaining, err = s.balanceRepo.GetLeaveFullBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	expenseTotal, expenseRemaining, err = s.balanceRepo.GetExpenseFullBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	discountTotal, discountRemaining, err = s.balanceRepo.GetDiscountFullBalance(ctx, tx, userID)
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
