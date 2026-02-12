package repositories

import (
	"context"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"

	"github.com/jackc/pgx/v5"
)

const (
	expenseQueryCreate = `INSERT INTO expense_requests
		 (employee_id, amount, category, reason, status, rule_id)
		 VALUES ($1, $2, $3, $4, $5, $6)`
	expenseQueryGetByID = `SELECT employee_id, status, amount
		 FROM expense_requests
		 WHERE id=$1`
	expenseQueryUpdateStatus = `UPDATE expense_requests
		 SET status=$1,
		     approved_by_id=$2,
		     approval_comment=$3
		 WHERE id=$4`
	expenseQueryGetPendingForManager = `SELECT er.id, u.name, er.amount, er.category, er.reason, er.created_at 
		 FROM expense_requests er
		 JOIN users u ON er.employee_id = u.id
		 WHERE er.status='PENDING' AND u.manager_id=$1
		 ORDER BY er.created_at DESC
		 LIMIT $2 OFFSET $3`
	expenseQueryGetPendingForAdmin = `SELECT er.id, u.name, er.amount, er.category, er.reason, er.created_at
		 FROM expense_requests er
		 JOIN users u ON er.employee_id = u.id
		 WHERE er.status='PENDING'
		 ORDER BY er.created_at DESC
		 LIMIT $1 OFFSET $2`
	expenseQueryCancel             = `UPDATE expense_requests SET status='CANCELLED' WHERE id=$1`
	expenseQueryGetPendingRequests = "SELECT id, created_at FROM expense_requests WHERE status='PENDING'"
	expenseQueryCountPendingForManager = `SELECT COUNT(*) FROM expense_requests er JOIN users u ON er.employee_id = u.id WHERE er.status='PENDING' AND u.manager_id=$1`
	expenseQueryCountPendingForAdmin   = `SELECT COUNT(*) FROM expense_requests WHERE status='PENDING'`
)

type expenseRequestRepository struct {
	db interfaces.DB
}

// NewExpenseRequestRepository creates a new instance
func NewExpenseRequestRepository(ctx context.Context, db interfaces.DB) interfaces.ExpenseRequestRepository {
	return &expenseRequestRepository{db: db}
}

func (r *expenseRequestRepository) Create(ctx context.Context, tx interfaces.Tx, req *models.ExpenseRequest) error {
	_, err := tx.Exec(
		ctx,
		expenseQueryCreate,
		req.EmployeeID,
		req.Amount,
		req.Category,
		req.Reason,
		req.Status,
		req.RuleID,
	)

	return utils.MapPgError(err)
}

func (r *expenseRequestRepository) GetByID(ctx context.Context, tx interfaces.Tx, requestID int64) (*models.ExpenseRequest, error) {
	var req models.ExpenseRequest

	err := tx.QueryRow(
		ctx,
		expenseQueryGetByID,
		requestID,
	).Scan(&req.EmployeeID, &req.Status, &req.Amount)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrExpenseRequestNotFound
		}
		return nil, utils.MapPgError(err)
	}

	req.ID = requestID
	return &req, nil
}

func (r *expenseRequestRepository) UpdateStatus(ctx context.Context, tx interfaces.Tx, requestID int64, status string, approverID int64, comment string) error {
	_, err := tx.Exec(
		ctx,
		expenseQueryUpdateStatus,
		status, approverID, comment, requestID,
	)

	return utils.MapPgError(err)
}

func (r *expenseRequestRepository) GetPendingForManager(ctx context.Context, managerID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	var total int
	err := r.db.QueryRow(ctx, expenseQueryCountPendingForManager, managerID).Scan(&total)
	if err != nil {
		return nil, 0, utils.MapPgError(err)
	}

	rows, err := r.db.Query(
		ctx,
		expenseQueryGetPendingForManager,
		managerID, limit, offset,
	)
	if err != nil {
		return nil, total, utils.MapPgError(err)
	}
	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var id int64
		var name, category string
		var reason *string
		var amount float64
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &amount, &category, &reason, &createdAt); err != nil {
			return nil, total, utils.MapPgError(err)
		}

		result = append(result, map[string]interface{}{
			"id":         id,
			"employee":   name,
			"amount":     amount,
			"category":   category,
			"reason":     reason,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	return result, total, nil
}

func (r *expenseRequestRepository) GetPendingForAdmin(ctx context.Context, limit, offset int) ([]map[string]interface{}, int, error) {
	var total int
	err := r.db.QueryRow(ctx, expenseQueryCountPendingForAdmin).Scan(&total)
	if err != nil {
		return nil, 0, utils.MapPgError(err)
	}

	rows, err := r.db.Query(
		ctx,
		expenseQueryGetPendingForAdmin,
		limit, offset,
	)
	if err != nil {
		return nil, total, utils.MapPgError(err)
	}
	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var id int64
		var name, category string
		var reason *string
		var amount float64
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &amount, &category, &reason, &createdAt); err != nil {
			return nil, total, utils.MapPgError(err)
		}

		result = append(result, map[string]interface{}{
			"id":         id,
			"employee":   name,
			"amount":     amount,
			"category":   category,
			"reason":     reason,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	return result, total, nil
}

func (r *expenseRequestRepository) Cancel(ctx context.Context, tx interfaces.Tx, requestID int64) error {
	_, err := tx.Exec(ctx, expenseQueryCancel, requestID)
	return utils.MapPgError(err)
}

func (r *expenseRequestRepository) GetPendingRequests(ctx context.Context) ([]struct {
	ID        int64
	CreatedAt time.Time
}, error) {
	rows, err := r.db.Query(ctx, expenseQueryGetPendingRequests)
	if err != nil {
		return nil, utils.MapPgError(err)
	}
	defer rows.Close()

	var result []struct {
		ID        int64
		CreatedAt time.Time
	}
	for rows.Next() {
		var item struct {
			ID        int64
			CreatedAt time.Time
		}
		if err := rows.Scan(&item.ID, &item.CreatedAt); err != nil {
			return nil, utils.MapPgError(err)
		}
		result = append(result, item)
	}
	return result, nil
}
