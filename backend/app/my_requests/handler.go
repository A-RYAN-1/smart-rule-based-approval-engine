package my_requests

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

type MyRequestsHandler struct {
	myRequestsService interfaces.MyRequestsService
}

func NewMyRequestsHandler(ctx context.Context, myRequestsService interfaces.MyRequestsService) *MyRequestsHandler {
	return &MyRequestsHandler{myRequestsService: myRequestsService}
}

func (h *MyRequestsHandler) GetMyRequests(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		handleRequestError(c, apperrors.ErrUnauthorizedUser, "failed to fetch requests")
		return
	}

	reqType := c.Query("type") // Expected: LEAVE, EXPENSE, DISCOUNT, etc.
	if reqType == "" {
		handleRequestError(c, apperrors.ErrInvalidRequestPayload, "request type is required")
		return
	}

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
	data, total, err := h.myRequestsService.GetMyRequests(ctx, userID, reqType, limit, offset)
	if err != nil {
		handleRequestError(c, err, "failed to fetch requests")
		return
	}

	response.Success(c, "requests fetched successfully", gin.H{
		"requests": data,
		"total":    total,
	})
}

func (h *MyRequestsHandler) GetMyAllRequests(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		handleRequestError(c, apperrors.ErrUnauthorizedUser, "failed to fetch requests")
		return
	}

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
	data, err := h.myRequestsService.GetMyAllRequests(ctx, userID, limit, offset)
	if err != nil {
		handleRequestError(c, err, "failed to fetch requests")
		return
	}

	response.Success(c, "requests fetched successfully", data)
}
func (h *MyRequestsHandler) GetPendingAllRequests(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt64("user_id")
	if userID == 0 {
		handleRequestError(c, apperrors.ErrUnauthorizedUser, "failed to fetch requests")
		return
	}

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
	data, err := h.myRequestsService.GetPendingAllRequests(ctx, role, userID, limit, offset)
	if err != nil {
		handleRequestError(c, err, "failed to fetch requests")
		return
	}

	response.Success(c, "pending requests fetched successfully", data)
}

func handleRequestError(c *gin.Context, err error, message string) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrUserNotFound:
		status = http.StatusNotFound
	case apperrors.ErrUnauthorizedUser:
		status = http.StatusUnauthorized
	case apperrors.ErrInvalidRequestPayload:
		status = http.StatusBadRequest
	}

	response.Error(c, status, message, err.Error())
}

// Helper methods for specific routes
func (h *MyRequestsHandler) GetMyLeaves(c *gin.Context) {
	c.Request.URL.RawQuery = "type=LEAVE"
	h.GetMyRequests(c)
}

func (h *MyRequestsHandler) GetMyExpenses(c *gin.Context) {
	c.Request.URL.RawQuery = "type=EXPENSE"
	h.GetMyRequests(c)
}

func (h *MyRequestsHandler) GetMyDiscounts(c *gin.Context) {
	c.Request.URL.RawQuery = "type=DISCOUNT"
	h.GetMyRequests(c)
}
