package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/reports"
	"github.com/ankita-advitot/rule_based_approval_engine/app/reports/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReportHandler_GetDashboardSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		mockSetup      func(s *mocks.ReportService)
		expectedStatus int
	}{
		{
			name: "Success",
			role: "ADMIN",
			mockSetup: func(s *mocks.ReportService) {
				s.EXPECT().GetDashboardSummary(mock.Anything, "ADMIN").Return(map[string]interface{}{"data": "ok"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Unauthorized",
			role: "EMPLOYEE",
			mockSetup: func(s *mocks.ReportService) {
				s.EXPECT().GetDashboardSummary(mock.Anything, "EMPLOYEE").Return(nil, apperrors.ErrUnauthorized)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewReportService(t)
			tt.mockSetup(mockS)

			handler := reports.NewReportHandler(context.Background(), mockS)
			r := gin.New()
			r.GET("/summary", func(c *gin.Context) {
				c.Set("role", tt.role)
				handler.GetDashboardSummary(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/summary", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
