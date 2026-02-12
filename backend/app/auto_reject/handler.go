package auto_reject

import (
	"context"
	"net/http"

	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	autoRejectService *AutoRejectService
}

func NewSystemHandler(ctx context.Context, autoRejectService *AutoRejectService) *SystemHandler {
	return &SystemHandler{autoRejectService: autoRejectService}
}

func (h *SystemHandler) RunAutoReject(c *gin.Context) {
	ctx := c.Request.Context()
	err1 := h.autoRejectService.AutoRejectLeaveRequests(ctx)
	err2 := h.autoRejectService.AutoRejectExpenseRequests(ctx)
	err3 := h.autoRejectService.AutoRejectDiscountRequests(ctx)

	if err1 != nil {
		handleAutoRejectError(c, err1)
		return
	}
	if err2 != nil {
		handleAutoRejectError(c, err2)
		return
	}
	if err3 != nil {
		handleAutoRejectError(c, err3)
		return
	}

	response.Success(
		c,
		"auto reject executed successfully",
		nil,
	)
}

func handleAutoRejectError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	// More specific mappings can be added here if needed
	response.Error(c, status, err.Error(), nil)
}
