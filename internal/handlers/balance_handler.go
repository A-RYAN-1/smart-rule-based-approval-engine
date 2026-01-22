package handlers

import (
	"context"

	"rule-based-approval-engine/internal/database"

	"github.com/gin-gonic/gin"
)

func GetMyBalances(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var leaveTotal, leaveRemaining int
	var expenseTotal, expenseRemaining float64
	var discountTotal, discountRemaining float64

	err := database.DB.QueryRow(
		context.Background(),
		`SELECT total_allocated, remaining_count FROM leaves WHERE user_id=$1`,
		userID,
	).Scan(&leaveTotal, &leaveRemaining)
	if err != nil {
		c.JSON(500, gin.H{"error": "leave fetch failed"})
		return
	}
	err = database.DB.QueryRow(
		context.Background(),
		`SELECT total_amount, remaining_amount FROM expense WHERE user_id=$1`,
		userID,
	).Scan(&expenseTotal, &expenseRemaining)
	if err != nil {
		c.JSON(500, gin.H{"error": "expense fetch failed"})
		return
	}

	err = database.DB.QueryRow(
		context.Background(),
		`SELECT total_discount, remaining_discount FROM discount WHERE user_id=$1`,
		userID,
	).Scan(&discountTotal, &discountRemaining)
	if err != nil {
		c.JSON(500, gin.H{"error": "discount fetch failed"})
		return
	}

	c.JSON(200, gin.H{
		"leave": gin.H{
			"total":     leaveTotal,
			"remaining": leaveRemaining,
		},
		"expense": gin.H{
			"total":     expenseTotal,
			"remaining": expenseRemaining,
		},
		"discount": gin.H{
			"total":     discountTotal,
			"remaining": discountRemaining,
		},
	})
}
