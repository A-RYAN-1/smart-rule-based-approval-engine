package handlers

import (
	"strconv"
	"time"

	"rule-based-approval-engine/internal/services"
	"rule-based-approval-engine/internal/utils"

	"github.com/gin-gonic/gin"
)

type LeaveApplyRequest struct {
	FromDate  string `json:"from_date"`
	ToDate    string `json:"to_date"`
	LeaveType string `json:"leave_type"`
	Reason    string `json:"reason"`
}

func ApplyLeave(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req LeaveApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	from, _ := time.Parse("2006-01-02", req.FromDate)
	to, _ := time.Parse("2006-01-02", req.ToDate)

	days := utils.CalculateLeaveDays(from, to)

	err := services.ApplyLeave(userID, from, to, days, req.LeaveType, req.Reason)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "leave request submitted"})
}
func CancelLeave(c *gin.Context) {
	userID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request id"})
		return
	}

	err = services.CancelLeave(userID, requestID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "leave request cancelled"})
}
