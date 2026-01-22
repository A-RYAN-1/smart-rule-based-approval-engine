package handlers

import (
	"strconv"

	"rule-based-approval-engine/internal/services"

	"github.com/gin-gonic/gin"
)

func GetPendingLeaves(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetInt64("user_id")

	leaves, err := services.GetPendingLeaveRequests(role, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, leaves)
}
func ApproveLeave(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := services.ApproveLeave(role, approverID, requestID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "leave approved"})
}
func RejectLeave(c *gin.Context) {
	role := c.GetString("role")
	approverID := c.GetInt64("user_id")

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request id"})
		return
	}

	err = services.RejectLeave(role, approverID, requestID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "leave rejected"})
}
