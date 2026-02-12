package expense_service

import (
	"context"
	"strings"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

// handles business logic for expense requests
type ExpenseService struct {
	expenseReqRepo interfaces.ExpenseRequestRepository
	balanceRepo    interfaces.BalanceRepository
	ruleService    interfaces.RuleService
	userRepo       interfaces.UserRepository
	db             interfaces.DB
}

// creates a new instance of ExpenseService
func NewExpenseService(
	ctx context.Context,
	expenseReqRepo interfaces.ExpenseRequestRepository,
	balanceRepo interfaces.BalanceRepository,
	ruleService interfaces.RuleService,
	userRepo interfaces.UserRepository,
	db interfaces.DB,
) interfaces.ExpenseService {
	return &ExpenseService{
		expenseReqRepo: expenseReqRepo,
		balanceRepo:    balanceRepo,
		ruleService:    ruleService,
		userRepo:       userRepo,
		db:             db,
	}
}

// processes an expense application
func (s *ExpenseService) ApplyExpense(
	ctx context.Context,
	userID int64,
	amount float64,
	category string,
	reason string,
) (string, string, error) {
	// validations
	if userID <= 0 {
		return "", "", apperrors.ErrInvalidUser
	}

	if amount <= 0 {
		return "", "", apperrors.ErrInvalidExpenseAmount
	}

	if strings.TrimSpace(category) == "" {
		return "", "", apperrors.ErrInvalidExpenseCategory
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", apperrors.ErrTransactionBegin
	}
	defer tx.Rollback(ctx)

	// expense balance
	remaining, err := s.balanceRepo.GetExpenseBalance(ctx, tx, userID)
	if err != nil {
		return "", "", err
	}

	if amount > remaining {
		return "", "", apperrors.ErrExpenseLimitExceeded
	}

	// user grade
	gradeID, err := s.userRepo.GetGrade(ctx, tx, userID)
	if err != nil {
		return "", "", err
	}

	// fetch rule
	rule, err := s.ruleService.GetRule(ctx, "EXPENSE", gradeID)
	if err != nil {
		return "", "", apperrors.ErrRuleNotFound
	}

	// apply rule
	result := utils.MakeDecision("EXPENSE", rule.Condition, amount)
	status := result.Status
	message := result.Message

	// create request
	expenseReq := &models.ExpenseRequest{
		EmployeeID: userID,
		Amount:     amount,
		Category:   category,
		Reason:     reason,
		Status:     status,
		RuleID:     &rule.ID,
	}

	err = s.expenseReqRepo.Create(ctx, tx, expenseReq)
	if err != nil {
		return "", "", apperrors.ErrInsertFailed
	}

	// deduct if auto-approved
	if status == constants.StatusAutoApproved {
		err = s.balanceRepo.DeductExpenseBalance(ctx, tx, userID, amount)
		if err != nil {
			return "", "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", apperrors.ErrTransactionCommit
	}

	return message, status, nil
}

// cancels an expense request
func (s *ExpenseService) CancelExpense(ctx context.Context, userID, requestID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	expenseReq, err := s.expenseReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return err
	}

	// Verify ownership
	if expenseReq.EmployeeID != userID {
		return apperrors.ErrExpenseRequestNotFound
	}

	// reuse CanCancel from apply_cancel_rules.go
	if err := utils.CanCancel(expenseReq.Status); err != nil {
		return err
	}

	err = s.expenseReqRepo.Cancel(ctx, tx, requestID)
	if err != nil {
		return err
	}

	if expenseReq.Status == constants.StatusAutoApproved {
		err = s.balanceRepo.RestoreExpenseBalance(ctx, tx, userID, expenseReq.Amount)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// ExpenseApprovalService handles business logic for expense approval operations
type ExpenseApprovalService struct {
	expenseReqRepo interfaces.ExpenseRequestRepository
	balanceRepo    interfaces.BalanceRepository
	userRepo       interfaces.UserRepository
	db             interfaces.DB
}

// NewExpenseApprovalService creates a new instance of ExpenseApprovalService
func NewExpenseApprovalService(
	ctx context.Context,
	expenseReqRepo interfaces.ExpenseRequestRepository,
	balanceRepo interfaces.BalanceRepository,
	userRepo interfaces.UserRepository,
	db interfaces.DB,
) interfaces.ExpenseApprovalService {
	return &ExpenseApprovalService{
		expenseReqRepo: expenseReqRepo,
		balanceRepo:    balanceRepo,
		userRepo:       userRepo,
		db:             db,
	}
}

// GetPendingExpenseRequests retrieves pending expense requests based on role
func (s *ExpenseApprovalService) GetPendingExpenseRequests(ctx context.Context, role string, approverID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	if role == constants.RoleManager {
		return s.expenseReqRepo.GetPendingForManager(ctx, approverID, limit, offset)
	} else if role == constants.RoleAdmin {
		return s.expenseReqRepo.GetPendingForAdmin(ctx, limit, offset)
	} else {
		return nil, 0, apperrors.ErrUnauthorized
	}
}

// approves an expense request
func (s *ExpenseApprovalService) ApproveExpense(
	ctx context.Context,
	role string,
	approverID, requestID int64,
	comment string,
) error {
	// check role
	if role == constants.RoleEmployee {
		return apperrors.ErrEmployeeCannotApprove
	}

	// validate comment
	if comment == "" {
		return apperrors.ErrCommentRequired
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	expenseReq, err := s.expenseReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return err
	}

	if approverID == expenseReq.EmployeeID {
		return apperrors.ErrSelfApprovalNotAllowed
	}

	// validate status
	if err := utils.ValidatePendingStatus(expenseReq.Status); err != nil {
		return err
	}

	// fetch role
	requesterRole, err := s.userRepo.GetRole(ctx, tx, expenseReq.EmployeeID)
	if err != nil {
		return err
	}

	// validate role
	if err := utils.ValidateApproverRole(role, requesterRole); err != nil {
		return err
	}

	// Deduct balance
	err = s.balanceRepo.DeductExpenseBalance(ctx, tx, expenseReq.EmployeeID, expenseReq.Amount)
	if err != nil {
		return err
	}

	// update request
	err = s.expenseReqRepo.UpdateStatus(ctx, tx, requestID, "APPROVED", approverID, comment)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// rejects an expense request
func (s *ExpenseApprovalService) RejectExpense(
	ctx context.Context,
	role string,
	approverID, requestID int64,
	comment string,
) error {
	// check role
	if role == constants.RoleEmployee {
		return apperrors.ErrEmployeeCannotApprove
	}

	// validate comment
	if comment == "" {
		return apperrors.ErrCommentRequired
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	expenseReq, err := s.expenseReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return err
	}

	if approverID == expenseReq.EmployeeID {
		return apperrors.ErrSelfApprovalNotAllowed
	}

	// 2. Validate pending status
	if err := utils.ValidatePendingStatus(expenseReq.Status); err != nil {
		return err
	}

	// 4. Fetch requester role
	requesterRole, err := s.userRepo.GetRole(ctx, tx, expenseReq.EmployeeID)
	if err != nil {
		return err
	}

	// 5. Validate approver role
	if err := utils.ValidateApproverRole(role, requesterRole); err != nil {
		return err
	}

	// 6. Default rejection comment
	if comment == "" {
		comment = "Rejected"
	}

	// 7. Update request
	err = s.expenseReqRepo.UpdateStatus(ctx, tx, requestID, "REJECTED", approverID, comment)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
