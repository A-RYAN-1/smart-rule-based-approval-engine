package expense_service

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

// handles expense-related HTTP requests
type ExpenseHandler struct {
	expenseService interfaces.ExpenseService
}

// creates a new ExpenseHandler instance
func NewExpenseHandler(ctx context.Context, expenseService interfaces.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func (h *ExpenseHandler) ApplyExpense(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		handleApplyExpenseError(c, apperrors.ErrUnauthorizedUser)
		return
	}

	// 1. Initial Validation
	var req ExpenseApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleApplyExpenseError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	ctx := c.Request.Context()
	// 2. Service method calling
	message, status, err := h.expenseService.ApplyExpense(
		ctx,
		userID,
		req.Amount,
		req.Category,
		req.Reason,
	)

	if err != nil {
		handleApplyExpenseError(c, err)
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
func handleApplyExpenseError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrInvalidExpenseAmount, apperrors.ErrInvalidExpenseCategory,
		apperrors.ErrExpenseLimitExceeded, apperrors.ErrInvalidRequestPayload:
		status = http.StatusBadRequest
	case apperrors.ErrExpenseBalanceMissing, apperrors.ErrUserNotFound:
		status = http.StatusNotFound
	case apperrors.ErrUnauthorizedUser:
		status = http.StatusUnauthorized
	}

	response.Error(c, status, err.Error(), nil)
}

func (h *ExpenseHandler) CancelExpense(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 1. Validation
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleCancelExpenseError(c, apperrors.ErrInvalidID)
		return
	}

	ctx := c.Request.Context()
	// 2. Service method calling
	err = h.expenseService.CancelExpense(ctx, userID, requestID)
	if err != nil {
		handleCancelExpenseError(c, err)
		return
	}

	response.Success(c, "expense request cancelled successfully", nil)
}

func handleCancelExpenseError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrExpenseRequestNotFound:
		status = http.StatusNotFound
	case apperrors.ErrRequestCannotCancel, apperrors.ErrInvalidID:
		status = http.StatusBadRequest
	}

	response.Error(c, status, err.Error(), nil)
}

// handles expense approval-related HTTP requests
type ExpenseApprovalHandler struct {
	expenseApprovalService interfaces.ExpenseApprovalService
}

// creates a new ExpenseApprovalHandler instance
func NewExpenseApprovalHandler(ctx context.Context, expenseApprovalService interfaces.ExpenseApprovalService) *ExpenseApprovalHandler {
	return &ExpenseApprovalHandler{expenseApprovalService: expenseApprovalService}
}

func (h *ExpenseApprovalHandler) GetPendingExpenses(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt64("user_id")

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := c.Request.Context()
	expenses, total, err := h.expenseApprovalService.GetPendingExpenseRequests(ctx, role, userID, limit, offset)
	if err != nil {
		handleExpenseApprovalError(c, err)
		return
	}

	response.Success(c, "pending expense requests fetched successfully", gin.H{
		"requests": expenses,
		"total":    total,
	})
}

func (h *ExpenseApprovalHandler) ApproveExpense(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	// 1. Validation
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleExpenseApprovalError(c, apperrors.ErrInvalidID)
		return
	}

	// 2. Bind JSON
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil && err.Error() != "EOF" {
		handleExpenseApprovalError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	comment, _ := body["comment"].(string)

	ctx := c.Request.Context()
	// 3. Service method calling
	err = h.expenseApprovalService.ApproveExpense(ctx, role, approverID, requestID, comment)
	if err != nil {
		handleExpenseApprovalError(c, err)
		return
	}

	response.Success(c, "expense approved successfully", nil)
}

func (h *ExpenseApprovalHandler) RejectExpense(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	// 1. Validation
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleExpenseApprovalError(c, apperrors.ErrInvalidID)
		return
	}

	// 2. Bind JSON
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		handleExpenseApprovalError(c, apperrors.ErrInvalidRequestPayload)
		return
	}

	comment, ok := body["comment"].(string)
	if !ok || comment == "" {
		handleExpenseApprovalError(c, apperrors.ErrCommentMissing)
		return
	}

	ctx := c.Request.Context()
	// 3. Service method calling
	err = h.expenseApprovalService.RejectExpense(ctx, role, approverID, requestID, comment)
	if err != nil {
		handleExpenseApprovalError(c, err)
		return
	}

	response.Success(c, "expense rejected successfully", nil)
}

func handleExpenseApprovalError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrUnauthorizedApprover, apperrors.ErrUnauthorizedRole,
		apperrors.ErrSelfApprovalNotAllowed:
		status = http.StatusForbidden
	case apperrors.ErrExpenseRequestNotFound, apperrors.ErrUserNotFound:
		status = http.StatusNotFound
	case apperrors.ErrRequestNotPending, apperrors.ErrCommentRequired,
		apperrors.ErrCommentMissing, apperrors.ErrInvalidID,
		apperrors.ErrInvalidRequestPayload:
		status = http.StatusBadRequest
	}

	response.Error(c, status, err.Error(), nil)
}
