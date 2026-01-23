package apperrors

import "errors"

var (
	ErrLeaveBalanceExceeded = errors.New("leave balance exceeded")
	ErrUserNotFound         = errors.New("user not found")
	ErrLeaveBalanceMissing  = errors.New("leave balance not found")
	ErrRuleNotFound         = errors.New("approval rule not configured")
	ErrInvalidLeaveDays     = errors.New("invalid leave days")
	ErrLeaveOverlap         = errors.New(
		"you already have a leave request for this date. first cancel the previous request then only you are allowed to apply",
	)
)
