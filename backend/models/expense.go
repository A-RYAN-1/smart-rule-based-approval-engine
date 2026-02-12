package models

type Expense struct {
	ID              int64
	UserID          int64
	ExpenseType     string
	TotalAmount     float64
	RemainingAmount float64
}
