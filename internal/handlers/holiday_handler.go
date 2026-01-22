package handlers

import (
	"strconv"
	"time"

	"rule-based-approval-engine/internal/services"

	"github.com/gin-gonic/gin"
)

type HolidayRequest struct {
	Date        string `json:"date"`
	Description string `json:"description"`
}

func AddHoliday(c *gin.Context) {
	role := c.GetString("role")
	adminID := c.GetInt64("user_id")

	var req HolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid date format"})
		return
	}

	err = services.AddHoliday(role, adminID, date, req.Description)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "holiday added"})
}

func GetHolidays(c *gin.Context) {
	role := c.GetString("role")

	holidays, err := services.GetHolidays(role)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, holidays)
}

func DeleteHoliday(c *gin.Context) {
	role := c.GetString("role")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid holiday id"})
		return
	}

	err = services.DeleteHoliday(role, id)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "holiday removed"})
}
