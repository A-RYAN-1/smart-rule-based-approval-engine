package repositories

import (
	"context"
	"time"

	"rule-based-approval-engine/internal/models"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type leaveRequestRepository struct {
	db *pgxpool.Pool
}

// NewLeaveRequestRepository creates a new instance of LeaveRequestRepository
func NewLeaveRequestRepository(db *pgxpool.Pool) LeaveRequestRepository {
	return &leaveRequestRepository{db: db}
}

func (r *leaveRequestRepository) Create(ctx context.Context, tx pgx.Tx, req *models.LeaveRequest) error {
	_, err := tx.Exec(
		ctx,
		`INSERT INTO leave_requests
		 (employee_id, from_date, to_date, reason, leave_type, status, rule_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		req.EmployeeID,
		req.FromDate,
		req.ToDate,
		req.Reason,
		req.LeaveType,
		req.Status,
		req.RuleID,
	)

	return err
}

func (r *leaveRequestRepository) GetByID(ctx context.Context, tx pgx.Tx, requestID int64) (*models.LeaveRequest, error) {
	var req models.LeaveRequest

	err := tx.QueryRow(
		ctx,
		`SELECT employee_id, status, from_date, to_date
		 FROM leave_requests
		 WHERE id=$1`,
		requestID,
	).Scan(&req.EmployeeID, &req.Status, &req.FromDate, &req.ToDate)

	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrLeaveRequestNotFound
	}
	if err != nil {
		return nil, err
	}

	req.ID = requestID
	return &req, nil
}

func (r *leaveRequestRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, requestID int64, status string, approverID int64, comment string) error {
	_, err := tx.Exec(
		ctx,
		`UPDATE leave_requests
		 SET status=$1,
		     approved_by_id=$2,
		     approval_comment=$3
		 WHERE id=$4`,
		status, approverID, comment, requestID,
	)

	return err
}

func (r *leaveRequestRepository) GetPendingForManager(ctx context.Context, managerID int64) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT lr.id, u.name, lr.from_date, lr.to_date, lr.leave_type, lr.reason, lr.created_at 
		 FROM leave_requests lr
		 JOIN users u ON lr.employee_id = u.id
		 WHERE lr.status='PENDING'
		   AND u.manager_id=$1`,
		managerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var id int64
		var name, leaveType, reason string
		var fromDate, toDate, createdAt time.Time

		if err := rows.Scan(&id, &name, &fromDate, &toDate, &leaveType, &reason, &createdAt); err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"id":         id,
			"employee":   name,
			"from_date":  fromDate.Format("2006-01-02"),
			"to_date":    toDate.Format("2006-01-02"),
			"leave_type": leaveType,
			"reason":     reason,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

func (r *leaveRequestRepository) GetPendingForAdmin(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT lr.id, u.name, lr.from_date, lr.to_date, lr.leave_type, lr.reason, lr.created_at
		 FROM leave_requests lr
		 JOIN users u ON lr.employee_id = u.id
		 WHERE lr.status='PENDING'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var id int64
		var name, leaveType, reason string
		var fromDate, toDate, createdAt time.Time

		if err := rows.Scan(&id, &name, &fromDate, &toDate, &leaveType, &reason, &createdAt); err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"id":         id,
			"employee":   name,
			"from_date":  fromDate.Format("2006-01-02"),
			"to_date":    toDate.Format("2006-01-02"),
			"leave_type": leaveType,
			"reason":     reason,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

func (r *leaveRequestRepository) CheckOverlap(ctx context.Context, userID int64, fromDate, toDate time.Time) (bool, error) {
	var dummy int

	err := r.db.QueryRow(
		ctx,
		`SELECT 1
		 FROM leave_requests
		 WHERE employee_id = $1
		   AND status IN ('PENDING', 'APPROVED', 'AUTO_APPROVED') 
		   AND from_date <= $2
		   AND to_date >= $3
		 LIMIT 1`,
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
		return false, err
	}

	// overlap exists
	return true, nil
}

func (r *leaveRequestRepository) Cancel(ctx context.Context, tx pgx.Tx, requestID int64) error {
	_, err := tx.Exec(
		ctx,
		`UPDATE leave_requests 
		 SET status='CANCELLED' 
		 WHERE id=$1`,
		requestID,
	)

	return err
}
