package handlers

import (
	"log"
	"net/http"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type BalanceHandler struct {
	balanceService *services.BalanceService
}

func NewBalanceHandler(balanceService *services.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) GetMyBalances(c *gin.Context) {
	userID := c.GetInt64("user_id")
	ctx := c.Request.Context()

	balances, err := h.balanceService.GetMyBalances(ctx, userID)
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to fetch balances",
			err.Error(),
		)
		log.Printf("Error fetching balances: %v", err)
		return
	}

	response.Success(
		c,
		"balances fetched successfully",
		balances,
	)
}
