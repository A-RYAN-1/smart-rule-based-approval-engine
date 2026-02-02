package handlers

import (
	"net/http"
	"strconv"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"
	"rule-based-approval-engine/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// handles leave approval-related HTTP requests
type LeaveApprovalHandler struct {
	leaveApprovalService *services.LeaveApprovalService
}

// creates a new LeaveApprovalHandler instance
func NewLeaveApprovalHandler(leaveApprovalService *services.LeaveApprovalService) *LeaveApprovalHandler {
	return &LeaveApprovalHandler{leaveApprovalService: leaveApprovalService}
}

func (h *LeaveApprovalHandler) GetPendingLeaves(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt64("user_id")

	ctx := c.Request.Context()
	leaves, err := h.leaveApprovalService.GetPendingLeaveRequests(ctx, role, userID)
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to fetch pending leave requests",
			err.Error(),
		)
		return
	}

	response.Success(
		c,
		"pending leave requests fetched successfully",
		leaves,
	)
}

func (h *LeaveApprovalHandler) ApproveLeave(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid leave request id",
			nil,
		)
		return
	}

	// Read request body as map
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid request body",
			err.Error(),
		)
		return
	}

	approvalComment, _ := body["comment"].(string)

	ctx := c.Request.Context()
	err = h.leaveApprovalService.ApproveLeave(
		ctx,
		role,
		approverID,
		requestID,
		approvalComment,
	)
	if err != nil {
		handleApprovalError(c, err, "unable to approve leave request")
		return
	}

	response.Success(
		c,
		"leave approved successfully",
		nil,
	)
}

func (h *LeaveApprovalHandler) RejectLeave(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid leave request id",
			nil,
		)
		return
	}

	// Read request body as map
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid request body",
			err.Error(),
		)
		return
	}

	rejectionComment, ok := body["comment"].(string)
	if !ok || rejectionComment == "" {
		response.Error(
			c,
			http.StatusBadRequest,
			"rejection comment is required",
			nil,
		)
		return
	}

	ctx := c.Request.Context()
	err = h.leaveApprovalService.RejectLeave(
		ctx,
		role,
		approverID,
		requestID,
		rejectionComment,
	)
	if err != nil {
		handleApprovalError(c, err, "unable to reject leave request")
		return
	}

	response.Success(
		c,
		"leave rejected successfully",
		nil,
	)
}

func handleApprovalError(c *gin.Context, err error, message string) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrUnauthorizedApprover, apperrors.ErrUnauthorizedRole, apperrors.ErrSelfApprovalNotAllowed:
		status = http.StatusForbidden
	case apperrors.ErrLeaveRequestNotFound, apperrors.ErrUserNotFound:
		status = http.StatusNotFound
	case apperrors.ErrRequestNotPending, apperrors.ErrCommentRequired:
		status = http.StatusBadRequest
	}

	response.Error(c, status, message, err.Error())
}
