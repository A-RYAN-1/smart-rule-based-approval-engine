package tests

import (
	"context"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/reports"
	"github.com/ankita-advitot/rule_based_approval_engine/app/reports/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/stretchr/testify/assert"
)

func TestReportServiceGetDashboardSummary(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		role          string
		mockSetup     func(r *mocks.ReportRepository)
		expectedError error
	}{
		{
			name: "Success - Admin",
			role: "ADMIN",
			mockSetup: func(r *mocks.ReportRepository) {
				r.EXPECT().GetRequestStatusDistribution(ctx).Return(map[string]int{"APPROVED": 5}, nil)
				r.EXPECT().GetRequestsByTypeReport(ctx).Return([]models.RequestTypeReport{
					{Type: "LEAVE", TotalRequests: 10},
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:          "Unauthorized - Non Admin",
			role:          "EMPLOYEE",
			mockSetup:     func(r *mocks.ReportRepository) {},
			expectedError: apperrors.ErrUnauthorized,
		},
		{
			name: "Repo Error - Distribution",
			role: "ADMIN",
			mockSetup: func(r *mocks.ReportRepository) {
				r.EXPECT().GetRequestStatusDistribution(ctx).Return(nil, apperrors.ErrDatabase)
			},
			expectedError: apperrors.ErrDatabase,
		},
		{
			name: "Repo Error - Type Report",
			role: "ADMIN",
			mockSetup: func(r *mocks.ReportRepository) {
				r.EXPECT().GetRequestStatusDistribution(ctx).Return(map[string]int{"APPROVED": 5}, nil)
				r.EXPECT().GetRequestsByTypeReport(ctx).Return(nil, apperrors.ErrDatabase)
			},
			expectedError: apperrors.ErrDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewReportRepository(t)
			tt.mockSetup(mockRepo)

			service := reports.NewReportService(ctx, mockRepo)
			result, err := service.GetDashboardSummary(ctx, tt.role)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
