package handlers

import (
	"net/http"

	"rule-based-approval-engine/internal/services"

	"github.com/gin-gonic/gin"
)

func RunAutoReject(c *gin.Context) {
	services.AutoRejectLeaveRequests()
	services.AutoRejectExpenseRequests()
	services.AutoRejectDiscountRequests()

	c.JSON(http.StatusOK, gin.H{
		"message": "auto reject executed",
	})
}
