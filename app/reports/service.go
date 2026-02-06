package reports

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
)

type ReportService struct {
	reportRepo interfaces.ReportRepository
}

func NewReportService(ctx context.Context, reportRepo interfaces.ReportRepository) interfaces.ReportService {
	return &ReportService{reportRepo: reportRepo}
}

func (s *ReportService) GetDashboardSummary(ctx context.Context, role string) (map[string]interface{}, error) {
	if role != constants.RoleAdmin {
		return nil, apperrors.ErrUnauthorized
	}

	dist, err := s.reportRepo.GetRequestStatusDistribution(ctx)
	if err != nil {
		return nil, err
	}

	types, err := s.reportRepo.GetRequestsByTypeReport(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"distribution": dist,
		"type_report":  types,
	}, nil
}
