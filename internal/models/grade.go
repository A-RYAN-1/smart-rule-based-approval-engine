package models

type Grade struct {
	ID                   int64
	Name                 string
	AnnualLeaveLimit     int
	AnnualExpenseLimit   float64
	DiscountLimitPercent float64
}
