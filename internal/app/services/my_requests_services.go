package services

import (
	"context"
	"rule-based-approval-engine/internal/app/repositories"
)

type MyRequestsServices interface {
	GetMyAllRequests(ctx context.Context, userID int64, limit, offset int) (map[string]interface{}, error)
}

type myRequestsServices struct {
	repo repositories.AggregatedRepository
}

func NewMyRequestsServices(repo repositories.AggregatedRepository) MyRequestsServices {
	return &myRequestsServices{repo: repo}
}

func (s *myRequestsServices) GetMyAllRequests(ctx context.Context, userID int64, limit, offset int) (map[string]interface{}, error) {
	leaves, expenses, discounts, total, err := s.repo.GetMyAllRequests(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"leave_request":    leaves,
		"expense_request":  expenses,
		"discount_request": discounts,
		"total":            total,
	}

	return result, nil
}
