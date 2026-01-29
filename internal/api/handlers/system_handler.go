package handlers

import (
	"net/http"
	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	autoRejectService *services.AutoRejectService
}

func NewSystemHandler(autoRejectService *services.AutoRejectService) *SystemHandler {
	return &SystemHandler{autoRejectService: autoRejectService}
}

func (h *SystemHandler) RunAutoReject(c *gin.Context) {
	ctx := c.Request.Context()
	err1 := h.autoRejectService.AutoRejectLeaveRequests(ctx)
	err2 := h.autoRejectService.AutoRejectExpenseRequests(ctx)
	err3 := h.autoRejectService.AutoRejectDiscountRequests(ctx)

	if err1 != nil || err2 != nil || err3 != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"auto reject failed",
			"one or more rejection processes encountered an error",
		)
		return
	}

	response.Success(
		c,
		"auto reject executed successfully",
		nil,
	)
}
