package domain_service

type DiscountApplyRequest struct {
	DiscountPercentage float64 `json:"discount_percentage"`
	Reason             string  `json:"reason"`
}
