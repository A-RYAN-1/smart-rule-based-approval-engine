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
	discountQueryCreate = `INSERT INTO discount_requests
		 (employee_id, discount_percentage, reason, status, rule_id)
		 VALUES ($1, $2, $3, $4, $5)`
	discountQueryGetByID = `SELECT id, employee_id, discount_percentage, reason, status, rule_id, approved_by_id, created_at
		 FROM discount_requests WHERE id=$1`
	discountQueryUpdateStatus = `UPDATE discount_requests
		 SET status=$1, approved_by_id=$2, approval_comment=$3
		 WHERE id=$4`
	discountQueryGetPendingForManager = `
		SELECT dr.id, dr.employee_id, u.name, dr.discount_percentage, dr.reason, dr.created_at
		FROM discount_requests dr
		JOIN users u ON dr.employee_id = u.id
		WHERE dr.status='PENDING' AND u.manager_id=$1
		ORDER BY dr.created_at DESC
		LIMIT $2 OFFSET $3
	`
	discountQueryGetPendingForAdmin = `
		SELECT dr.id, dr.employee_id, u.name, dr.discount_percentage, dr.reason, dr.created_at
		FROM discount_requests dr
		JOIN users u ON dr.employee_id = u.id
		WHERE dr.status='PENDING'
		ORDER BY dr.created_at DESC
		LIMIT $1 OFFSET $2
	`
	discountQueryCancel                 = `UPDATE discount_requests SET status='CANCELLED' WHERE id=$1`
	discountQueryGetPendingRequests     = "SELECT id, created_at FROM discount_requests WHERE status='PENDING'"
	discountQueryCountPendingForManager = `SELECT COUNT(*) FROM discount_requests dr JOIN users u ON dr.employee_id = u.id WHERE dr.status='PENDING' AND u.manager_id=$1`
	discountQueryCountPendingForAdmin   = `SELECT COUNT(*) FROM discount_requests WHERE status='PENDING'`
)

type discountRequestRepository struct {
	db interfaces.DB
}

func NewDiscountRequestRepository(ctx context.Context, db interfaces.DB) interfaces.DiscountRequestRepository {
	return &discountRequestRepository{db: db}
}

func (r *discountRequestRepository) Create(ctx context.Context, tx interfaces.Tx, req *models.DiscountRequest) error {
	_, err := tx.Exec(
		ctx,
		discountQueryCreate,
		req.EmployeeID, req.DiscountPercentage, req.Reason, req.Status, req.RuleID,
	)
	return utils.MapPgError(err)
}

func (r *discountRequestRepository) GetByID(ctx context.Context, tx interfaces.Tx, requestID int64) (*models.DiscountRequest, error) {
	reqObj := &models.DiscountRequest{}
	err := tx.QueryRow(
		ctx,
		discountQueryGetByID,
		requestID,
	).Scan(&reqObj.ID, &reqObj.EmployeeID, &reqObj.DiscountPercentage, &reqObj.Reason, &reqObj.Status, &reqObj.RuleID, &reqObj.ApprovedByID, &reqObj.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrDiscountRequestNotFound
		}
		return nil, utils.MapPgError(err)
	}
	return reqObj, nil
}

func (r *discountRequestRepository) UpdateStatus(ctx context.Context, tx interfaces.Tx, requestID int64, status string, approverID int64, comment string) error {
	_, err := tx.Exec(
		ctx,
		discountQueryUpdateStatus,
		status, approverID, comment, requestID,
	)
	return utils.MapPgError(err)
}

func (r *discountRequestRepository) GetPendingForManager(ctx context.Context, managerID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	var total int
	err := r.db.QueryRow(ctx, discountQueryCountPendingForManager, managerID).Scan(&total)
	if err != nil {
		return nil, 0, utils.MapPgError(err)
	}

	rows, err := r.db.Query(ctx, discountQueryGetPendingForManager, managerID, limit, offset)
	if err != nil {
		return nil, total, utils.MapPgError(err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var (
			id         int64
			employeeID int64
			name       string
			reason     string
			percent    float64
			created    interface{}
		)
		if err := rows.Scan(&id, &employeeID, &name, &percent, &reason, &created); err != nil {
			return nil, total, utils.MapPgError(err)
		}

		var createdAt time.Time
		switch v := created.(type) {
		case time.Time:
			createdAt = v
		case *time.Time:
			if v != nil {
				createdAt = *v
			}
		}

		result = append(result, map[string]interface{}{
			"id":                  id,
			"user_id":             employeeID,
			"employee":            name,
			"discount_percentage": percent,
			"reason":              reason,
			"status":              "PENDING",
			"created_at":          createdAt.Format(time.RFC3339),
		})
	}
	return result, total, nil
}

func (r *discountRequestRepository) GetPendingForAdmin(ctx context.Context, limit, offset int) ([]map[string]interface{}, int, error) {
	var total int
	err := r.db.QueryRow(ctx, discountQueryCountPendingForAdmin).Scan(&total)
	if err != nil {
		return nil, 0, utils.MapPgError(err)
	}

	rows, err := r.db.Query(ctx, discountQueryGetPendingForAdmin, limit, offset)
	if err != nil {
		return nil, total, utils.MapPgError(err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var (
			id         int64
			employeeID int64
			name       string
			reason     string
			percent    float64
			created    interface{}
		)
		if err := rows.Scan(&id, &employeeID, &name, &percent, &reason, &created); err != nil {
			return nil, total, utils.MapPgError(err)
		}

		var createdAt time.Time
		switch v := created.(type) {
		case time.Time:
			createdAt = v
		case *time.Time:
			if v != nil {
				createdAt = *v
			}
		}

		result = append(result, map[string]interface{}{
			"id":                  id,
			"user_id":             employeeID,
			"employee":            name,
			"discount_percentage": percent,
			"reason":              reason,
			"status":              "PENDING",
			"created_at":          createdAt.Format(time.RFC3339),
		})
	}
	return result, total, nil
}

func (r *discountRequestRepository) Cancel(ctx context.Context, tx interfaces.Tx, requestID int64) error {
	_, err := r.db.Exec(ctx, discountQueryCancel, requestID)
	return utils.MapPgError(err)
}

func (r *discountRequestRepository) GetPendingRequests(ctx context.Context) ([]struct {
	ID        int64
	CreatedAt time.Time
}, error) {
	rows, err := r.db.Query(ctx, discountQueryGetPendingRequests)
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
