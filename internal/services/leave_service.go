package services

import (
	"context"
	"errors"
	"time"

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
) error {

	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Fetch remaining leave balance
	var remaining int
	err = tx.QueryRow(
		ctx,
		`SELECT remaining_count FROM leaves WHERE user_id=$1`,
		userID,
	).Scan(&remaining)

	if err != nil {
		return err
	}

	if days > remaining {
		return errors.New("leave balance exceeded")
	}

	// 2. Fetch user grade
	var gradeID int64
	err = tx.QueryRow(
		ctx,
		`SELECT grade_id FROM users WHERE id=$1`,
		userID,
	).Scan(&gradeID)

	if err != nil {
		return err
	}

	// 3. Fetch rule
	rule, err := GetRule("LEAVE", gradeID)
	if err != nil {
		return err
	}

	// 4. Decide
	decision := Decide("LEAVE", rule.Condition, float64(days))

	status := "PENDING"
	if decision == "AUTO_APPROVE" {
		status = "AUTO_APPROVED"
	}

	// 5. Insert leave request
	var requestID int64
	err = tx.QueryRow(
		ctx,
		`INSERT INTO leave_requests 
		 (employee_id, from_date, to_date, reason, leave_type, status, rule_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 RETURNING id`,
		userID, from, to, reason, leaveType, status, rule.ID,
	).Scan(&requestID)

	if err != nil {
		return err
	}

	// 6. Deduct balance if auto-approved
	if status == "AUTO_APPROVED" {
		_, err = tx.Exec(
			ctx,
			`UPDATE leaves 
			 SET remaining_count = remaining_count - $1
			 WHERE user_id=$2`,
			days, userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
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
