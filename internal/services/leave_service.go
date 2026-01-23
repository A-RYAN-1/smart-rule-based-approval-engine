package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"rule-based-approval-engine/internal/apperrors"
	"rule-based-approval-engine/internal/database"
	"rule-based-approval-engine/internal/utils"
)

func ApplyLeave(
	userID int64,
	from time.Time,
	to time.Time,
	days int,
	leaveType string,
	reason string,
) (string, error) {

	// ---- Input validations ----
	if userID <= 0 {
		return "", errors.New("invalid user")
	}

	if days <= 0 {
		return "", apperrors.ErrInvalidLeaveDays
	}

	if from.After(to) {
		return "", errors.New("from date cannot be after to date")
	}

	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return "", errors.New("unable to start transaction")
	}
	defer tx.Rollback(ctx)

	// ---- Fetch remaining leave balance ----
	var remaining int
	err = tx.QueryRow(
		ctx,
		`SELECT remaining_count FROM leaves WHERE user_id=$1`,
		userID,
	).Scan(&remaining)

	if err == sql.ErrNoRows {
		return "", apperrors.ErrLeaveBalanceMissing
	}
	if err != nil {
		return "", errors.New("failed to fetch leave balance")
	}

	if days > remaining {
		return "", apperrors.ErrLeaveBalanceExceeded
	}

	// ---- Fetch user grade ----
	var gradeID int64
	err = tx.QueryRow(
		ctx,
		`SELECT grade_id FROM users WHERE id=$1`,
		userID,
	).Scan(&gradeID)

	if err == sql.ErrNoRows {
		return "", apperrors.ErrUserNotFound
	}
	if err != nil {
		return "", errors.New("failed to fetch user grade")
	}

	// ---- Fetch rule ----
	rule, err := GetRule("LEAVE", gradeID)
	if err != nil {
		return "", apperrors.ErrRuleNotFound
	}

	// ---- Decision ----
	decision := Decide("LEAVE", rule.Condition, float64(days))

	status := "PENDING"
	message := "Leave submitted to manager for approval"

	if decision == "AUTO_APPROVE" {
		status = "AUTO_APPROVED"
		message = "Leave approved by system"
	}

	// ---- Insert request ----
	_, err = tx.Exec(
		ctx,
		`INSERT INTO leave_requests
		 (employee_id, from_date, to_date, reason, leave_type, status, rule_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		userID, from, to, reason, leaveType, status, rule.ID,
	)

	if err != nil {
		return "", errors.New("failed to create leave request")
	}

	// ---- Deduct balance if auto-approved ----
	if status == "AUTO_APPROVED" {
		_, err = tx.Exec(
			ctx,
			`UPDATE leaves
			 SET remaining_count = remaining_count - $1
			 WHERE user_id=$2`,
			days, userID,
		)
		if err != nil {
			return "", errors.New("failed to update leave balance")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", errors.New("failed to commit transaction")
	}

	return message, nil
}

func CancelLeave(userID, requestID int64) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	var from, to time.Time

	err = tx.QueryRow(
		ctx,
		`SELECT status, from_date, to_date 
		 FROM leave_requests 
		 WHERE id=$1 AND employee_id=$2`,
		requestID, userID,
	).Scan(&status, &from, &to)

	if err != nil {
		return err
	}

	if status == "APPROVED" || status == "REJECTED" {
		return errors.New("cannot cancel finalized request")
	}

	days := utils.CalculateLeaveDays(from, to)

	_, err = tx.Exec(
		ctx,
		`UPDATE leave_requests 
		 SET status='CANCELLED' 
		 WHERE id=$1`,
		requestID,
	)
	if err != nil {
		return err
	}

	if status == "AUTO_APPROVED" {
		_, err = tx.Exec(
			ctx,
			`UPDATE leaves 
			 SET remaining_count = remaining_count + $1
			 WHERE user_id=$2`,
			days, userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
