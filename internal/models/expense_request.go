package models

import "time"

type ExpenseRequest struct {
	ID           int64
	EmployeeID   int64
	Amount       float64
	Category     string
	Reason       string
	Status       string
	RuleID       *int64
	ApprovedByID *int64
	CreatedAt    time.Time
}
