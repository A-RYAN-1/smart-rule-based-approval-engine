package rules

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
)

// RuleService handles business logic for rule management
type RuleService struct {
	ruleRepo  interfaces.RuleRepository
	gradeRepo interfaces.GradeRepository
	db        interfaces.DB
}

// NewRuleService creates a new instance of RuleService
func NewRuleService(ctx context.Context, ruleRepo interfaces.RuleRepository, gradeRepo interfaces.GradeRepository, db interfaces.DB) interfaces.RuleService {
	return &RuleService{
		ruleRepo:  ruleRepo,
		gradeRepo: gradeRepo,
		db:        db,
	}
}

// GetRule retrieves a rule by request type and grade ID
func (s *RuleService) GetRule(ctx context.Context, requestType string, gradeID int64) (*models.Rule, error) {
	return s.ruleRepo.GetByTypeAndGrade(ctx, requestType, gradeID)
}

// CreateRule creates or updates a rule (admin only)
func (s *RuleService) CreateRule(ctx context.Context, role string, rule models.Rule) error {
	if role != constants.RoleAdmin {
		return apperrors.ErrUnauthorized
	}

	if err := s.validateRule(ctx, rule); err != nil {
		return err
	}

	err := s.ruleRepo.Create(ctx, &rule)
	if err != nil {
		return apperrors.ErrDatabase
	}

	return nil
}

// GetRules retrieves all rules (admin only)
func (s *RuleService) GetRules(ctx context.Context, role string) ([]models.Rule, error) {
	if role != constants.RoleAdmin {
		return nil, apperrors.ErrUnauthorized
	}

	return s.ruleRepo.GetAll(ctx)
}

// updates an existing rule (admin only)
func (s *RuleService) UpdateRule(ctx context.Context, role string, ruleID int64, rule models.Rule) error {
	if role != constants.RoleAdmin {
		return apperrors.ErrUnauthorized
	}

	if err := s.validateRule(ctx, rule); err != nil {
		return err
	}

	return s.ruleRepo.Update(ctx, ruleID, &rule)
}

// deletes a rule by ID (admin only)
func (s *RuleService) DeleteRule(ctx context.Context, role string, ruleID int64) error {
	if role != constants.RoleAdmin {
		return apperrors.ErrUnauthorized
	}

	return s.ruleRepo.Delete(ctx, ruleID)
}

func (s *RuleService) validateRule(ctx context.Context, rule models.Rule) error {
	if rule.RequestType == "" {
		return apperrors.ErrRequestTypeRequired
	}

	if rule.Action != constants.StatusAutoApprove {
		return apperrors.ErrActionRequired
	}

	if rule.GradeID == 0 {
		return apperrors.ErrGradeIDRequired
	}

	if rule.Condition == nil || len(rule.Condition) == 0 {
		return apperrors.ErrConditionRequired
	}

	// Fetch grade limits for validation
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperrors.ErrTransactionBegin
	}
	defer tx.Rollback(ctx)

	leaveLimit, expenseLimit, discountLimit, err := s.gradeRepo.GetLimits(ctx, tx, rule.GradeID)
	if err != nil {
		return err
	}

	// Validate Condition values
	for key, val := range rule.Condition {
		numVal, ok := val.(float64)
		if !ok {
			continue // Skip non-numeric if any
		}

		if numVal < 0 {
			return apperrors.ErrNegativeValue
		}

		switch rule.RequestType {
		case "LEAVE":
			if key == "max_days" && numVal > float64(leaveLimit) {
				return apperrors.ErrQuotaExceeded
			}
		case "EXPENSE":
			if key == "max_amount" && numVal > expenseLimit {
				return apperrors.ErrQuotaExceeded
			}
		case "DISCOUNT":
			if key == "max_percent" && numVal > discountLimit {
				return apperrors.ErrQuotaExceeded
			}
		}
	}

	return nil
}
