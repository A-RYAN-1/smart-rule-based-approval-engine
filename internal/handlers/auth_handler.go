package handlers

import (
	"net/http"

	"rule-based-approval-engine/internal/services"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := services.RegisterUser(req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "user registered"})
}
func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	token, role, err := services.LoginUser(req.Email, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	// ✅ Set JWT as HttpOnly cookie
	c.SetCookie(
		"access_token", // name
		token,          // value
		3600*24,        // maxAge (1 day)
		"/",            // path
		"",             // domain (localhost ok)
		false,          // secure (true in prod HTTPS)
		true,           // httpOnly ✅
	)

	c.JSON(200, gin.H{
		"message": "login successful",
		"role":    role,
	})
}
func Logout(c *gin.Context) {
	c.SetCookie(
		"access_token",
		"",
		-1, // expire immediately
		"/",
		"",
		false,
		true,
	)

	c.JSON(200, gin.H{"message": "logged out"})
}
