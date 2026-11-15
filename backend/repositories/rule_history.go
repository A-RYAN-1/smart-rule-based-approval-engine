package repositories

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/app/rule_dryrun"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
)

// One query per request type: the value a rule is compared with, for the newest
// requests of employees in a grade. Cancelled requests are left out.
var ruleHistoryQueries = map[string]string{
	"LEAVE": `SELECT r.days::float8 FROM leave_requests r JOIN users u ON u.id = r.employee_id
		WHERE u.grade_id = $1 AND r.status::text <> 'CANCELLED' ORDER BY r.created_at DESC, r.id DESC LIMIT $2`,
	"EXPENSE": `SELECT r.amount::float8 FROM expense_requests r JOIN users u ON u.id = r.employee_id
		WHERE u.grade_id = $1 AND r.status::text <> 'CANCELLED' ORDER BY r.created_at DESC, r.id DESC LIMIT $2`,
	"DISCOUNT": `SELECT r.discount_percentage::float8 FROM discount_requests r JOIN users u ON u.id = r.employee_id
		WHERE u.grade_id = $1 AND r.status::text <> 'CANCELLED' ORDER BY r.created_at DESC, r.id DESC LIMIT $2`,
}

type ruleHistory struct{ db interfaces.DB }

// NewRuleHistory reads recent real requests for the rule dry run.
func NewRuleHistory(db interfaces.DB) rule_dryrun.History { return &ruleHistory{db: db} }

func (h *ruleHistory) RecentValues(ctx context.Context, requestType string, gradeID int64, limit int) ([]float64, error) {
	query, ok := ruleHistoryQueries[requestType]
	if !ok {
		return nil, apperrors.ErrRequestTypeRequired
	}
	rows, err := h.db.Query(ctx, query, gradeID, limit)
	if err != nil {
		return nil, apperrors.ErrDatabase
	}
	defer rows.Close()
	values := []float64{}
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, apperrors.ErrDatabase
		}
		values = append(values, v)
	}
	if rows.Err() != nil {
		return nil, apperrors.ErrDatabase
	}
	return values, nil
}
