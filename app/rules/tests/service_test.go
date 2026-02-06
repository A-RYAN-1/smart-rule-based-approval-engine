package tests

import (
	"context"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/rules"
	"github.com/ankita-advitot/rule_based_approval_engine/app/rules/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"

	"github.com/stretchr/testify/assert"
)

func TestRuleService_CreateRule(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		role          string
		rule          models.Rule
		mockSetup     func(r *mocks.RuleRepository)
		expectedError error
	}{
		{
			name: "Success",
			role: constants.RoleAdmin,
			rule: models.Rule{
				RequestType: "LEAVE",
				Action:      "APPROVE",
				GradeID:     1,
				Condition:   map[string]interface{}{"days": 3},
			},
			mockSetup: func(r *mocks.RuleRepository) {
				r.EXPECT().Create(ctx, &models.Rule{
					RequestType: "LEAVE",
					Action:      "APPROVE",
					GradeID:     1,
					Condition:   map[string]interface{}{"days": 3},
				}).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Unauthorized - Not Admin",
			role: constants.RoleEmployee,
			rule: models.Rule{
				RequestType: "LEAVE",
				Action:      "APPROVE",
				GradeID:     1,
				Condition:   map[string]interface{}{"days": 3},
			},
			mockSetup:     func(r *mocks.RuleRepository) {},
			expectedError: apperrors.ErrUnauthorized,
		},
		{
			name: "Missing RequestType",
			role: constants.RoleAdmin,
			rule: models.Rule{
				Action:    "APPROVE",
				GradeID:   1,
				Condition: map[string]interface{}{"days": 3},
			},
			mockSetup:     func(r *mocks.RuleRepository) {},
			expectedError: apperrors.ErrRequestTypeRequired,
		},
		{
			name: "Missing Action",
			role: constants.RoleAdmin,
			rule: models.Rule{
				RequestType: "LEAVE",
				GradeID:     1,
				Condition:   map[string]interface{}{"days": 3},
			},
			mockSetup:     func(r *mocks.RuleRepository) {},
			expectedError: apperrors.ErrActionRequired,
		},
		{
			name: "Missing GradeID",
			role: constants.RoleAdmin,
			rule: models.Rule{
				RequestType: "LEAVE",
				Action:      "APPROVE",
				Condition:   map[string]interface{}{"days": 3},
			},
			mockSetup:     func(r *mocks.RuleRepository) {},
			expectedError: apperrors.ErrGradeIDRequired,
		},
		{
			name: "Missing Condition",
			role: constants.RoleAdmin,
			rule: models.Rule{
				RequestType: "LEAVE",
				Action:      "APPROVE",
				GradeID:     1,
			},
			mockSetup:     func(r *mocks.RuleRepository) {},
			expectedError: apperrors.ErrConditionRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRuleRepository(t)
			tt.mockSetup(mockRepo)

			service := rules.NewRuleService(ctx, mockRepo)
			err := service.CreateRule(ctx, tt.role, tt.rule)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRuleService_GetRules(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		role          string
		mockSetup     func(r *mocks.RuleRepository)
		expectedError error
	}{
		{
			name: "Success",
			role: constants.RoleAdmin,
			mockSetup: func(r *mocks.RuleRepository) {
				r.EXPECT().GetAll(ctx).Return([]models.Rule{{ID: 1}}, nil)
			},
			expectedError: nil,
		},
		{
			name:          "Unauthorized",
			role:          constants.RoleEmployee,
			mockSetup:     func(r *mocks.RuleRepository) {},
			expectedError: apperrors.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRuleRepository(t)
			tt.mockSetup(mockRepo)

			service := rules.NewRuleService(ctx, mockRepo)
			_, err := service.GetRules(ctx, tt.role)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRuleService_UpdateRule(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := mocks.NewRuleRepository(t)
		rule := models.Rule{Action: "MANUAL"}
		mockRepo.EXPECT().Update(ctx, int64(1), &rule).Return(nil)

		service := rules.NewRuleService(ctx, mockRepo)
		err := service.UpdateRule(ctx, constants.RoleAdmin, 1, rule)

		assert.NoError(t, err)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		mockRepo := mocks.NewRuleRepository(t)
		service := rules.NewRuleService(ctx, mockRepo)
		err := service.UpdateRule(ctx, constants.RoleEmployee, 1, models.Rule{})

		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	})
}

func TestRuleService_DeleteRule(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := mocks.NewRuleRepository(t)
		mockRepo.EXPECT().Delete(ctx, int64(1)).Return(nil)

		service := rules.NewRuleService(ctx, mockRepo)
		err := service.DeleteRule(ctx, constants.RoleAdmin, 1)

		assert.NoError(t, err)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		mockRepo := mocks.NewRuleRepository(t)
		service := rules.NewRuleService(ctx, mockRepo)
		err := service.DeleteRule(ctx, constants.RoleEmployee, 1)

		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := mocks.NewRuleRepository(t)
		mockRepo.EXPECT().Delete(ctx, int64(99)).Return(apperrors.ErrDatabase)

		service := rules.NewRuleService(ctx, mockRepo)
		err := service.DeleteRule(ctx, constants.RoleAdmin, 99)

		assert.ErrorIs(t, err, apperrors.ErrDatabase)
	})
}
