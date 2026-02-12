package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/domain_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/domain_service/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDiscountHandler_ApplyDiscount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		reqBody        interface{}
		mockSetup      func(s *mocks.DiscountService)
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			reqBody: domain_service.DiscountApplyRequest{
				DiscountPercentage: 5.0,
				Reason:             "Reward",
			},
			mockSetup: func(s *mocks.DiscountService) {
				s.EXPECT().ApplyDiscount(mock.Anything, int64(1), 5.0, "Reward").Return("Approved", "AUTO_APPROVED", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid Percentage",
			userID: 1,
			reqBody: domain_service.DiscountApplyRequest{
				DiscountPercentage: -5.0,
			},
			mockSetup: func(s *mocks.DiscountService) {
				s.EXPECT().ApplyDiscount(mock.Anything, int64(1), -5.0, "").Return("", "", apperrors.ErrInvalidDiscountPercent)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized",
			userID:         0,
			reqBody:        domain_service.DiscountApplyRequest{},
			mockSetup:      func(s *mocks.DiscountService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid JSON",
			userID:         1,
			reqBody:        "{ invalid",
			mockSetup:      func(s *mocks.DiscountService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewDiscountService(t)
			tt.mockSetup(mockS)

			handler := domain_service.NewDiscountHandler(nil, mockS)
			r := gin.New()
			r.POST("/apply", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				handler.ApplyDiscount(c)
			})

			var body []byte
			if s, ok := tt.reqBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}
			req := httptest.NewRequest(http.MethodPost, "/apply", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDiscountHandler_CancelDiscount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		reqID          string
		mockSetup      func(s *mocks.DiscountService)
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			reqID:  "123",
			mockSetup: func(s *mocks.DiscountService) {
				s.EXPECT().CancelDiscount(mock.Anything, int64(1), int64(123)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			userID:         1,
			reqID:          "abc",
			mockSetup:      func(s *mocks.DiscountService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewDiscountService(t)
			tt.mockSetup(mockS)

			handler := domain_service.NewDiscountHandler(nil, mockS)
			r := gin.New()
			r.POST("/cancel/:id", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				handler.CancelDiscount(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/cancel/"+tt.reqID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDiscountApprovalHandler_ApproveDiscount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		approverID     int64
		reqID          string
		reqBody        interface{}
		mockSetup      func(s *mocks.DiscountApprovalService)
		expectedStatus int
	}{
		{
			name:       "Success",
			role:       "ADMIN",
			approverID: 3,
			reqID:      "10",
			reqBody:    map[string]interface{}{"comment": "OK"},
			mockSetup: func(s *mocks.DiscountApprovalService) {
				s.EXPECT().ApproveDiscount(mock.Anything, "ADMIN", int64(3), int64(10), "OK").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewDiscountApprovalService(t)
			tt.mockSetup(mockS)

			handler := domain_service.NewDiscountApprovalHandler(nil, mockS)
			r := gin.New()
			r.POST("/approve/:id", func(c *gin.Context) {
				c.Set("user_id", tt.approverID)
				c.Set("role", tt.role)
				handler.ApproveDiscount(c)
			})

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/approve/"+tt.reqID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDiscountApprovalHandler_RejectDiscount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		approverID     int64
		reqID          string
		reqBody        interface{}
		mockSetup      func(s *mocks.DiscountApprovalService)
		expectedStatus int
	}{
		{
			name:       "Success",
			role:       "MANAGER",
			approverID: 2,
			reqID:      "10",
			reqBody:    map[string]interface{}{"comment": "No"},
			mockSetup: func(s *mocks.DiscountApprovalService) {
				s.EXPECT().RejectDiscount(mock.Anything, "MANAGER", int64(2), int64(10), "No").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing Comment",
			role:           "MANAGER",
			approverID:     2,
			reqID:          "10",
			reqBody:        map[string]interface{}{},
			mockSetup:      func(s *mocks.DiscountApprovalService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewDiscountApprovalService(t)
			tt.mockSetup(mockS)

			handler := domain_service.NewDiscountApprovalHandler(nil, mockS)
			r := gin.New()
			r.POST("/reject/:id", func(c *gin.Context) {
				c.Set("user_id", tt.approverID)
				c.Set("role", tt.role)
				handler.RejectDiscount(c)
			})

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/reject/"+tt.reqID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestBalanceHandler_GetBalances(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		mockSetup      func(s *mocks.BalanceService)
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			mockSetup: func(s *mocks.BalanceService) {
				s.EXPECT().GetBalances(mock.Anything, int64(1)).Return(map[string]interface{}{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Service Error",
			userID: 1,
			mockSetup: func(s *mocks.BalanceService) {
				s.EXPECT().GetBalances(mock.Anything, int64(1)).Return(nil, apperrors.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewBalanceService(t)
			tt.mockSetup(mockS)

			handler := domain_service.NewBalanceHandler(nil, mockS)
			r := gin.New()
			r.GET("/balances", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				handler.GetBalances(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/balances", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
