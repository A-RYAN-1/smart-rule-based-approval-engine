package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/leave_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/leave_service/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLeaveHandler_CountDays(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		mockSetup      func(s *mocks.LeaveService)
		expectedStatus int
		expectedDays   float64
	}{
		{
			name:  "returns the service count",
			query: "?from=2031-03-07&to=2031-03-10",
			mockSetup: func(s *mocks.LeaveService) {
				s.EXPECT().CountLeaveDays(mock.Anything, day("2031-03-07"), day("2031-03-10")).Return(2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedDays:   2,
		},
		{
			name:           "missing dates are a bad request",
			query:          "",
			mockSetup:      func(s *mocks.LeaveService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "a malformed date is a bad request",
			query:          "?from=07-03-2031&to=2031-03-10",
			mockSetup:      func(s *mocks.LeaveService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "from after to is a bad request",
			query: "?from=2031-03-10&to=2031-03-07",
			mockSetup: func(s *mocks.LeaveService) {
				s.EXPECT().CountLeaveDays(mock.Anything, day("2031-03-10"), day("2031-03-07")).Return(0, apperrors.ErrInvalidDateRange)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := mocks.NewLeaveService(t)
			tt.mockSetup(service)
			handler := leave_service.NewLeaveHandler(context.Background(), service)

			router := gin.New()
			router.GET("/leaves/working-days", handler.CountDays)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/leaves/working-days"+tt.query, nil))

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusOK {
				var body struct {
					Data struct {
						Days float64 `json:"days"`
					} `json:"data"`
				}
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tt.expectedDays, body.Data.Days)
			}
		})
	}
}
