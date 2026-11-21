package models

import "time"

type LeaveRequest struct {
	ID           int64
	EmployeeID   int64
	FromDate     time.Time
	ToDate       time.Time
	Days         int
	LeaveType    string
	Reason       string
	Status       string
	ApprovedByID *int64
	RuleID       *int64
	RuleLimit    *float64
	CreatedAt    time.Time
}
