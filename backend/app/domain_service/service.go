package domain_service

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

type DiscountService struct {
	discountReqRepo interfaces.DiscountRequestRepository
	balanceRepo     interfaces.BalanceRepository
	ruleService     interfaces.RuleService
	userRepo        interfaces.UserRepository
	db              interfaces.DB
}

func NewDiscountService(
	ctx context.Context,
	discountReqRepo interfaces.DiscountRequestRepository,
	balanceRepo interfaces.BalanceRepository,
	ruleService interfaces.RuleService,
	userRepo interfaces.UserRepository,
	db interfaces.DB,
) interfaces.DiscountService {
	return &DiscountService{
		discountReqRepo: discountReqRepo,
		balanceRepo:     balanceRepo,
		ruleService:     ruleService,
		userRepo:        userRepo,
		db:              db,
	}
}

func (s *DiscountService) ApplyDiscount(
	ctx context.Context,
	userID int64,
	percent float64,
	reason string,
) (string, string, error) {

	// validations
	if userID <= 0 {
		return "", "", apperrors.ErrInvalidUser
	}

	if percent <= 0 {
		return "", "", apperrors.ErrInvalidDiscountPercent
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", apperrors.ErrTransactionBegin
	}
	defer tx.Rollback(ctx)

	// fetch remaining
	remaining, err := s.balanceRepo.GetDiscountBalance(ctx, tx, userID)
	if err != nil {
		return "", "", err
	}

	if percent > remaining {
		return "", "", apperrors.ErrDiscountLimitExceeded
	}

	// user grade
	gradeID, err := s.userRepo.GetGrade(ctx, tx, userID)
	if err != nil {
		return "", "", err
	}

	// fetch rule
	rule, err := s.ruleService.GetRule(ctx, "DISCOUNT", gradeID)
	if err != nil {
		return "", "", apperrors.ErrRuleNotFound
	}

	// apply rule
	result := utils.MakeDecision("DISCOUNT", rule.Condition, percent)
	status := result.Status
	message := result.Message

	// create request
	discountReq := &models.DiscountRequest{
		EmployeeID:         userID,
		DiscountPercentage: percent,
		Reason:             reason,
		Status:             status,
		RuleID:             &rule.ID,
	}

	err = s.discountReqRepo.Create(ctx, tx, discountReq)
	if err != nil {
		return "", "", apperrors.ErrInsertFailed
	}

	// deduct if auto-approved
	if status == constants.StatusAutoApproved {
		err = s.balanceRepo.DeductDiscountBalance(ctx, tx, userID, percent)
		if err != nil {
			return "", "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", apperrors.ErrTransactionCommit
	}

	return message, status, nil
}

func (s *DiscountService) CancelDiscount(ctx context.Context, userID, requestID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperrors.ErrTransactionBegin
	}
	defer tx.Rollback(ctx)

	discountReq, err := s.discountReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return apperrors.ErrDiscountRequestNotFound
	}

	if discountReq.EmployeeID != userID {
		return apperrors.ErrDiscountRequestNotFound
	}

	// reuse CanCancel from apply_cancel_rules.go
	if err := utils.CanCancel(discountReq.Status); err != nil {
		return err
	}

	err = s.discountReqRepo.Cancel(ctx, tx, requestID)
	if err != nil {
		return apperrors.ErrUpdateFailed
	}

	if discountReq.Status == constants.StatusAutoApproved {
		err = s.balanceRepo.RestoreDiscountBalance(ctx, tx, userID, discountReq.DiscountPercentage)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

type DiscountApprovalService struct {
	discountReqRepo interfaces.DiscountRequestRepository
	balanceRepo     interfaces.BalanceRepository
	userRepo        interfaces.UserRepository
	db              interfaces.DB
}

func NewDiscountApprovalService(
	ctx context.Context,
	discountReqRepo interfaces.DiscountRequestRepository,
	balanceRepo interfaces.BalanceRepository,
	userRepo interfaces.UserRepository,
	db interfaces.DB,
) interfaces.DiscountApprovalService {
	return &DiscountApprovalService{
		discountReqRepo: discountReqRepo,
		balanceRepo:     balanceRepo,
		userRepo:        userRepo,
		db:              db,
	}
}

func (s *DiscountApprovalService) GetPendingRequests(ctx context.Context, role string, approverID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	if role == constants.RoleManager {
		return s.discountReqRepo.GetPendingForManager(ctx, approverID, limit, offset)
	} else if role == constants.RoleAdmin {
		return s.discountReqRepo.GetPendingForAdmin(ctx, limit, offset)
	} else {
		return nil, 0, apperrors.ErrUnauthorized
	}
}

func (s *DiscountApprovalService) ApproveDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error {
	if role == constants.RoleEmployee {
		return apperrors.ErrEmployeeCannotApprove
	}

	if comment == "" {
		return apperrors.ErrCommentRequired
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	discountReq, err := s.discountReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return apperrors.ErrDiscountRequestNotFound
	}

	if approverID == discountReq.EmployeeID {
		return apperrors.ErrSelfApprovalNotAllowed
	}

	if err := utils.ValidatePendingStatus(discountReq.Status); err != nil {
		return err
	}

	requesterRole, err := s.userRepo.GetRole(ctx, tx, discountReq.EmployeeID)
	if err != nil {
		return err
	}

	if err := utils.ValidateApproverRole(role, requesterRole); err != nil {
		return err
	}

	// Update request
	err = s.discountReqRepo.UpdateStatus(ctx, tx, requestID, "APPROVED", approverID, comment)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *DiscountApprovalService) RejectDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error {
	if role == constants.RoleEmployee {
		return apperrors.ErrEmployeeCannotApprove
	}

	if comment == "" {
		return apperrors.ErrCommentRequired
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	discountReq, err := s.discountReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return apperrors.ErrDiscountRequestNotFound
	}

	if approverID == discountReq.EmployeeID {
		return apperrors.ErrSelfApprovalNotAllowed
	}

	if err := utils.ValidatePendingStatus(discountReq.Status); err != nil {
		return err
	}

	requesterRole, err := s.userRepo.GetRole(ctx, tx, discountReq.EmployeeID)
	if err != nil {
		return err
	}

	if err := utils.ValidateApproverRole(role, requesterRole); err != nil {
		return err
	}

	// Update request
	err = s.discountReqRepo.UpdateStatus(ctx, tx, requestID, "REJECTED", approverID, comment)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type BalanceService struct {
	balanceRepo interfaces.BalanceRepository
	db          interfaces.DB
}

func NewBalanceService(ctx context.Context, balanceRepo interfaces.BalanceRepository, db interfaces.DB) interfaces.BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		db:          db,
	}
}

func (s *BalanceService) GetBalances(ctx context.Context, userID int64) (map[string]interface{}, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var leaveTotal, leaveRemaining int
	var expenseTotal, expenseRemaining float64
	var discountTotal, discountRemaining float64

	leaveRemaining, err = s.balanceRepo.GetLeaveBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	// Note: interface has GetLeaveFullBalance too
	leaveTotal, leaveRemaining, err = s.balanceRepo.GetLeaveFullBalance(ctx, tx, userID)

	expenseTotal, expenseRemaining, err = s.balanceRepo.GetExpenseFullBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	discountTotal, discountRemaining, err = s.balanceRepo.GetDiscountFullBalance(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"leave": map[string]interface{}{
			"total":     leaveTotal,
			"remaining": leaveRemaining,
		},
		"expense": map[string]interface{}{
			"total":     expenseTotal,
			"remaining": expenseRemaining,
		},
		"discount": map[string]interface{}{
			"total":     discountTotal,
			"remaining": discountRemaining,
		},
	}, nil
}
