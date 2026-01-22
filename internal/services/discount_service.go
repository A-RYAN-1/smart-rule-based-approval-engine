package services

import (
	"context"
	"errors"

	"rule-based-approval-engine/internal/database"
)

func ApplyDiscount(
	userID int64,
	percent float64,
	reason string,
) error {

	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Fetch remaining discount
	var remaining float64
	err = tx.QueryRow(
		ctx,
		`SELECT remaining_discount FROM discount WHERE user_id=$1`,
		userID,
	).Scan(&remaining)

	if err != nil {
		return err
	}

	if percent > remaining {
		return errors.New("discount limit exceeded")
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
	rule, err := GetRule("DISCOUNT", gradeID)
	if err != nil {
		return err
	}

	// 4. Decide
	decision := Decide("DISCOUNT", rule.Condition, percent)

	status := "PENDING"
	if decision == "AUTO_APPROVE" {
		status = "AUTO_APPROVED"
	}

	// 5. Insert discount request
	var requestID int64
	err = tx.QueryRow(
		ctx,
		`INSERT INTO discount_requests
		 (employee_id, discount_percentage, reason, status, rule_id)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id`,
		userID, percent, reason, status, rule.ID,
	).Scan(&requestID)

	if err != nil {
		return err
	}

	// 6. Deduct discount if auto-approved
	if status == "AUTO_APPROVED" {
		_, err = tx.Exec(
			ctx,
			`UPDATE discount
			 SET remaining_discount = remaining_discount - $1
			 WHERE user_id=$2`,
			percent, userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
func CancelDiscount(userID, requestID int64) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	var percent float64

	err = tx.QueryRow(
		ctx,
		`SELECT status, discount_percentage
		 FROM discount_requests
		 WHERE id=$1 AND employee_id=$2`,
		requestID, userID,
	).Scan(&status, &percent)

	if err != nil {
		return err
	}

	if status == "APPROVED" || status == "REJECTED" {
		return errors.New("cannot cancel finalized request")
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE discount_requests
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
			`UPDATE discount
			 SET remaining_discount = remaining_discount + $1
			 WHERE user_id=$2`,
			percent, userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
