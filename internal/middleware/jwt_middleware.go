package middleware

import (
	"rule-based-approval-engine/internal/utils"

	"github.com/gin-gonic/gin"
)

var jwtSecret = []byte("super-secret-key")

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		// ✅ Read token from HttpOnly cookie
		tokenString, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set values in context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
