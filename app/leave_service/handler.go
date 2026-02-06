package leave_service

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"

	"github.com/gin-gonic/gin"
)

// handles leave-related HTTP requests
type LeaveHandler struct {
	leaveService interfaces.LeaveService
}

// creates a new LeaveHandler instance
func NewLeaveHandler(ctx context.Context, leaveService interfaces.LeaveService) *LeaveHandler {
	return &LeaveHandler{leaveService: leaveService}
}

func (h *LeaveHandler) ApplyLeave(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		handleApplyLeaveError(c, apperrors.ErrUnauthorizedUser)
		return
	}

	var req LeaveApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleApplyLeaveError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	from, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		handleApplyLeaveError(c, apperrors.ErrInvalidDateFormat)
		return
	}

	to, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		handleApplyLeaveError(c, apperrors.ErrInvalidDateFormat)
		return
	}

	days := utils.CalculateLeaveDays(from, to)
	if days <= 0 {
		handleApplyLeaveError(c, apperrors.ErrInvalidLeaveDays)
		return
	}

	ctx := c.Request.Context()
	message, status, err := h.leaveService.ApplyLeave(
		ctx, userID, from, to, days, req.LeaveType, req.Reason,
	)

	if err != nil {
		handleApplyLeaveError(c, err)
		return
	}

	response.Success(c, message, gin.H{
		"status": status,
	})
}

func handleApplyLeaveError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch err {
	case apperrors.ErrLeaveBalanceExceeded, apperrors.ErrInvalidLeaveDays,
		apperrors.ErrLeaveOverlap, apperrors.ErrPastDate,
		apperrors.ErrInvalidDateFormat, apperrors.ErrInvalidRequestPayload:
		status = http.StatusBadRequest
	case apperrors.ErrUserNotFound, apperrors.ErrLeaveBalanceMissing:
		status = http.StatusNotFound
	case apperrors.ErrUnauthorizedUser:
		status = http.StatusUnauthorized
	}

	response.Error(c, status, err.Error(), nil)
}

func (h *LeaveHandler) CancelLeave(c *gin.Context) {
	userID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleCancelLeaveError(c, apperrors.ErrInvalidID)
		return
	}

	ctx := c.Request.Context()
	err = h.leaveService.CancelLeave(ctx, userID, requestID)
	if err != nil {
		handleCancelLeaveError(c, err)
		return
	}

	response.Success(c, "leave request cancelled successfully", nil)
}

func handleCancelLeaveError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrLeaveRequestNotFound:
		status = http.StatusNotFound
	case apperrors.ErrRequestCannotCancel, apperrors.ErrInvalidID:
		status = http.StatusBadRequest
	}

	response.Error(c, status, err.Error(), nil)
}

// handles leave approval-related HTTP requests
type LeaveApprovalHandler struct {
	leaveApprovalService interfaces.LeaveApprovalService
}

// creates a new LeaveApprovalHandler instance
func NewLeaveApprovalHandler(ctx context.Context, leaveApprovalService interfaces.LeaveApprovalService) *LeaveApprovalHandler {
	return &LeaveApprovalHandler{leaveApprovalService: leaveApprovalService}
}

func (h *LeaveApprovalHandler) GetPendingLeaves(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt64("user_id")

	ctx := c.Request.Context()
	leaves, err := h.leaveApprovalService.GetPendingLeaveRequests(ctx, role, userID)
	if err != nil {
		handleApprovalError(c, err)
		return
	}

	response.Success(c, "pending leave requests fetched successfully", leaves)
}

func (h *LeaveApprovalHandler) ApproveLeave(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleApprovalError(c, apperrors.ErrInvalidID)
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && err.Error() != "EOF" {
		handleApprovalError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	approvalComment, _ := body["comment"].(string)

	ctx := c.Request.Context()
	err = h.leaveApprovalService.ApproveLeave(ctx, role, approverID, requestID, approvalComment)
	if err != nil {
		handleApprovalError(c, err)
		return
	}

	response.Success(c, "leave approved successfully", nil)
}

func (h *LeaveApprovalHandler) RejectLeave(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleApprovalError(c, apperrors.ErrInvalidID)
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		handleApprovalError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	rejectionComment, ok := body["comment"].(string)
	if !ok || rejectionComment == "" {
		handleApprovalError(c, apperrors.ErrCommentMissing)
		return
	}

	ctx := c.Request.Context()
	err = h.leaveApprovalService.RejectLeave(ctx, role, approverID, requestID, rejectionComment)
	if err != nil {
		handleApprovalError(c, err)
		return
	}

	response.Success(c, "leave rejected successfully", nil)
}

func handleApprovalError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrUnauthorizedApprover, apperrors.ErrUnauthorizedRole, apperrors.ErrSelfApprovalNotAllowed:
		status = http.StatusForbidden
	case apperrors.ErrLeaveRequestNotFound, apperrors.ErrUserNotFound:
		status = http.StatusNotFound
	case apperrors.ErrRequestNotPending, apperrors.ErrCommentRequired,
		apperrors.ErrCommentMissing, apperrors.ErrInvalidID,
		apperrors.ErrInvalidRequestPayload:
		status = http.StatusBadRequest
	}

	response.Error(c, status, err.Error(), nil)
}
