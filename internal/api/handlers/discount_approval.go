package handlers

import (
	"net/http"
	"strconv"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"
	"rule-based-approval-engine/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type DiscountApprovalHandler struct {
	discountService *services.DiscountService
}

func NewDiscountApprovalHandler(discountService *services.DiscountService) *DiscountApprovalHandler {
	return &DiscountApprovalHandler{discountService: discountService}
}

func (h *DiscountApprovalHandler) GetPendingDiscounts(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt64("user_id")

	ctx := c.Request.Context()
	discounts, err := h.discountService.GetPendingDiscountRequests(ctx, role, userID)
	if err != nil {
		handleApproveRejectDiscountError(c, err, "failed to fetch pending discount requests")
		return
	}

	response.Success(
		c,
		"pending discount requests fetched successfully",
		discounts,
	)
}

func (h *DiscountApprovalHandler) ApproveDiscount(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid discount request id", nil)
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && err.Error() != "EOF" {
		response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	comment, _ := body["comment"].(string)

	ctx := c.Request.Context()
	err = h.discountService.ApproveDiscount(ctx, role, approverID, requestID, comment)
	if err != nil {
		handleApproveRejectDiscountError(c, err, "unable to approve discount request")
		return
	}

	response.Success(
		c,
		"discount request approved successfully",
		nil,
	)
}

func (h *DiscountApprovalHandler) RejectDiscount(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid discount request id", nil)
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	comment, ok := body["comment"].(string)
	if !ok || comment == "" {
		response.Error(c, http.StatusBadRequest, "comment is required", nil)
		return
	}

	ctx := c.Request.Context()
	err = h.discountService.RejectDiscount(ctx, role, approverID, requestID, comment)
	if err != nil {
		handleApproveRejectDiscountError(c, err, "unable to reject discount request")
		return
	}

	response.Success(
		c,
		"discount request rejected successfully",
		nil,
	)
}

func handleApproveRejectDiscountError(c *gin.Context, err error, message string) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrUnauthorizedApprover, apperrors.ErrUnauthorizedRole, apperrors.ErrSelfApprovalNotAllowed:
		status = http.StatusForbidden
	case apperrors.ErrDiscountRequestNotFound:
		status = http.StatusNotFound
	case apperrors.ErrDiscountRequestNotPending, apperrors.ErrCommentRequired:
		status = http.StatusBadRequest
	}

	response.Error(c, status, message, err.Error())
}
