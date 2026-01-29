package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type gradeRepository struct {
	db *pgxpool.Pool
}

// NewGradeRepository creates a new instance of GradeRepository
func NewGradeRepository(db *pgxpool.Pool) GradeRepository {
	return &gradeRepository{db: db}
}

func (r *gradeRepository) GetLimits(ctx context.Context, tx pgx.Tx, gradeID int64) (leaveLimit int, expenseLimit float64, discountLimit float64, err error) {
	err = tx.QueryRow(
		ctx,
		`SELECT annual_leave_limit, annual_expense_limit, discount_limit_percent
		 FROM grades WHERE id=$1`,
		gradeID,
	).Scan(&leaveLimit, &expenseLimit, &discountLimit)

	return
}
