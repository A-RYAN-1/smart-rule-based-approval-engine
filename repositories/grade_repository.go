package repositories

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

const (
	gradeQueryGetLimits = `SELECT annual_leave_limit, annual_expense_limit, discount_limit_percent
		 FROM grades WHERE id=$1`
)

type gradeRepository struct {
	db interfaces.DB
}

// NewGradeRepository creates a new instance
func NewGradeRepository(ctx context.Context, db interfaces.DB) interfaces.GradeRepository {
	return &gradeRepository{db: db}
}

func (r *gradeRepository) GetLimits(ctx context.Context, tx interfaces.Tx, gradeID int64) (leaveLimit int, expenseLimit float64, discountLimit float64, err error) {
	err = tx.QueryRow(
		ctx,
		gradeQueryGetLimits,
		gradeID,
	).Scan(&leaveLimit, &expenseLimit, &discountLimit)

	err = utils.MapPgError(err)
	return
}
