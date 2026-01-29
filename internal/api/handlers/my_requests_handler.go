package handlers

import (
	"net/http"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"
	"rule-based-approval-engine/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type MyRequestsHandler struct {
	myRequestsService *services.MyRequestsService
}

func NewMyRequestsHandler(myRequestsService *services.MyRequestsService) *MyRequestsHandler {
	return &MyRequestsHandler{myRequestsService: myRequestsService}
}

func (h *MyRequestsHandler) GetMyLeaves(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized user", nil)
		return
	}

	ctx := c.Request.Context()
	data, err := h.myRequestsService.GetMyLeaveRequests(ctx, userID)
	if err != nil {
		handleRequestError(c, err, "failed to fetch leave requests")
		return
	}

	response.Success(
		c,
		"leave requests fetched successfully",
		data,
	)
}

func (h *MyRequestsHandler) GetMyExpenses(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized user", nil)
		return
	}

	ctx := c.Request.Context()
	data, err := h.myRequestsService.GetMyExpenseRequests(ctx, userID)
	if err != nil {
		handleRequestError(c, err, "failed to fetch expense requests")
		return
	}

	response.Success(
		c,
		"expense requests fetched successfully",
		data,
	)
}

func (h *MyRequestsHandler) GetMyDiscounts(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized user", nil)
		return
	}

	ctx := c.Request.Context()
	data, err := h.myRequestsService.GetMyDiscountRequests(ctx, userID)
	if err != nil {
		handleRequestError(c, err, "failed to fetch discount requests")
		return
	}

	response.Success(
		c,
		"discount requests fetched successfully",
		data,
	)
}

func handleRequestError(c *gin.Context, err error, message string) {
	status := http.StatusInternalServerError

	if err == apperrors.ErrUserNotFound {
		status = http.StatusNotFound
	}

	response.Error(c, status, message, err.Error())
}
