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

func (s *MyRequestsService) GetMyRequests(ctx context.Context, userID int64, reqType string) ([]map[string]interface{}, error) {
	switch reqType {
	case "LEAVE":
		return s.myRequestsRepo.GetMyLeaveRequests(ctx, userID)
	case "EXPENSE":
		return s.myRequestsRepo.GetMyExpenseRequests(ctx, userID)
	case "DISCOUNT":
		return s.myRequestsRepo.GetMyDiscountRequests(ctx, userID)
	default:
		return nil, nil
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
