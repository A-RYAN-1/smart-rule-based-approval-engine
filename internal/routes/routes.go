package routes

import (
	"rule-based-approval-engine/internal/handlers"
	"rule-based-approval-engine/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/me", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"user_id": c.GetInt64("user_id"),
				"role":    c.GetString("role"),
			})
		})
		protected.GET("/balances", handlers.GetMyBalances)
		leaves := protected.Group("/leaves")
		{
			leaves.POST("/request", handlers.ApplyLeave)
			leaves.POST("/:id/cancel", handlers.CancelLeave)
			leaves.GET("/pending", handlers.GetPendingLeaves)
			leaves.POST("/:id/approve", handlers.ApproveLeave)
			// leaves.POST("/:id/reject", handlers.RejectLeave)

		}
		expenses := protected.Group("/expenses")
		{
			expenses.POST("/request", handlers.ApplyExpense)
			expenses.POST("/:id/cancel", handlers.CancelExpense)
		}
		discounts := protected.Group("/discounts")
		{
			discounts.POST("/request", handlers.ApplyDiscount)
			discounts.POST("/:id/cancel", handlers.CancelDiscount)
		}

	}
}
