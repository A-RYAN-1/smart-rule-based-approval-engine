package services

import (
	"context"
	"errors"

	"rule-based-approval-engine/internal/database"
)

func ApplyExpense(
	userID int64,
	amount float64,
	category string,
	reason string,
) error {

	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Fetch remaining expense balance
	var remaining float64
	err = tx.QueryRow(
		ctx,
		`SELECT remaining_amount FROM expense WHERE user_id=$1`,
		userID,
	).Scan(&remaining)

	if err != nil {
		return err
	}

	if amount > remaining {
		return errors.New("expense limit exceeded")
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
	rule, err := GetRule("EXPENSE", gradeID)
	if err != nil {
		return err
	}

	// 4. Decide
	decision := Decide("EXPENSE", rule.Condition, amount)

	status := "PENDING"
	if decision == "AUTO_APPROVE" {
		status = "AUTO_APPROVED"
	}

	// 5. Insert expense request
	var requestID int64
	err = tx.QueryRow(
		ctx,
		`INSERT INTO expense_requests
		 (employee_id, amount, category, reason, status, rule_id)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id`,
		userID, amount, category, reason, status, rule.ID,
	).Scan(&requestID)

	if err != nil {
		return err
	}

	// 6. Deduct balance if auto-approved
	if status == "AUTO_APPROVED" {
		_, err = tx.Exec(
			ctx,
			`UPDATE expense
			 SET remaining_amount = remaining_amount - $1
			 WHERE user_id=$2`,
			amount, userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
func CancelExpense(userID, requestID int64) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	var amount float64

	err = tx.QueryRow(
		ctx,
		`SELECT status, amount
		 FROM expense_requests
		 WHERE id=$1 AND employee_id=$2`,
		requestID, userID,
	).Scan(&status, &amount)

	if err != nil {
		return err
	}

	if status == "APPROVED" || status == "REJECTED" {
		return errors.New("cannot cancel finalized request")
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE expense_requests
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
			`UPDATE expense
			 SET remaining_amount = remaining_amount + $1
			 WHERE user_id=$2`,
			amount, userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
