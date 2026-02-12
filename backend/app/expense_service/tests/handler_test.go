package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/expense_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/expense_service/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExpenseHandler_ApplyExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		reqBody        interface{}
		mockSetup      func(s *mocks.ExpenseService)
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			reqBody: expense_service.ExpenseApplyRequest{
				Amount:   100.0,
				Category: "Travel",
				Reason:   "Conference",
			},
			mockSetup: func(s *mocks.ExpenseService) {
				s.EXPECT().ApplyExpense(mock.Anything, int64(1), 100.0, "Travel", "Conference").Return("Expense submitted", "PENDING", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid Amount",
			userID: 1,
			reqBody: expense_service.ExpenseApplyRequest{
				Amount:   -50.0,
				Category: "Travel",
			},
			mockSetup: func(s *mocks.ExpenseService) {
				s.EXPECT().ApplyExpense(mock.Anything, int64(1), -50.0, "Travel", "").Return("", "", apperrors.ErrInvalidExpenseAmount)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			userID:         1,
			reqBody:        "{ invalid",
			mockSetup:      func(s *mocks.ExpenseService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized - No UserID",
			userID:         0,
			reqBody:        expense_service.ExpenseApplyRequest{},
			mockSetup:      func(s *mocks.ExpenseService) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewExpenseService(t)
			tt.mockSetup(mockService)

			handler := expense_service.NewExpenseHandler(nil, mockService)
			r := gin.New()
			r.POST("/apply", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				handler.ApplyExpense(c)
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

func TestExpenseHandler_CancelExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		reqID          string
		mockSetup      func(s *mocks.ExpenseService)
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			reqID:  "123",
			mockSetup: func(s *mocks.ExpenseService) {
				s.EXPECT().CancelExpense(mock.Anything, int64(1), int64(123)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Request Not Found",
			userID: 1,
			reqID:  "456",
			mockSetup: func(s *mocks.ExpenseService) {
				s.EXPECT().CancelExpense(mock.Anything, int64(1), int64(456)).Return(apperrors.ErrExpenseRequestNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid ID",
			userID:         1,
			reqID:          "abc",
			mockSetup:      func(s *mocks.ExpenseService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewExpenseService(t)
			tt.mockSetup(mockService)

			handler := expense_service.NewExpenseHandler(nil, mockService)
			r := gin.New()
			r.POST("/cancel/:id", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				handler.CancelExpense(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/cancel/"+tt.reqID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestExpenseApprovalHandler_ApproveExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		approverID     int64
		reqID          string
		reqBody        interface{}
		mockSetup      func(s *mocks.ExpenseApprovalService)
		expectedStatus int
	}{
		{
			name:       "Success",
			role:       "ADMIN",
			approverID: 1,
			reqID:      "50",
			reqBody:    map[string]interface{}{"comment": "Approved by admin"},
			mockSetup: func(s *mocks.ExpenseApprovalService) {
				s.EXPECT().ApproveExpense(mock.Anything, "ADMIN", int64(1), int64(50), "Approved by admin").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			role:           "ADMIN",
			approverID:     1,
			reqID:          "bad-id",
			reqBody:        map[string]interface{}{},
			mockSetup:      func(s *mocks.ExpenseApprovalService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewExpenseApprovalService(t)
			tt.mockSetup(mockService)

			handler := expense_service.NewExpenseApprovalHandler(nil, mockService)
			r := gin.New()
			r.POST("/approve/:id", func(c *gin.Context) {
				c.Set("user_id", tt.approverID)
				c.Set("role", tt.role)
				handler.ApproveExpense(c)
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

func TestExpenseApprovalHandler_RejectExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		approverID     int64
		reqID          string
		reqBody        interface{}
		mockSetup      func(s *mocks.ExpenseApprovalService)
		expectedStatus int
	}{
		{
			name:       "Success",
			role:       "MANAGER",
			approverID: 2,
			reqID:      "10",
			reqBody:    map[string]interface{}{"comment": "Too expensive"},
			mockSetup: func(s *mocks.ExpenseApprovalService) {
				s.EXPECT().RejectExpense(mock.Anything, "MANAGER", int64(2), int64(10), "Too expensive").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing Comment",
			role:           "MANAGER",
			approverID:     2,
			reqID:          "10",
			reqBody:        map[string]interface{}{},
			mockSetup:      func(s *mocks.ExpenseApprovalService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewExpenseApprovalService(t)
			tt.mockSetup(mockService)

			handler := expense_service.NewExpenseApprovalHandler(nil, mockService)
			r := gin.New()
			r.POST("/reject/:id", func(c *gin.Context) {
				c.Set("user_id", tt.approverID)
				c.Set("role", tt.role)
				handler.RejectExpense(c)
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

func TestExpenseApprovalHandler_GetPendingExpenses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		userID         int64
		mockSetup      func(s *mocks.ExpenseApprovalService)
		expectedStatus int
	}{
		{
			name:   "Success - Manager",
			role:   "MANAGER",
			userID: 2,
			mockSetup: func(s *mocks.ExpenseApprovalService) {
				s.EXPECT().GetPendingExpenseRequests(mock.Anything, "MANAGER", int64(2)).Return([]map[string]interface{}{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewExpenseApprovalService(t)
			tt.mockSetup(mockService)

			handler := expense_service.NewExpenseApprovalHandler(nil, mockService)
			r := gin.New()
			r.GET("/pending", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Set("role", tt.role)
				handler.GetPendingExpenses(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/pending", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
