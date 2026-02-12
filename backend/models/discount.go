package models

type Discount struct {
	ID                int64
	UserID            int64
	DiscountType      string
	TotalDiscount     float64
	RemainingDiscount float64
}
