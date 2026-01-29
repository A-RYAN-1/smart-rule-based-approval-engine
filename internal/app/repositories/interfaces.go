package repositories

import (
	"context"
	"time"

	"rule-based-approval-engine/internal/models"

	"github.com/jackc/pgx/v5"
)

// UserRepository handles user data access operations
type UserRepository interface {
	// GetByEmail retrieves a user by email address
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// Create inserts a new user within a transaction
	Create(ctx context.Context, tx pgx.Tx, user *models.User) (int64, error)

	// CheckEmailExists checks if an email already exists within a transaction
	CheckEmailExists(ctx context.Context, tx pgx.Tx, email string) (bool, error)

	// GetRole retrieves a user's role within a transaction
	GetRole(ctx context.Context, tx pgx.Tx, userID int64) (string, error)

	// GetGrade retrieves a user's grade ID within a transaction
	GetGrade(ctx context.Context, tx pgx.Tx, userID int64) (int64, error)
}

// LeaveRequestRepository handles leave request data access operations
type LeaveRequestRepository interface {
	// Create inserts a new leave request within a transaction
	Create(ctx context.Context, tx pgx.Tx, req *models.LeaveRequest) error

	// GetByID retrieves a leave request by ID within a transaction
	GetByID(ctx context.Context, tx pgx.Tx, requestID int64) (*models.LeaveRequest, error)

	// UpdateStatus updates the status of a leave request within a transaction
	UpdateStatus(ctx context.Context, tx pgx.Tx, requestID int64, status string, approverID int64, comment string) error

	// GetPendingForManager retrieves pending leave requests for a specific manager
	GetPendingForManager(ctx context.Context, managerID int64) ([]map[string]interface{}, error)

	// GetPendingForAdmin retrieves all pending leave requests
	GetPendingForAdmin(ctx context.Context) ([]map[string]interface{}, error)

	// CheckOverlap checks if a leave request overlaps with existing requests
	CheckOverlap(ctx context.Context, userID int64, fromDate, toDate time.Time) (bool, error)

	// Cancel cancels a leave request within a transaction
	Cancel(ctx context.Context, tx pgx.Tx, requestID int64) error
}

// ExpenseRequestRepository handles expense request data access operations
type ExpenseRequestRepository interface {
	// Create inserts a new expense request within a transaction
	Create(ctx context.Context, tx pgx.Tx, req *models.ExpenseRequest) error

	// GetByID retrieves an expense request by ID within a transaction
	GetByID(ctx context.Context, tx pgx.Tx, requestID int64) (*models.ExpenseRequest, error)

	// UpdateStatus updates the status of an expense request within a transaction
	UpdateStatus(ctx context.Context, tx pgx.Tx, requestID int64, status string, approverID int64, comment string) error

	// GetPendingForManager retrieves pending expense requests for a specific manager
	GetPendingForManager(ctx context.Context, managerID int64) ([]map[string]interface{}, error)

	// GetPendingForAdmin retrieves all pending expense requests
	GetPendingForAdmin(ctx context.Context) ([]map[string]interface{}, error)

	// Cancel cancels an expense request within a transaction
	Cancel(ctx context.Context, tx pgx.Tx, requestID int64) error
}

// RuleRepository handles rule data access operations
type RuleRepository interface {
	// GetByTypeAndGrade retrieves a rule by request type and grade ID
	GetByTypeAndGrade(ctx context.Context, requestType string, gradeID int64) (*models.Rule, error)

	// Create inserts or updates a rule
	Create(ctx context.Context, rule *models.Rule) error

	// GetAll retrieves all rules
	GetAll(ctx context.Context) ([]models.Rule, error)

	// Update updates an existing rule
	Update(ctx context.Context, ruleID int64, rule *models.Rule) error

	// Delete deletes a rule by ID
	Delete(ctx context.Context, ruleID int64) error
}

// BalanceRepository handles balance data access operations for leave and expense wallets
type BalanceRepository interface {
	// GetLeaveBalance retrieves the remaining leave balance for a user within a transaction
	GetLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64) (int, error)

	// GetExpenseBalance retrieves the remaining expense balance for a user within a transaction
	GetExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64) (float64, error)

	// DeductLeaveBalance deducts leave days from a user's balance within a transaction
	DeductLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64, days int) error

	// DeductExpenseBalance deducts expense amount from a user's balance within a transaction
	DeductExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64, amount float64) error

	// RestoreLeaveBalance restores leave days to a user's balance within a transaction
	RestoreLeaveBalance(ctx context.Context, tx pgx.Tx, userID int64, days int) error

	// RestoreExpenseBalance restores expense amount to a user's balance within a transaction
	RestoreExpenseBalance(ctx context.Context, tx pgx.Tx, userID int64, amount float64) error

	// InitializeBalances initializes leave, expense, and discount balances for a new user within a transaction
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
