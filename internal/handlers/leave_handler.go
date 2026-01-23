package handlers

import (
	"net/http"
	"strconv"
	"time"

	"rule-based-approval-engine/internal/apperrors"
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
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized user",
		})
		return
	}

	var req LeaveApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request payload",
		})
		return
	}

	// ---- Date parsing with validation ----
	from, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid from_date format (YYYY-MM-DD)",
		})
		return
	}

	to, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid to_date format (YYYY-MM-DD)",
		})
		return
	}

	days := utils.CalculateLeaveDays(from, to)
	if days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "leave duration must be at least one day",
		})
		return
	}

	// ---- Call service ----
	message, err := services.ApplyLeave(
		userID,
		from,
		to,
		days,
		req.LeaveType,
		req.Reason,
	)

	if err != nil {
		handleApplyLeaveError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": message,
	})
}

func handleApplyLeaveError(c *gin.Context, err error) {

	switch err {

	case apperrors.ErrLeaveBalanceExceeded:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "insufficient leave balance",
		})

	case apperrors.ErrInvalidLeaveDays:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid leave duration",
		})

	case apperrors.ErrRuleNotFound:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "leave approval rules not configured",
		})

	case apperrors.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})

	case apperrors.ErrLeaveBalanceMissing:
		c.JSON(http.StatusNotFound, gin.H{
			"error": "leave balance not found",
		})
	case apperrors.ErrLeaveOverlap:
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to apply leave",
		})
	}
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
