package repositories

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/app/rule_drafting"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

const (
	gradeQueryList      = `SELECT id, name FROM grades ORDER BY id`
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

// NewGradeLister returns the grade listing used by rule drafting.
func NewGradeLister(ctx context.Context, db interfaces.DB) rule_drafting.GradeSource {
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

// ListGrades returns every grade, lowest id first, for the rule-drafting prompt.
func (r *gradeRepository) ListGrades(ctx context.Context) ([]rule_drafting.Grade, error) {
	rows, err := r.db.Query(ctx, gradeQueryList)
	if err != nil {
		return nil, utils.MapPgError(err)
	}
	defer rows.Close()

	var grades []rule_drafting.Grade
	for rows.Next() {
		var g rule_drafting.Grade
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, utils.MapPgError(err)
		}
		grades = append(grades, g)
	}
	return grades, utils.MapPgError(rows.Err())
}
