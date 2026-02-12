package auth

import (
	"context"
	"net/http"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

// handles authentication-related HTTP requests
type AuthHandler struct {
	authService interfaces.AuthService
}

// creates a new AuthHandler instance
func NewAuthHandler(ctx context.Context, authService interfaces.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleAuthError(c, apperrors.ErrInvalidInput, err)
		return
	}

	ctx := c.Request.Context()
	err := h.authService.RegisterUser(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		handleAuthError(c, err, nil)
		return
	}

	response.Created(c, "user registered successfully", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleAuthError(c, apperrors.ErrInvalidInput, err)
		return
	}

	ctx := c.Request.Context()
	token, role, err := h.authService.LoginUser(ctx, req.Email, req.Password)
	if err != nil {
		handleAuthError(c, err, nil)
		return
	}

	c.SetCookie(
		"access_token",
		token,
		3600, // 1 hour
		"/",
		"",
		false,
		true,
	)

	// Fetch user details to return in response
	user, err := h.authService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Fallback if user fetch fails, though login succeeded
		response.Success(
			c,
			"login successful",
			gin.H{
				"token": token,
				"role":  role,
			},
		)
		return
	}

	response.Success(
		c,
		"login successful",
		gin.H{
			"token": token,
			"user":  user,
		},
	)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		handleAuthError(c, apperrors.ErrUnauthorized, nil)
		return
	}

	ctx := c.Request.Context()
	user, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		handleAuthError(c, err, nil)
		return
	}

	response.Success(c, "user details fetched successfully", user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	response.Success(c, "logged out successfully", nil)
}

func handleAuthError(c *gin.Context, err error, detail error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrInvalidCredentials, apperrors.ErrUnauthorized:
		status = http.StatusUnauthorized
	case apperrors.ErrEmailAlreadyRegistered:
		status = http.StatusConflict
	case apperrors.ErrEmailRequired, apperrors.ErrPasswordRequired, apperrors.ErrInvalidInput:
		status = http.StatusBadRequest
	}

	errDetail := ""
	if detail != nil {
		errDetail = detail.Error()
	}

	response.Error(c, status, err.Error(), errDetail)
}
