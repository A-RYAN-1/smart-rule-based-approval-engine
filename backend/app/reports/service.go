package reports

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
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

	// Get pending counts by type
	pendingLeave, err := s.reportRepo.GetPendingLeaveCount(ctx)
	if err != nil {
		return nil, err
	}

	pendingExpense, err := s.reportRepo.GetPendingExpenseCount(ctx)
	if err != nil {
		return nil, err
	}

	pendingDiscount, err := s.reportRepo.GetPendingDiscountCount(ctx)
	if err != nil {
		return nil, err
	}

	// Build pending distribution by type
	pendingByType := map[string]interface{}{
		"leave":    pendingLeave,
		"expense":  pendingExpense,
		"discount": pendingDiscount,
	}

	return map[string]interface{}{
		"total_pending": dist["pending"],
		"approved":      dist["approved"],
		"rejected":      dist["rejected"],
		"auto_rejected": dist["auto_rejected"],
		"pending":       dist["pending"],
		"cancelled":     dist["cancelled"],
		"distribution": map[string]interface{}{
			"pending":       pendingByType,
			"approved":      dist["approved"],
			"rejected":      dist["rejected"],
			"auto_rejected": dist["auto_rejected"],
			"cancelled":     dist["cancelled"],
		},
		"type_report": types,
	}, nil
}

func (s *ReportService) GetRequestStatusDistribution(ctx context.Context) (map[string]int, error) {
	return s.reportRepo.GetRequestStatusDistribution(ctx)
}

func (s *ReportService) GetRequestsByTypeReport(ctx context.Context) ([]models.RequestTypeReport, error) {
	return s.reportRepo.GetRequestsByTypeReport(ctx)
}
