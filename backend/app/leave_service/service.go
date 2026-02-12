package leave_service

import (
	"context"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

// handles business logic for leave requests
type LeaveService struct {
	leaveReqRepo interfaces.LeaveRequestRepository
	balanceRepo  interfaces.BalanceRepository
	ruleService  interfaces.RuleService
	userRepo     interfaces.UserRepository
	db           interfaces.DB
}

// creates a new instance of LeaveService
func NewLeaveService(
	ctx context.Context,
	leaveReqRepo interfaces.LeaveRequestRepository,
	balanceRepo interfaces.BalanceRepository,
	ruleService interfaces.RuleService,
	userRepo interfaces.UserRepository,
	db interfaces.DB,
) interfaces.LeaveService {
	return &LeaveService{
		leaveReqRepo: leaveReqRepo,
		balanceRepo:  balanceRepo,
		ruleService:  ruleService,
		userRepo:     userRepo,
		db:           db,
	}
}

// processes a leave application
func (s *LeaveService) ApplyLeave(
	ctx context.Context,
	userID int64,
	from time.Time,
	to time.Time,
	days int,
	leaveType string,
	reason string,
) (string, string, error) {
	// validations
	if userID <= 0 {
		return "", "", apperrors.ErrInvalidUser
	}

	if days <= 0 {
		return "", "", apperrors.ErrInvalidLeaveDays
	}

	if from.After(to) {
		return "", "", apperrors.ErrInvalidDateRange
	}

	// date validation
	today := time.Now().Truncate(24 * time.Hour)
	if from.Before(today) {
		return "", "", apperrors.ErrPastDate
	}

	// check overlap
	overlap, err := s.leaveReqRepo.CheckOverlap(ctx, userID, from, to)
	if err != nil {
		return "", "", apperrors.ErrLeaveVerificationFailed
	}

	if overlap {
		return "", "", apperrors.ErrLeaveOverlap
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", apperrors.ErrTransactionBegin
	}
	defer tx.Rollback(ctx)

	// leave balance
	remaining, err := s.balanceRepo.GetLeaveBalance(ctx, tx, userID)
	if err != nil {
		return "", "", err
	}

	if days > remaining {
		return "", "", apperrors.ErrLeaveBalanceExceeded
	}

	// user grade
	gradeID, err := s.userRepo.GetGrade(ctx, tx, userID)
	if err != nil {
		return "", "", err
	}

	// fetch rule
	rule, err := s.ruleService.GetRule(ctx, "LEAVE", gradeID)
	if err != nil {
		return "", "", apperrors.ErrRuleNotFound
	}

	// apply rule
	result := utils.MakeDecision("LEAVE", rule.Condition, float64(days))
	status := result.Status
	message := result.Message

	leaveReq := &models.LeaveRequest{
		EmployeeID: userID,
		FromDate:   from,
		ToDate:     to,
		Reason:     reason,
		LeaveType:  leaveType,
		Status:     status,
		RuleID:     &rule.ID,
	}

	err = s.leaveReqRepo.Create(ctx, tx, leaveReq)
	if err != nil {
		return "", "", utils.MapPgError(err)
	}

	// deduct if auto-approved
	if status == constants.StatusAutoApproved {
		err = s.balanceRepo.DeductLeaveBalance(ctx, tx, userID, days)
		if err != nil {
			return "", "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", apperrors.ErrTransactionCommit
	}

	return message, status, nil
}

// cancels a leave request
func (s *LeaveService) CancelLeave(ctx context.Context, userID, requestID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	leaveReq, err := s.leaveReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return err
	}

	// Verify ownership
	if leaveReq.EmployeeID != userID {
		return apperrors.ErrLeaveRequestNotFound
	}

	// reuse CanCancel from apply_cancel_rules.go
	if err := utils.CanCancel(leaveReq.Status); err != nil {
		return err
	}

	days := utils.CalculateLeaveDays(leaveReq.FromDate, leaveReq.ToDate)

	err = s.leaveReqRepo.Cancel(ctx, tx, requestID)
	if err != nil {
		return err
	}

	if leaveReq.Status == constants.StatusAutoApproved {
		err = s.balanceRepo.RestoreLeaveBalance(ctx, tx, userID, days)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// handles business logic for leave approval operations
type LeaveApprovalService struct {
	leaveReqRepo interfaces.LeaveRequestRepository
	balanceRepo  interfaces.BalanceRepository
	userRepo     interfaces.UserRepository
	db           interfaces.DB
}

// creates a new instance of LeaveApprovalService
func NewLeaveApprovalService(
	ctx context.Context,
	leaveReqRepo interfaces.LeaveRequestRepository,
	balanceRepo interfaces.BalanceRepository,
	userRepo interfaces.UserRepository,
	db interfaces.DB,
) interfaces.LeaveApprovalService {
	return &LeaveApprovalService{
		leaveReqRepo: leaveReqRepo,
		balanceRepo:  balanceRepo,
		userRepo:     userRepo,
		db:           db,
	}
}

// retrieves pending leave requests based on role
func (s *LeaveApprovalService) GetPendingLeaveRequests(ctx context.Context, role string, approverID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	switch role {
	case constants.RoleManager:
		return s.leaveReqRepo.GetPendingForManager(ctx, approverID, limit, offset)
	case constants.RoleAdmin:
		return s.leaveReqRepo.GetPendingForAdmin(ctx, limit, offset)
	default:
		return nil, 0, apperrors.ErrUnauthorizedRole
	}
}

// approves a leave request
func (s *LeaveApprovalService) ApproveLeave(
	ctx context.Context,
	role string,
	approverID, requestID int64,
	approvalComment string,
) error {
	// check role
	if role == constants.RoleEmployee {
		return apperrors.ErrEmployeeCannotApprove
	}

	// validate comment
	if approvalComment == "" {
		return apperrors.ErrCommentRequired
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	leaveReq, err := s.leaveReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return err
	}

	if approverID == leaveReq.EmployeeID {
		return apperrors.ErrSelfApprovalNotAllowed
	}

	if err := utils.ValidatePendingStatus(leaveReq.Status); err != nil {
		return err
	}

	// Authorization against requester
	requesterRole, err := s.userRepo.GetRole(ctx, tx, leaveReq.EmployeeID)
	if err != nil {
		return err
	}

	if err := utils.ValidateApproverRole(role, requesterRole); err != nil {
		return err
	}

	days := utils.CalculateLeaveDays(leaveReq.FromDate, leaveReq.ToDate)

	// Deduct leave balance
	err = s.balanceRepo.DeductLeaveBalance(ctx, tx, leaveReq.EmployeeID, days)
	if err != nil {
		return err
	}

	// Default comment if not provided
	if approvalComment == "" {
		approvalComment = constants.StatusApproved
	}

	// Update request
	err = s.leaveReqRepo.UpdateStatus(ctx, tx, requestID, "APPROVED", approverID, approvalComment)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// RejectLeave rejects a leave request
func (s *LeaveApprovalService) RejectLeave(
	ctx context.Context,
	role string,
	approverID, requestID int64,
	rejectionComment string,
) error {
	// check role
	if role == constants.RoleEmployee {
		return apperrors.ErrEmployeeCannotApprove
	}

	// validate comment
	if rejectionComment == "" {
		return apperrors.ErrCommentRequired
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	leaveReq, err := s.leaveReqRepo.GetByID(ctx, tx, requestID)
	if err != nil {
		return err
	}

	if err := utils.ValidatePendingStatus(leaveReq.Status); err != nil {
		return err
	}

	if approverID == leaveReq.EmployeeID {
		return apperrors.ErrSelfApprovalNotAllowed
	}

	// Authorization
	requesterRole, err := s.userRepo.GetRole(ctx, tx, leaveReq.EmployeeID)
	if err != nil {
		return err
	}

	if err := utils.ValidateApproverRole(role, requesterRole); err != nil {
		return err
	}

	// Update request
	err = s.leaveReqRepo.UpdateStatus(ctx, tx, requestID, "REJECTED", approverID, rejectionComment)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
