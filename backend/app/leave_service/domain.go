package leave_service

type LeaveApplyRequest struct {
	FromDate  string `json:"from_date"`
	ToDate    string `json:"to_date"`
	LeaveType string `json:"leave_type"`
	Reason    string `json:"reason"`
}
