package services

import (
	"context"
	"rule-based-approval-engine/internal/app/repositories"
)

type MyRequestsService struct {
	myRequestsRepo repositories.MyRequestsRepository
}

func NewMyRequestsService(myRequestsRepo repositories.MyRequestsRepository) *MyRequestsService {
	return &MyRequestsService{myRequestsRepo: myRequestsRepo}
}

func (s *MyRequestsService) GetMyLeaveRequests(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	return s.myRequestsRepo.GetMyLeaveRequests(ctx, userID)
}

func (s *MyRequestsService) GetMyExpenseRequests(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	return s.myRequestsRepo.GetMyExpenseRequests(ctx, userID)
}

func (s *MyRequestsService) GetMyDiscountRequests(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	return s.myRequestsRepo.GetMyDiscountRequests(ctx, userID)
}
