package handlers

import (
	"net/http"
	"strconv"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type MyRequestsHandlers struct {
	service services.MyRequestsServices
}

func NewMyRequestsHandlers(service services.MyRequestsServices) *MyRequestsHandlers {
	return &MyRequestsHandlers{service: service}
}

func (h *MyRequestsHandlers) GetMyAllRequests(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
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
	data, err := h.service.GetMyAllRequests(ctx, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to fetch requests", err.Error())
		return
	}

	response.Success(c, "requests fetched successfully", data)
}
