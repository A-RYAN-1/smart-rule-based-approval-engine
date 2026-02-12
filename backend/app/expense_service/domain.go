package expense_service

type ExpenseApplyRequest struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Reason   string  `json:"reason"`
}
