package repositories

import (
	"context"
	"time"

	"rule-based-approval-engine/internal/models"

	"github.com/jackc/pgx/v5"
)

// UserRepository
type UserRepository interface {
	// find by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// find by ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// create user
	Create(ctx context.Context, tx pgx.Tx, user *models.User) (int64, error)

	// check email
	CheckEmailExists(ctx context.Context, tx pgx.Tx, email string) (bool, error)

	// get role
	GetRole(ctx context.Context, tx pgx.Tx, userID int64) (string, error)

	// get grade
	GetGrade(ctx context.Context, tx pgx.Tx, userID int64) (int64, error)
}

// LeaveRequestRepository
type LeaveRequestRepository interface {
	// create request
	Create(ctx context.Context, tx pgx.Tx, req *models.LeaveRequest) error

	// find by ID
	GetByID(ctx context.Context, tx pgx.Tx, requestID int64) (*models.LeaveRequest, error)

	// update status
	UpdateStatus(ctx context.Context, tx pgx.Tx, requestID int64, status string, approverID int64, comment string) error

	// pending for manager
	GetPendingForManager(ctx context.Context, managerID int64) ([]map[string]interface{}, error)

	// pending for admin
	GetPendingForAdmin(ctx context.Context) ([]map[string]interface{}, error)

	// check overlap
	CheckOverlap(ctx context.Context, userID int64, fromDate, toDate time.Time) (bool, error)

	// cancel request
	Cancel(ctx context.Context, tx pgx.Tx, requestID int64) error
}

// ExpenseRequestRepository
type ExpenseRequestRepository interface {
	// create request
	Create(ctx context.Context, tx pgx.Tx, req *models.ExpenseRequest) error
	// find by ID
	GetByID(ctx context.Context, tx pgx.Tx, requestID int64) (*models.ExpenseRequest, error)
	// update status
	UpdateStatus(ctx context.Context, tx pgx.Tx, requestID int64, status string, approverID int64, comment string) error
	// pending for manager
	GetPendingForManager(ctx context.Context, managerID int64) ([]map[string]interface{}, error)
	// pending for admin
	GetPendingForAdmin(ctx context.Context) ([]map[string]interface{}, error)
	// cancel request
	Cancel(ctx context.Context, tx pgx.Tx, requestID int64) error
}

// RuleRepository
type RuleRepository interface {
	// get by type/grade
	GetByTypeAndGrade(ctx context.Context, requestType string, gradeID int64) (*models.Rule, error)

	// create/update rule
	Create(ctx context.Context, rule *models.Rule) error

	// get all rules
	GetAll(ctx context.Context) ([]models.Rule, error)

	// update rule
	Update(ctx context.Context, ruleID int64, rule *models.Rule) error

	// delete rule
	Delete(ctx context.Context, ruleID int64) error
}

// BalanceRepository
type BalanceRepository interface {
	// get leave balance
	GetLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64) (int, error)

	// get expense balance
	GetExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64) (float64, error)

	// deduct leave
	DeductLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64, days int) error

	// deduct expense
	DeductExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64, amount float64) error

	// restore leave
	RestoreLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64, days int) error

	// restore expense
	RestoreExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64, amount float64) error

	// init balances
	InitializeBalances(ctx context.Context, tx pgx.Tx, userID int64, gradeID int64) error
}

// GradeRepository handles grade data access operations
type GradeRepository interface {
	// GetLimits retrieves the leave, expense, and discount limits for a grade within a transaction
	GetLimits(ctx context.Context, tx pgx.Tx, gradeID int64) (leaveLimit int, expenseLimit float64, discountLimit float64, err error)
}

// MyRequestsRepository handles read-only queries for a user's own requests
type MyRequestsRepository interface {
	GetMyLeaveRequests(ctx context.Context, userID int64) ([]map[string]interface{}, error)
	GetMyExpenseRequests(ctx context.Context, userID int64) ([]map[string]interface{}, error)
	GetMyDiscountRequests(ctx context.Context, userID int64) ([]map[string]interface{}, error)
}

// HolidayRepository handles holiday data access
type HolidayRepository interface {
	AddHoliday(ctx context.Context, date time.Time, desc string, adminID int64) error
	GetHolidays(ctx context.Context) ([]map[string]interface{}, error)
	DeleteHoliday(ctx context.Context, holidayID int64) error
	IsHoliday(ctx context.Context, date time.Time) (bool, error)
}

// ReportRepository handles analytical and statistical queries
type ReportRepository interface {
	GetRequestStatusDistribution(ctx context.Context) (map[string]int, error)
	GetRequestsByTypeReport(ctx context.Context) ([]models.RequestTypeReport, error)
}
