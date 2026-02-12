package models

import "time"

type DiscountRequest struct {
	ID                 int64
	EmployeeID         int64
	DiscountPercentage float64
	Reason             string
	Status             string
	RuleID             *int64
	ApprovedByID       *int64
	CreatedAt          time.Time
}
