package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/auth"
	"github.com/ankita-advitot/rule_based_approval_engine/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		reqBody        interface{}
		mockSetup      func(s *mocks.AuthService)
		expectedStatus int
	}{
		{
			name: "Success",
			reqBody: auth.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(s *mocks.AuthService) {
				s.EXPECT().RegisterUser(mock.Anything, "John Doe", "john@example.com", "password123").Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			reqBody:        "{ invalid json",
			mockSetup:      func(s *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error - Conflict",
			reqBody: auth.RegisterRequest{
				Name:     "John Doe",
				Email:    "exists@example.com",
				Password: "password123",
			},
			mockSetup: func(s *mocks.AuthService) {
				s.EXPECT().RegisterUser(mock.Anything, "John Doe", "exists@example.com", "password123").Return(apperrors.ErrEmailAlreadyRegistered)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Missing Name",
			reqBody: auth.RegisterRequest{
				Name:     "",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup:      func(s *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Email",
			reqBody: auth.RegisterRequest{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			mockSetup:      func(s *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Password",
			reqBody: auth.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "",
			},
			mockSetup:      func(s *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewAuthService(t)
			tt.mockSetup(mockService)

			handler := auth.NewAuthHandler(nil, mockService)
			r := gin.Default()
			r.POST("/register", handler.Register)

			var body []byte
			if s, ok := tt.reqBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		reqBody        interface{}
		mockSetup      func(s *mocks.AuthService)
		expectedStatus int
	}{
		{
			name: "Success",
			reqBody: auth.LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(s *mocks.AuthService) {
				s.EXPECT().LoginUser(mock.Anything, "john@example.com", "password123").Return("mock-token", "ADMIN", nil)
				s.EXPECT().GetUserByEmail(mock.Anything, "john@example.com").Return(&models.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
					Role:  "ADMIN",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Credentials",
			reqBody: auth.LoginRequest{
				Email:    "john@example.com",
				Password: "wrong",
			},
			mockSetup: func(s *mocks.AuthService) {
				s.EXPECT().LoginUser(mock.Anything, "john@example.com", "wrong").Return("", "", apperrors.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Missing Email",
			reqBody: auth.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			mockSetup:      func(s *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Email Format",
			reqBody: auth.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			mockSetup:      func(s *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Password",
			reqBody: auth.LoginRequest{
				Email:    "john@example.com",
				Password: "",
			},
			mockSetup:      func(s *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewAuthService(t)
			tt.mockSetup(mockService)

			handler := auth.NewAuthHandler(nil, mockService)
			r := gin.Default()
			r.POST("/login", handler.Login)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := auth.NewAuthHandler(nil, nil)
	r := gin.Default()
	r.POST("/logout", handler.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
