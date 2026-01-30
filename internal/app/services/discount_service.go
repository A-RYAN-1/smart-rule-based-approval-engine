package services

import (
	"context"

	"rule-based-approval-engine/internal/app/repositories"
	"rule-based-approval-engine/internal/app/services/helpers"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DiscountService struct {
	ruleService *RuleService
	userRepo    repositories.UserRepository
	db          *pgxpool.Pool
}

func NewDiscountService(ruleService *RuleService, userRepo repositories.UserRepository, db *pgxpool.Pool) *DiscountService {
	return &DiscountService{
		ruleService: ruleService,
		userRepo:    userRepo,
		db:          db,
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
	var remaining float64
	err = tx.QueryRow(
		ctx,
		`SELECT remaining_discount FROM discount WHERE user_id=$1`,
		userID,
	).Scan(&remaining)

	if err != nil {
		return "", "", apperrors.ErrBalanceFetchFailed
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
	result := helpers.MakeDecision("DISCOUNT", rule.Condition, percent)
	status := result.Status
	message := result.Message

	// create request
	_, err = tx.Exec(
		ctx,
		`INSERT INTO discount_requests
		 (employee_id, discount_percentage, reason, status, rule_id)
		 VALUES ($1,$2,$3,$4,$5)`,
		userID, percent, reason, status, rule.ID,
	)
	if err != nil {
		return "", "", apperrors.ErrInsertFailed
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

	var status string

	err = tx.QueryRow(
		ctx,
		`SELECT status FROM discount_requests WHERE id=$1 AND employee_id=$2`,
		requestID, userID,
	).Scan(&status)

	if err != nil {
		return apperrors.ErrDiscountRequestNotFound
	}

	// reuse CanCancel from apply_cancel_rules.go
	if err := helpers.CanCancel(status); err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE discount_requests SET status='CANCELLED' WHERE id=$1`,
		requestID,
	)
	if err != nil {
		return apperrors.ErrUpdateFailed
	}

	return tx.Commit(ctx)
}

func (s *DiscountService) GetPendingDiscountRequests(ctx context.Context, role string, approverID int64) ([]map[string]interface{}, error) {
	var rows interface {
		Next() bool
		Scan(dest ...interface{}) error
		Close()
	}
	var err error

	if role == "MANAGER" {
		rows, err = s.db.Query(ctx, `
			SELECT dr.id, u.name, dr.discount_percentage, dr.reason, dr.created_at
			FROM discount_requests dr
			JOIN users u ON dr.employee_id = u.id
			WHERE dr.status='PENDING' AND u.manager_id=$1
		`, approverID)
	} else if role == "ADMIN" {
		rows, err = s.db.Query(ctx, `
			SELECT dr.id, u.name, dr.discount_percentage, dr.reason, dr.created_at
			FROM discount_requests dr
			JOIN users u ON dr.employee_id = u.id
			WHERE dr.status='PENDING'
		`)
	} else {
		return nil, apperrors.ErrUnauthorized
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, reason string
		var percent float64
		var createdAt interface{}
		if err := rows.Scan(&id, &name, &percent, &reason, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"id":                  id,
			"employee":            name,
			"discount_percentage": percent,
			"reason":              reason,
			"created_at":          createdAt,
		})
	}
	return result, nil
}

func (s *DiscountService) ApproveDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error {
	// check role
	if role == "EMPLOYEE" {
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

	var employeeID int64
	var status string
	err = tx.QueryRow(ctx, `SELECT employee_id, status FROM discount_requests WHERE id=$1`, requestID).Scan(&employeeID, &status)
	if err != nil {
		return apperrors.ErrDiscountRequestNotFound
	}

	if approverID == employeeID {
		return apperrors.ErrSelfApprovalNotAllowed
	}

	if status != "PENDING" {
		return apperrors.ErrDiscountRequestNotPending
	}

	var requesterRole string
	err = tx.QueryRow(ctx, `SELECT role FROM users WHERE id=$1`, employeeID).Scan(&requesterRole)
	if err != nil {
		return err
	}

	if err := helpers.ValidateApproverRole(role, requesterRole); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE discount_requests
		SET status='APPROVED', approved_by_id=$1, approval_comment=$2
		WHERE id=$3
	`, approverID, comment, requestID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *DiscountService) RejectDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error {
	// check role
	if role == "EMPLOYEE" {
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

	var employeeID int64
	var status string
	err = tx.QueryRow(ctx, `SELECT employee_id, status FROM discount_requests WHERE id=$1`, requestID).Scan(&employeeID, &status)
	if err != nil {
		return apperrors.ErrDiscountRequestNotFound
	}

	if approverID == employeeID {
		return apperrors.ErrSelfApprovalNotAllowed
	}

	if status != "PENDING" {
		return apperrors.ErrDiscountRequestNotPending
	}

	var requesterRole string
	err = tx.QueryRow(ctx, `SELECT role FROM users WHERE id=$1`, employeeID).Scan(&requesterRole)
	if err != nil {
		return err
	}

	if err := helpers.ValidateApproverRole(role, requesterRole); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE discount_requests
		SET status='REJECTED', approved_by_id=$1, approval_comment=$2
		WHERE id=$3
	`, approverID, comment, requestID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
