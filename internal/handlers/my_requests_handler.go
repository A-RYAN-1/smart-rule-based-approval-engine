package handlers

import (
	"rule-based-approval-engine/internal/services"

	"github.com/gin-gonic/gin"
)

func GetMyLeaves(c *gin.Context) {
	userID := c.GetInt64("user_id")

	data, err := services.GetMyLeaveRequests(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch leaves"})
		return
	}

	c.JSON(200, data)
}
func GetMyExpenses(c *gin.Context) {
	userID := c.GetInt64("user_id")

	data, err := services.GetMyExpenseRequests(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch expenses"})
		return
	}

	c.JSON(200, data)
}
func GetMyDiscounts(c *gin.Context) {
	userID := c.GetInt64("user_id")

	data, err := services.GetMyDiscountRequests(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch discounts"})
		return
	}

	c.JSON(200, data)
}
