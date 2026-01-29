package services

import (
	"context"
	"rule-based-approval-engine/internal/app/repositories"
	"rule-based-approval-engine/internal/models"
)

type ReportService struct {
	reportRepo repositories.ReportRepository
}

func NewReportService(reportRepo repositories.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

func (s *ReportService) GetRequestStatusDistribution(ctx context.Context) (map[string]int, error) {
	return s.reportRepo.GetRequestStatusDistribution(ctx)
}

func (s *ReportService) GetRequestsByTypeReport(ctx context.Context) ([]models.RequestTypeReport, error) {
	return s.reportRepo.GetRequestsByTypeReport(ctx)
}
