package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/app/leave_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/leave_service/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLeaveHandler_ApplyLeave(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		reqBody        interface{}
		mockSetup      func(s *mocks.LeaveService)
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			reqBody: leave_service.LeaveApplyRequest{
				FromDate:  "2023-10-01",
				ToDate:    "2023-10-02",
				LeaveType: "SICK",
				Reason:    "Flu",
			},
			mockSetup: func(s *mocks.LeaveService) {
				from, _ := time.Parse("2006-01-02", "2023-10-01")
				to, _ := time.Parse("2006-01-02", "2023-10-02")
				s.EXPECT().ApplyLeave(mock.Anything, int64(1), from, to, 2, "SICK", "Flu").Return("Approved", "APPROVED", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Invalid Date Format",
			userID: 1,
			reqBody: leave_service.LeaveApplyRequest{
				FromDate: "01-10-2023",
				ToDate:   "2023-10-02",
			},
			mockSetup:      func(s *mocks.LeaveService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			userID:         1,
			reqBody:        "{ invalid",
			mockSetup:      func(s *mocks.LeaveService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Service Error - Overlap",
			userID: 1,
			reqBody: leave_service.LeaveApplyRequest{
				FromDate:  "2023-10-01",
				ToDate:    "2023-10-02",
				LeaveType: "SICK",
			},
			mockSetup: func(s *mocks.LeaveService) {
				s.EXPECT().ApplyLeave(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("", "", apperrors.ErrLeaveOverlap)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized - No UserID",
			userID:         0,
			reqBody:        leave_service.LeaveApplyRequest{},
			mockSetup:      func(s *mocks.LeaveService) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewLeaveService(t)
			tt.mockSetup(mockService)

			handler := leave_service.NewLeaveHandler(nil, mockService)
			r := gin.New()
			r.POST("/apply", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				handler.ApplyLeave(c)
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

func TestLeaveHandler_CancelLeave(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		reqID          string
		mockSetup      func(s *mocks.LeaveService)
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			reqID:  "123",
			mockSetup: func(s *mocks.LeaveService) {
				s.EXPECT().CancelLeave(mock.Anything, int64(1), int64(123)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Request Not Found",
			userID: 1,
			reqID:  "456",
			mockSetup: func(s *mocks.LeaveService) {
				s.EXPECT().CancelLeave(mock.Anything, int64(1), int64(456)).Return(apperrors.ErrLeaveRequestNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid ID",
			userID:         1,
			reqID:          "abc",
			mockSetup:      func(s *mocks.LeaveService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewLeaveService(t)
			tt.mockSetup(mockService)

			handler := leave_service.NewLeaveHandler(nil, mockService)
			r := gin.New()
			r.POST("/cancel/:id", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				handler.CancelLeave(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/cancel/"+tt.reqID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestLeaveApprovalHandler_ApproveLeave(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		approverID     int64
		reqID          string
		reqBody        interface{}
		mockSetup      func(s *mocks.LeaveApprovalService)
		expectedStatus int
	}{
		{
			name:       "Success",
			role:       "MANAGER",
			approverID: 2,
			reqID:      "10",
			reqBody:    map[string]interface{}{"comment": "OK"},
			mockSetup: func(s *mocks.LeaveApprovalService) {
				s.EXPECT().ApproveLeave(mock.Anything, "MANAGER", int64(2), int64(10), "OK").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			role:           "MANAGER",
			approverID:     2,
			reqID:          "invalid",
			reqBody:        map[string]interface{}{},
			mockSetup:      func(s *mocks.LeaveApprovalService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewLeaveApprovalService(t)
			tt.mockSetup(mockService)

			handler := leave_service.NewLeaveApprovalHandler(nil, mockService)
			r := gin.New()
			r.POST("/approve/:id", func(c *gin.Context) {
				c.Set("user_id", tt.approverID)
				c.Set("role", tt.role)
				handler.ApproveLeave(c)
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

func TestLeaveApprovalHandler_RejectLeave(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		approverID     int64
		reqID          string
		reqBody        interface{}
		mockSetup      func(s *mocks.LeaveApprovalService)
		expectedStatus int
	}{
		{
			name:       "Success",
			role:       "MANAGER",
			approverID: 2,
			reqID:      "10",
			reqBody:    map[string]interface{}{"comment": "Too much work"},
			mockSetup: func(s *mocks.LeaveApprovalService) {
				s.EXPECT().RejectLeave(mock.Anything, "MANAGER", int64(2), int64(10), "Too much work").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing Comment",
			role:           "MANAGER",
			approverID:     2,
			reqID:          "10",
			reqBody:        map[string]interface{}{},
			mockSetup:      func(s *mocks.LeaveApprovalService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewLeaveApprovalService(t)
			tt.mockSetup(mockService)

			handler := leave_service.NewLeaveApprovalHandler(nil, mockService)
			r := gin.New()
			r.POST("/reject/:id", func(c *gin.Context) {
				c.Set("user_id", tt.approverID)
				c.Set("role", tt.role)
				handler.RejectLeave(c)
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

func TestLeaveApprovalHandler_GetPendingLeaves(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		userID         int64
		mockSetup      func(s *mocks.LeaveApprovalService)
		expectedStatus int
	}{
		{
			name:   "Success - Manager",
			role:   "MANAGER",
			userID: 2,
			mockSetup: func(s *mocks.LeaveApprovalService) {
				s.EXPECT().GetPendingLeaveRequests(mock.Anything, "MANAGER", int64(2)).Return([]map[string]interface{}{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Success - Admin",
			role:   "ADMIN",
			userID: 1,
			mockSetup: func(s *mocks.LeaveApprovalService) {
				s.EXPECT().GetPendingLeaveRequests(mock.Anything, "ADMIN", int64(1)).Return([]map[string]interface{}{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewLeaveApprovalService(t)
			tt.mockSetup(mockService)

			handler := leave_service.NewLeaveApprovalHandler(nil, mockService)
			r := gin.New()
			r.GET("/pending", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Set("role", tt.role)
				handler.GetPendingLeaves(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/pending", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
