package tests

import (
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestApplyCancelRules_MakeDecision(t *testing.T) {
	tests := []struct {
		name        string
		requestType string
		condition   map[string]interface{}
		value       float64
		expected    utils.DecisionResult
	}{
		{
			name:        "Leave Auto Approved",
			requestType: "LEAVE",
			condition:   map[string]interface{}{"max_days": 5.0},
			value:       3,
			expected: utils.DecisionResult{
				Status:  constants.StatusAutoApproved,
				Message: "LEAVE approved by system",
			},
		},
		{
			name:        "Leave Manual Approval",
			requestType: "LEAVE",
			condition:   map[string]interface{}{"max_days": 5.0},
			value:       6,
			expected: utils.DecisionResult{
				Status:  constants.StatusPending,
				Message: "LEAVE submitted for approval",
			},
		},
		{
			name:        "Expense Auto Approved",
			requestType: "EXPENSE",
			condition:   map[string]interface{}{"max_amount": 100.0},
			value:       50,
			expected: utils.DecisionResult{
				Status:  constants.StatusAutoApproved,
				Message: "EXPENSE approved by system",
			},
		},
		{
			name:        "Expense Manual Approval",
			requestType: "EXPENSE",
			condition:   map[string]interface{}{"max_amount": 100.0},
			value:       150,
			expected: utils.DecisionResult{
				Status:  constants.StatusPending,
				Message: "EXPENSE submitted for approval",
			},
		},
		{
			name:        "Discount Auto Approved",
			requestType: "DISCOUNT",
			condition:   map[string]interface{}{"max_percent": 20.0},
			value:       15,
			expected: utils.DecisionResult{
				Status:  constants.StatusAutoApproved,
				Message: "DISCOUNT approved by system",
			},
		},
		{
			name:        "Malformed Condition",
			requestType: "LEAVE",
			condition:   map[string]interface{}{"max_days": "invalid"},
			value:       3,
			expected: utils.DecisionResult{
				Status:  constants.StatusPending,
				Message: "LEAVE submitted for approval",
			},
		},
		{
			name:        "Unknown Request Type",
			requestType: "UNKNOWN",
			condition:   map[string]interface{}{"max_days": 5.0},
			value:       3,
			expected: utils.DecisionResult{
				Status:  constants.StatusPending,
				Message: "UNKNOWN submitted for approval",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.MakeDecision(tt.requestType, tt.condition, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyCancelRules_CanCancel(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		expectedError error
	}{
		{
			name:          "Can Cancel Pending",
			status:        constants.StatusPending,
			expectedError: nil,
		},
		{
			name:          "Cannot Cancel Approved",
			status:        constants.StatusApproved,
			expectedError: apperrors.ErrRequestCannotCancel,
		},
		{
			name:          "Cannot Cancel Rejected",
			status:        constants.StatusRejected,
			expectedError: apperrors.ErrRequestCannotCancel,
		},
		{
			name:          "Cannot Cancel Cancelled",
			status:        constants.StatusCancelled,
			expectedError: apperrors.ErrRequestCannotCancel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.CanCancel(tt.status)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplyCancelRules_EvaluateRules(t *testing.T) {
	t.Run("EvaluateLeaveRule", func(t *testing.T) {
		assert.True(t, utils.EvaluateLeaveRule(map[string]interface{}{"max_days": 5.0}, 3))
		assert.False(t, utils.EvaluateLeaveRule(map[string]interface{}{"max_days": 5.0}, 6))
		assert.False(t, utils.EvaluateLeaveRule(map[string]interface{}{"min_days": 5.0}, 3)) // Missing max_days
	})

	t.Run("EvaluateExpenseRule", func(t *testing.T) {
		assert.True(t, utils.EvaluateExpenseRule(map[string]interface{}{"max_amount": 100.0}, 50))
		assert.False(t, utils.EvaluateExpenseRule(map[string]interface{}{"max_amount": 100.0}, 150))
		assert.False(t, utils.EvaluateExpenseRule(map[string]interface{}{"min_amount": 100.0}, 50))
	})

	t.Run("EvaluateDiscountRule", func(t *testing.T) {
		assert.True(t, utils.EvaluateDiscountRule(map[string]interface{}{"max_percent": 20.0}, 10))
		assert.False(t, utils.EvaluateDiscountRule(map[string]interface{}{"max_percent": 20.0}, 25))
		assert.False(t, utils.EvaluateDiscountRule(map[string]interface{}{"min_percent": 20.0}, 10))
	})
}
