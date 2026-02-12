package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/my_requests"
	"github.com/ankita-advitot/rule_based_approval_engine/app/my_requests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMyRequestsHandler_GetMyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		reqType        string
		mockSetup      func(s *mocks.MyRequestsService)
		expectedStatus int
	}{
		{
			name:    "Success",
			userID:  1,
			reqType: "LEAVE",
			mockSetup: func(s *mocks.MyRequestsService) {
				s.EXPECT().GetMyRequests(mock.Anything, int64(1), "LEAVE").Return([]map[string]interface{}{{"id": 1}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Missing Type",
			userID:  1,
			reqType: "",
			mockSetup: func(s *mocks.MyRequestsService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Service Error",
			userID:  1,
			reqType: "EXPENSE",
			mockSetup: func(s *mocks.MyRequestsService) {
				s.EXPECT().GetMyRequests(mock.Anything, int64(1), "EXPENSE").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewMyRequestsService(t)
			tt.mockSetup(mockS)

			handler := my_requests.NewMyRequestsHandler(context.Background(), mockS)
			r := gin.New()
			r.GET("/my-requests", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				handler.GetMyRequests(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/my-requests?type="+tt.reqType, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestMyRequestsHandler_GetMyAllRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockS := mocks.NewMyRequestsService(t)
		mockS.EXPECT().GetMyAllRequests(mock.Anything, int64(1), 10, 0).Return(map[string]interface{}{"total": 1}, nil)

		handler := my_requests.NewMyRequestsHandler(context.Background(), mockS)
		r := gin.New()
		r.GET("/all", func(c *gin.Context) {
			c.Set("user_id", int64(1))
			handler.GetMyAllRequests(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/all", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
