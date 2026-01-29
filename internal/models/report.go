package models

type RequestTypeReport struct {
	Type                string  `json:"type"`
	TotalRequests       int     `json:"total_requests"`
	AutoApproved        int     `json:"auto_approved"`
	AutoApprovedPercent float64 `json:"auto_approved_percentage"`
}
