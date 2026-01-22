package services

import (
	"context"
	"errors"
	"time"

	"rule-based-approval-engine/internal/database"
	"rule-based-approval-engine/internal/utils"

	"github.com/jackc/pgx/v5"
)

func GetPendingLeaveRequests(role string, approverID int64) ([]map[string]interface{}, error) {
	ctx := context.Background()

	var rows pgx.Rows
	var err error

	if role == "MANAGER" {
		rows, err = database.DB.Query(
			ctx,
			`SELECT lr.id, u.name, lr.from_date, lr.to_date, lr.leave_type
			 FROM leave_requests lr
			 JOIN users u ON lr.employee_id = u.id
			 WHERE lr.status='PENDING' AND u.manager_id=$1`,
			approverID,
		)
	} else if role == "ADMIN" {
		rows, err = database.DB.Query(
			ctx,
			`SELECT lr.id, u.name, lr.from_date, lr.to_date, lr.leave_type
			 FROM leave_requests lr
			 JOIN users u ON lr.employee_id = u.id
			 WHERE lr.status='PENDING' AND u.role='MANAGER'`,
		)
	} else {
		return nil, errors.New("unauthorized")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, leaveType string
		var from, to time.Time

		rows.Scan(&id, &name, &from, &to, &leaveType)

		result = append(result, map[string]interface{}{
			"id":         id,
			"employee":   name,
			"from_date":  from,
			"to_date":    to,
			"leave_type": leaveType,
		})
	}

	return result, nil
}
func ApproveLeave(role string, approverID, requestID int64) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var employeeID int64
	var status string
	var from, to time.Time

	err = tx.QueryRow(
		ctx,
		`SELECT employee_id, status, from_date, to_date
		 FROM leave_requests
		 WHERE id=$1`,
		requestID,
	).Scan(&employeeID, &status, &from, &to)

	if err != nil {
		return err
	}

	if status != "PENDING" {
		return errors.New("request not pending")
	}

	// Authorization
	var requesterRole string
	err = tx.QueryRow(
		ctx,
		`SELECT role FROM users WHERE id=$1`,
		employeeID,
	).Scan(&requesterRole)

	if role == "MANAGER" && requesterRole != "EMPLOYEE" {
		return errors.New("manager can approve only employees")
	}

	if role == "ADMIN" && requesterRole != "MANAGER" {
		return errors.New("admin can approve only managers")
	}

	days := utils.CalculateLeaveDays(from, to)

	// Deduct leave balance
	_, err = tx.Exec(
		ctx,
		`UPDATE leaves
		 SET remaining_count = remaining_count - $1
		 WHERE user_id=$2`,
		days, employeeID,
	)
	if err != nil {
		return err
	}

	// Update request
	_, err = tx.Exec(
		ctx,
		`UPDATE leave_requests
		 SET status='APPROVED', approved_by_id=$1
		 WHERE id=$2`,
		approverID, requestID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func RejectLeave(role string, approverID, requestID int64) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	var employeeID int64

	err = tx.QueryRow(
		ctx,
		`SELECT employee_id, status
		 FROM leave_requests
		 WHERE id=$1`,
		requestID,
	).Scan(&employeeID, &status)

	if err != nil {
		return err
	}

	if status != "PENDING" {
		return errors.New("request not pending")
	}

	// check requester role
	var requesterRole string
	err = tx.QueryRow(
		ctx,
		`SELECT role FROM users WHERE id=$1`,
		employeeID,
	).Scan(&requesterRole)

	if role == "MANAGER" && requesterRole != "EMPLOYEE" {
		return errors.New("manager can reject only employees")
	}

	if role == "ADMIN" && requesterRole != "MANAGER" {
		return errors.New("admin can reject only managers")
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE leave_requests
		 SET status='REJECTED',
		     approved_by_id=$1
		 WHERE id=$2`,
		approverID, requestID,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
