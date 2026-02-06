package tests

import (
	"context"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/my_requests"
	"github.com/ankita-advitot/rule_based_approval_engine/app/my_requests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestMyRequestsService(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	t.Run("GetMyRequests - EXPENSE", func(t *testing.T) {
		mockRepo := mocks.NewMyRequestsRepository(t)
		mockRepo.EXPECT().GetMyExpenseRequests(ctx, userID).Return([]map[string]interface{}{{"id": 2}}, nil)

		service := my_requests.NewMyRequestsService(ctx, mockRepo)
		result, err := service.GetMyRequests(ctx, userID, "EXPENSE")

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 2, result[0]["id"])
	})

	t.Run("GetMyRequests - DISCOUNT", func(t *testing.T) {
		mockRepo := mocks.NewMyRequestsRepository(t)
		mockRepo.EXPECT().GetMyDiscountRequests(ctx, userID).Return([]map[string]interface{}{{"id": 3}}, nil)

		service := my_requests.NewMyRequestsService(ctx, mockRepo)
		result, err := service.GetMyRequests(ctx, userID, "DISCOUNT")

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 3, result[0]["id"])
	})

	t.Run("GetMyRequests - Repository Error", func(t *testing.T) {
		mockRepo := mocks.NewMyRequestsRepository(t)
		mockRepo.EXPECT().GetMyLeaveRequests(ctx, userID).Return(nil, assert.AnError)

		service := my_requests.NewMyRequestsService(ctx, mockRepo)
		_, err := service.GetMyRequests(ctx, userID, "LEAVE")

		assert.Error(t, err)
	})

	t.Run("GetMyAllRequests - Repository Error", func(t *testing.T) {
		mockRepo := mocks.NewMyRequestsRepository(t)
		mockRepo.EXPECT().GetMyAllRequests(ctx, userID, 10, 0).Return(nil, nil, nil, 0, assert.AnError)

		service := my_requests.NewMyRequestsService(ctx, mockRepo)
		_, err := service.GetMyAllRequests(ctx, userID, 10, 0)

		assert.Error(t, err)
	})
}
