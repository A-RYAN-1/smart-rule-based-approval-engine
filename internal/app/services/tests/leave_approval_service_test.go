// File: internal/app/services/tests/leave_approval_service_test.go
package tests

import (
	"context"
	"testing"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/stretchr/testify/assert"
)

func TestLeaveApprovalService_ApproveLeave_Validation(t *testing.T) {
	service := services.NewLeaveApprovalService(nil, nil, nil, nil)
	ctx := context.Background()

	t.Run("Employee Cannot Approve", func(t *testing.T) {
		err := service.ApproveLeave(ctx, "EMPLOYEE", 1, 1, "Comment")
		assert.Equal(t, apperrors.ErrEmployeeCannotApprove, err)
	})

	t.Run("Comment Required", func(t *testing.T) {
		err := service.ApproveLeave(ctx, "ADMIN", 1, 1, "")
		assert.Equal(t, apperrors.ErrCommentRequired, err)
	})
}

func TestLeaveApprovalService_GetPending_Unauthorized(t *testing.T) {
	service := services.NewLeaveApprovalService(nil, nil, nil, nil)
	ctx := context.Background()

	_, err := service.GetPendingLeaveRequests(ctx, "EMPLOYEE", 1)
	assert.Equal(t, apperrors.ErrUnauthorizedRole, err)
}
