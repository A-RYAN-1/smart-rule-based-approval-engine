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
	leaveQueryCreate = `INSERT INTO leave_requests
		 (employee_id, from_date, to_date, reason, leave_type, status, rule_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`
	leaveQueryGetByID = `SELECT employee_id, status, from_date, to_date
		 FROM leave_requests
		 WHERE id=$1`
	leaveQueryUpdateStatus = `UPDATE leave_requests
		 SET status=$1,
		     approved_by_id=$2,
		     approval_comment=$3
		 WHERE id=$4`
	leaveQueryGetPendingForManager = `SELECT lr.id, lr.employee_id, u.name, lr.from_date, lr.to_date, lr.leave_type, lr.reason, lr.created_at 
		 FROM leave_requests lr
		 JOIN users u ON lr.employee_id = u.id
		 WHERE lr.status='PENDING'
		   AND u.manager_id=$1
		 ORDER BY lr.created_at DESC
		 LIMIT $2 OFFSET $3`
	leaveQueryGetPendingForAdmin = `SELECT lr.id, lr.employee_id, u.name, lr.from_date, lr.to_date, lr.leave_type, lr.reason, lr.created_at
		 FROM leave_requests lr
		 JOIN users u ON lr.employee_id = u.id
		 WHERE lr.status='PENDING'
		 ORDER BY lr.created_at DESC
		 LIMIT $1 OFFSET $2`
	leaveQueryCheckOverlap = `SELECT 1
		 FROM leave_requests
		 WHERE employee_id = $1
		   AND status IN ('PENDING', 'APPROVED', 'AUTO_APPROVED') 
		   AND from_date <= $2
		   AND to_date >= $3
		 LIMIT 1`
	leaveQueryCancel                 = `UPDATE leave_requests SET status='CANCELLED' WHERE id=$1`
	leaveQueryGetPendingRequests     = "SELECT id, created_at FROM leave_requests WHERE status='PENDING'"
	leaveQueryCountPendingForManager = `SELECT COUNT(*) FROM leave_requests lr JOIN users u ON lr.employee_id = u.id WHERE lr.status='PENDING' AND u.manager_id=$1`
	leaveQueryCountPendingForAdmin   = `SELECT COUNT(*) FROM leave_requests WHERE status='PENDING'`
)

type leaveRequestRepository struct {
	db interfaces.DB
}

func NewLeaveRequestRepository(ctx context.Context, db interfaces.DB) interfaces.LeaveRequestRepository {
	return &leaveRequestRepository{db: db}
}

func (r *leaveRequestRepository) Create(ctx context.Context, tx interfaces.Tx, req *models.LeaveRequest) error {
	_, err := tx.Exec(
		ctx,
		leaveQueryCreate,
		req.EmployeeID,
		req.FromDate,
		req.ToDate,
		req.Reason,
		req.LeaveType,
		req.Status,
		req.RuleID,
	)

	return utils.MapPgError(err)
}

func (r *leaveRequestRepository) GetByID(ctx context.Context, tx interfaces.Tx, requestID int64) (*models.LeaveRequest, error) {
	var req models.LeaveRequest

	err := tx.QueryRow(
		ctx,
		leaveQueryGetByID,
		requestID,
	).Scan(&req.EmployeeID, &req.Status, &req.FromDate, &req.ToDate)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrLeaveRequestNotFound
		}
		return nil, utils.MapPgError(err)
	}

	req.ID = requestID
	return &req, nil
}

func (r *leaveRequestRepository) UpdateStatus(ctx context.Context, tx interfaces.Tx, requestID int64, status string, approverID int64, comment string) error {
	_, err := tx.Exec(
		ctx,
		leaveQueryUpdateStatus,
		status, approverID, comment, requestID,
	)

	return utils.MapPgError(err)
}

func (r *leaveRequestRepository) GetPendingForManager(ctx context.Context, managerID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	var total int
	err := r.db.QueryRow(ctx, leaveQueryCountPendingForManager, managerID).Scan(&total)
	if err != nil {
		return nil, 0, utils.MapPgError(err)
	}

	rows, err := r.db.Query(
		ctx,
		leaveQueryGetPendingForManager,
		managerID, limit, offset,
	)
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
			leaveType  string
			reason     string
			fromDate   time.Time
			toDate     time.Time
			createdAt  time.Time
		)

		if err := rows.Scan(&id, &employeeID, &name, &fromDate, &toDate, &leaveType, &reason, &createdAt); err != nil {
			return nil, total, utils.MapPgError(err)
		}

		result = append(result, map[string]interface{}{
			"id":         id,
			"user_id":    employeeID,
			"employee":   name,
			"from_date":  fromDate.Format("2006-01-02"),
			"to_date":    toDate.Format("2006-01-02"),
			"leave_type": leaveType,
			"reason":     reason,
			"status":     "PENDING", // Since query filters by PENDING
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	return result, total, nil
}

func (r *leaveRequestRepository) GetPendingForAdmin(ctx context.Context, limit, offset int) ([]map[string]interface{}, int, error) {
	var total int
	err := r.db.QueryRow(ctx, leaveQueryCountPendingForAdmin).Scan(&total)
	if err != nil {
		return nil, 0, utils.MapPgError(err)
	}

	rows, err := r.db.Query(
		ctx,
		leaveQueryGetPendingForAdmin,
		limit, offset,
	)
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
			leaveType  string
			reason     string
			fromDate   time.Time
			toDate     time.Time
			createdAt  time.Time
		)

		if err := rows.Scan(&id, &employeeID, &name, &fromDate, &toDate, &leaveType, &reason, &createdAt); err != nil {
			return nil, total, utils.MapPgError(err)
		}

		result = append(result, map[string]interface{}{
			"id":         id,
			"user_id":    employeeID,
			"employee":   name,
			"from_date":  fromDate.Format("2006-01-02"),
			"to_date":    toDate.Format("2006-01-02"),
			"leave_type": leaveType,
			"reason":     reason,
			"status":     "PENDING",
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	return result, total, nil
}

func (r *leaveRequestRepository) CheckOverlap(ctx context.Context, userID int64, fromDate, toDate time.Time) (bool, error) {
	var dummy int

	err := r.db.QueryRow(
		ctx,
		leaveQueryCheckOverlap,
		userID,
		toDate,
		fromDate,
	).Scan(&dummy)

	// pgx NO ROWS = no overlap
	if err == pgx.ErrNoRows {
		return false, nil
	}

	// real system error
	if err != nil {
		return false, utils.MapPgError(err)
	}

	// overlap exists
	return true, nil
}

func (r *leaveRequestRepository) Cancel(ctx context.Context, tx interfaces.Tx, requestID int64) error {
	_, err := tx.Exec(ctx, leaveQueryCancel, requestID)
	return utils.MapPgError(err)
}

func (r *leaveRequestRepository) GetPendingRequests(ctx context.Context) ([]struct {
	ID        int64
	CreatedAt time.Time
}, error) {
	rows, err := r.db.Query(ctx, leaveQueryGetPendingRequests)
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
