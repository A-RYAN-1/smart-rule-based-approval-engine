package handlers

import (
	"rule-based-approval-engine/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DiscountApplyRequest struct {
	DiscountPercentage float64 `json:"discount_percentage"`
	Reason             string  `json:"reason"`
}

func ApplyDiscount(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req DiscountApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := services.ApplyDiscount(userID, req.DiscountPercentage, req.Reason)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "discount request submitted"})
}
func CancelDiscount(c *gin.Context) {
	userID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request id"})
		return
	}

	err = services.CancelDiscount(userID, requestID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "discount request cancelled"})
}
