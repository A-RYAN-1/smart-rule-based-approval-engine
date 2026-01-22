package handlers

import (
	"rule-based-approval-engine/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExpenseApplyRequest struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Reason   string  `json:"reason"`
}

func ApplyExpense(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req ExpenseApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := services.ApplyExpense(userID, req.Amount, req.Category, req.Reason)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "expense request submitted"})
}
func CancelExpense(c *gin.Context) {
	userID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request id"})
		return
	}

	err = services.CancelExpense(userID, requestID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "expense request cancelled"})
}
