// test cases for expense approval service
package tests

import (
	"context"
	"testing"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/stretchr/testify/assert"
)

func TestExpenseApprovalService_ApproveExpense_Validation(t *testing.T) {
	service := services.NewExpenseApprovalService(nil, nil, nil, nil)
	ctx := context.Background()

	t.Run("Employee Cannot Approve", func(t *testing.T) {
		err := service.ApproveExpense(ctx, "EMPLOYEE", 1, 1, "Comment")
		assert.Equal(t, apperrors.ErrEmployeeCannotApprove, err)
	})

	t.Run("Comment Required", func(t *testing.T) {
		err := service.ApproveExpense(ctx, "ADMIN", 1, 1, "")
		assert.Equal(t, apperrors.ErrCommentRequired, err)
	})
}
