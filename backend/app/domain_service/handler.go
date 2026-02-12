package domain_service

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

type DiscountHandler struct {
	discountService interfaces.DiscountService
}

func NewDiscountHandler(ctx context.Context, discountService interfaces.DiscountService) *DiscountHandler {
	return &DiscountHandler{discountService: discountService}
}

func (h *DiscountHandler) ApplyDiscount(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		handleApplyDiscountError(c, apperrors.ErrUnauthorizedUser)
		return
	}

	var req DiscountApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleApplyDiscountError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	ctx := c.Request.Context()
	message, status, err := h.discountService.ApplyDiscount(
		ctx,
		userID,
		req.DiscountPercentage,
		req.Reason,
	)

	if err != nil {
		handleApplyDiscountError(c, err)
		return
	}

	response.Created(
		c,
		message,
		gin.H{
			"status": status,
		},
	)
}

func handleApplyDiscountError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrInvalidDiscountPercent, apperrors.ErrDiscountLimitExceeded,
		apperrors.ErrInvalidRequestPayload:
		status = http.StatusBadRequest
	case apperrors.ErrDiscountBalanceMissing, apperrors.ErrUserNotFound:
		status = http.StatusNotFound
	case apperrors.ErrUnauthorizedUser:
		status = http.StatusUnauthorized
	}

	response.Error(c, status, err.Error(), nil)
}

func (h *DiscountHandler) CancelDiscount(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		handleCancelDiscountError(c, apperrors.ErrUnauthorizedUser)
		return
	}

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleCancelDiscountError(c, apperrors.ErrInvalidID)
		return
	}

	ctx := c.Request.Context()
	err = h.discountService.CancelDiscount(ctx, userID, requestID)
	if err != nil {
		handleCancelDiscountError(c, err)
		return
	}

	response.Success(c, "discount request cancelled successfully", nil)
}

func handleCancelDiscountError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrDiscountRequestNotFound:
		status = http.StatusNotFound
	case apperrors.ErrDiscountCannotCancel, apperrors.ErrInvalidID:
		status = http.StatusBadRequest
	case apperrors.ErrUnauthorizedUser:
		status = http.StatusUnauthorized
	}

	response.Error(c, status, err.Error(), nil)
}

type DiscountApprovalHandler struct {
	discountApprovalService interfaces.DiscountApprovalService
}

func NewDiscountApprovalHandler(ctx context.Context, discountApprovalService interfaces.DiscountApprovalService) *DiscountApprovalHandler {
	return &DiscountApprovalHandler{discountApprovalService: discountApprovalService}
}

func (h *DiscountApprovalHandler) GetPendingDiscounts(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt64("user_id")

	ctx := c.Request.Context()
	discounts, err := h.discountApprovalService.GetPendingRequests(ctx, role, userID)
	if err != nil {
		handleApproveRejectDiscountError(c, err)
		return
	}

	response.Success(c, "pending discount requests fetched successfully", discounts)
}

func (h *DiscountApprovalHandler) ApproveDiscount(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleApproveRejectDiscountError(c, apperrors.ErrInvalidID)
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && err.Error() != "EOF" {
		handleApproveRejectDiscountError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	comment, _ := body["comment"].(string)

	ctx := c.Request.Context()
	err = h.discountApprovalService.ApproveDiscount(ctx, role, approverID, requestID, comment)
	if err != nil {
		handleApproveRejectDiscountError(c, err)
		return
	}

	response.Success(c, "discount request approved successfully", nil)
}

func (h *DiscountApprovalHandler) RejectDiscount(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleApproveRejectDiscountError(c, apperrors.ErrInvalidID)
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		handleApproveRejectDiscountError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	comment, ok := body["comment"].(string)
	if !ok || comment == "" {
		handleApproveRejectDiscountError(c, apperrors.ErrCommentMissing)
		return
	}

	ctx := c.Request.Context()
	err = h.discountApprovalService.RejectDiscount(ctx, role, approverID, requestID, comment)
	if err != nil {
		handleApproveRejectDiscountError(c, err)
		return
	}

	response.Success(c, "discount request rejected successfully", nil)
}

func handleApproveRejectDiscountError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrUnauthorizedApprover, apperrors.ErrUnauthorizedRole,
		apperrors.ErrSelfApprovalNotAllowed:
		status = http.StatusForbidden
	case apperrors.ErrDiscountRequestNotFound:
		status = http.StatusNotFound
	case apperrors.ErrDiscountRequestNotPending, apperrors.ErrCommentRequired,
		apperrors.ErrCommentMissing, apperrors.ErrInvalidID,
		apperrors.ErrInvalidRequestPayload:
		status = http.StatusBadRequest
	}

	response.Error(c, status, err.Error(), nil)
}
