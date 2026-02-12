package my_requests

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
)

type MyRequestsService struct {
	myRequestsRepo interfaces.MyRequestsRepository
}

func NewMyRequestsService(ctx context.Context, myRequestsRepo interfaces.MyRequestsRepository) interfaces.MyRequestsService {
	return &MyRequestsService{
		myRequestsRepo: myRequestsRepo,
	}
}

func (s *MyRequestsService) GetMyRequests(ctx context.Context, userID int64, reqType string, limit, offset int) ([]map[string]interface{}, int, error) {
	switch reqType {
	case "LEAVE":
		return s.myRequestsRepo.GetMyLeaveRequests(ctx, userID, limit, offset)
	case "EXPENSE":
		return s.myRequestsRepo.GetMyExpenseRequests(ctx, userID, limit, offset)
	case "DISCOUNT":
		return s.myRequestsRepo.GetMyDiscountRequests(ctx, userID, limit, offset)
	default:
		return nil, 0, nil
	}
}

func (s *MyRequestsService) GetMyAllRequests(ctx context.Context, userID int64, limit, offset int) (map[string]interface{}, error) {
	leaves, expenses, discounts, total, err := s.myRequestsRepo.GetMyAllRequests(ctx, userID, limit, offset)
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
func (s *MyRequestsService) GetPendingAllRequests(ctx context.Context, role string, userID int64, limit, offset int) (map[string]interface{}, error) {
	leaves, expenses, discounts, total, err := s.myRequestsRepo.GetPendingAllRequests(ctx, role, userID, limit, offset)
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
