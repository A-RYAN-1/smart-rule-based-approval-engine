package models

type Leave struct {
	ID             int64
	UserID         int64
	LeaveType      string
	TotalAllocated int
	RemainingCount int
}
