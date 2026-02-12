package domain_service

import (
	"context"
	"net/http"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

type BalanceHandler struct {
	balanceService interfaces.BalanceService
}

func NewBalanceHandler(ctx context.Context, balanceService interfaces.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) GetBalances(c *gin.Context) {
	userID := c.GetInt64("user_id")
	ctx := c.Request.Context()

	balances, err := h.balanceService.GetBalances(ctx, userID)
	if err != nil {
		handleBalanceError(c, err)
		return
	}

	response.Success(
		c,
		"balances fetched successfully",
		balances,
	)
}

func handleBalanceError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrUserNotFound:
		status = http.StatusNotFound
	}

	response.Error(c, status, err.Error(), nil)
}
