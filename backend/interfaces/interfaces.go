package interfaces

//go:generate go run github.com/vektra/mockery/v2@latest --all --dir . --output ../mocks --case underscore --with-expecter

import (
	"context"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Rows interface {
	Scan(dest ...any) error
	Next() bool
	Close()
	Err() error
}

type DB interface {
	Begin(ctx context.Context) (Tx, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// UserRepository definitions
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	Create(ctx context.Context, tx Tx, user *models.User) (int64, error)
	CheckEmailExists(ctx context.Context, tx Tx, email string) (bool, error)
	GetRole(ctx context.Context, tx Tx, userID int64) (string, error)
	GetGrade(ctx context.Context, tx Tx, userID int64) (int64, error)
}

// BalanceRepository definitions
type BalanceRepository interface {
	GetLeaveBalance(ctx context.Context, tx Tx, userID int64) (int, error)
	GetLeaveFullBalance(ctx context.Context, tx Tx, userID int64) (total int, remaining int, err error)
	GetExpenseBalance(ctx context.Context, tx Tx, userID int64) (float64, error)
	GetExpenseFullBalance(ctx context.Context, tx Tx, userID int64) (total float64, remaining float64, err error)
	GetDiscountBalance(ctx context.Context, tx Tx, userID int64) (float64, error)
	GetDiscountFullBalance(ctx context.Context, tx Tx, userID int64) (total float64, remaining float64, err error)
	DeductLeaveBalance(ctx context.Context, tx Tx, userID int64, days int) error
	DeductExpenseBalance(ctx context.Context, tx Tx, userID int64, amount float64) error
	DeductDiscountBalance(ctx context.Context, tx Tx, userID int64, percent float64) error
	RestoreLeaveBalance(ctx context.Context, tx Tx, userID int64, days int) error
	RestoreExpenseBalance(ctx context.Context, tx Tx, userID int64, amount float64) error
	RestoreDiscountBalance(ctx context.Context, tx Tx, userID int64, percent float64) error
	InitializeBalances(ctx context.Context, tx Tx, userID int64, gradeID int64) error
}

// RuleRepository definitions
type RuleRepository interface {
	GetByTypeAndGrade(ctx context.Context, requestType string, gradeID int64) (*models.Rule, error)
	Create(ctx context.Context, rule *models.Rule) error
	GetAll(ctx context.Context) ([]models.Rule, error)
	Update(ctx context.Context, ruleID int64, rule *models.Rule) error
	Delete(ctx context.Context, ruleID int64) error
}

// LeaveRequestRepository definitions
type LeaveRequestRepository interface {
	Create(ctx context.Context, tx Tx, req *models.LeaveRequest) error
	GetByID(ctx context.Context, tx Tx, requestID int64) (*models.LeaveRequest, error)
	UpdateStatus(ctx context.Context, tx Tx, requestID int64, status string, approverID int64, comment string) error
	GetPendingForManager(ctx context.Context, managerID int64, limit, offset int) ([]map[string]interface{}, int, error)
	GetPendingForAdmin(ctx context.Context, limit, offset int) ([]map[string]interface{}, int, error)
	CheckOverlap(ctx context.Context, userID int64, fromDate, toDate time.Time) (bool, error)
	Cancel(ctx context.Context, tx Tx, requestID int64) error
	GetPendingRequests(ctx context.Context) ([]struct {
		ID        int64
		CreatedAt time.Time
	}, error)
}

// ExpenseRequestRepository definitions
type ExpenseRequestRepository interface {
	Create(ctx context.Context, tx Tx, req *models.ExpenseRequest) error
	GetByID(ctx context.Context, tx Tx, requestID int64) (*models.ExpenseRequest, error)
	UpdateStatus(ctx context.Context, tx Tx, requestID int64, status string, approverID int64, comment string) error
	GetPendingForManager(ctx context.Context, managerID int64, limit, offset int) ([]map[string]interface{}, int, error)
	GetPendingForAdmin(ctx context.Context, limit, offset int) ([]map[string]interface{}, int, error)
	Cancel(ctx context.Context, tx Tx, requestID int64) error
	GetPendingRequests(ctx context.Context) ([]struct {
		ID        int64
		CreatedAt time.Time
	}, error)
}

// DiscountRequestRepository definitions
type DiscountRequestRepository interface {
	Create(ctx context.Context, tx Tx, req *models.DiscountRequest) error
	GetByID(ctx context.Context, tx Tx, requestID int64) (*models.DiscountRequest, error)
	UpdateStatus(ctx context.Context, tx Tx, requestID int64, status string, approverID int64, comment string) error
	GetPendingForManager(ctx context.Context, managerID int64, limit, offset int) ([]map[string]interface{}, int, error)
	GetPendingForAdmin(ctx context.Context, limit, offset int) ([]map[string]interface{}, int, error)
	Cancel(ctx context.Context, tx Tx, requestID int64) error
	GetPendingRequests(ctx context.Context) ([]struct {
		ID        int64
		CreatedAt time.Time
	}, error)
}

// GradeRepository handles grade data access operations
type GradeRepository interface {
	GetLimits(ctx context.Context, tx Tx, gradeID int64) (leaveLimit int, expenseLimit float64, discountLimit float64, err error)
}

// HolidayRepository handles holiday data access
type HolidayRepository interface {
	AddHoliday(ctx context.Context, date time.Time, desc string, adminID int64) error
	GetHolidays(ctx context.Context) ([]map[string]interface{}, error)
	DeleteHoliday(ctx context.Context, holidayID int64) error
	IsHoliday(ctx context.Context, date time.Time) (bool, error)
}

// MyRequestsRepository handles read-only queries for a user's own requests
type MyRequestsRepository interface {
	GetMyLeaveRequests(ctx context.Context, userID int64, limit, offset int) ([]map[string]interface{}, int, error)
	GetMyExpenseRequests(ctx context.Context, userID int64, limit, offset int) ([]map[string]interface{}, int, error)
	GetMyDiscountRequests(ctx context.Context, userID int64, limit, offset int) ([]map[string]interface{}, int, error)
	GetMyAllRequests(ctx context.Context, userID int64, limit, offset int) (leaves []map[string]interface{}, expenses []map[string]interface{}, discounts []map[string]interface{}, total int, err error)
	GetPendingAllRequests(ctx context.Context, role string, approverID int64, limit, offset int) (leaves []map[string]interface{}, expenses []map[string]interface{}, discounts []map[string]interface{}, total int, err error)
}

// ReportRepository handles analytical and statistical queries
type ReportRepository interface {
	GetRequestStatusDistribution(ctx context.Context) (map[string]int, error)
	GetRequestsByTypeReport(ctx context.Context) ([]models.RequestTypeReport, error)
}

// Service interfaces
type AuthService interface {
	RegisterUser(ctx context.Context, name, email, password string) error
	LoginUser(ctx context.Context, email, password string) (string, string, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

type LeaveService interface {
	ApplyLeave(ctx context.Context, userID int64, from time.Time, to time.Time, days int, leaveType string, reason string) (string, string, error)
	CancelLeave(ctx context.Context, userID, requestID int64) error
}

type LeaveApprovalService interface {
	GetPendingLeaveRequests(ctx context.Context, role string, approverID int64, limit, offset int) ([]map[string]interface{}, int, error)
	ApproveLeave(ctx context.Context, role string, approverID, requestID int64, approvalComment string) error
	RejectLeave(ctx context.Context, role string, approverID, requestID int64, rejectionComment string) error
}

type ExpenseService interface {
	ApplyExpense(ctx context.Context, userID int64, amount float64, category string, reason string) (string, string, error)
	CancelExpense(ctx context.Context, userID, requestID int64) error
}

type ExpenseApprovalService interface {
	GetPendingExpenseRequests(ctx context.Context, role string, approverID int64, limit, offset int) ([]map[string]interface{}, int, error)
	ApproveExpense(ctx context.Context, role string, approverID, requestID int64, comment string) error
	RejectExpense(ctx context.Context, role string, approverID, requestID int64, comment string) error
}

type RuleService interface {
	GetRule(ctx context.Context, requestType string, gradeID int64) (*models.Rule, error)
	CreateRule(ctx context.Context, role string, rule models.Rule) error
	GetRules(ctx context.Context, role string) ([]models.Rule, error)
	UpdateRule(ctx context.Context, role string, ruleID int64, rule models.Rule) error
	DeleteRule(ctx context.Context, role string, ruleID int64) error
}

type DiscountService interface {
	ApplyDiscount(ctx context.Context, userID int64, percent float64, reason string) (string, string, error)
	CancelDiscount(ctx context.Context, userID, requestID int64) error
}

type DiscountApprovalService interface {
	GetPendingRequests(ctx context.Context, role string, approverID int64, limit, offset int) ([]map[string]interface{}, int, error)
	ApproveDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error
	RejectDiscount(ctx context.Context, role string, approverID, requestID int64, comment string) error
}

type BalanceService interface {
	GetBalances(ctx context.Context, userID int64) (map[string]interface{}, error)
}

type HolidayService interface {
	AddHoliday(ctx context.Context, role string, adminID int64, date time.Time, desc string) error
	GetHolidays(ctx context.Context, role string) ([]map[string]interface{}, error)
	DeleteHoliday(ctx context.Context, role string, holidayID int64) error
}

type ReportService interface {
	GetDashboardSummary(ctx context.Context, role string) (map[string]interface{}, error)
	GetRequestStatusDistribution(ctx context.Context) (map[string]int, error)
	GetRequestsByTypeReport(ctx context.Context) ([]models.RequestTypeReport, error)
}

type MyRequestsService interface {
	GetMyRequests(ctx context.Context, userID int64, reqType string, limit, offset int) ([]map[string]interface{}, int, error)
	GetMyAllRequests(ctx context.Context, userID int64, limit, offset int) (map[string]interface{}, error)
	GetPendingAllRequests(ctx context.Context, role string, userID int64, limit, offset int) (map[string]interface{}, error)
}

type AutoRejectService interface {
	AutoRejectExpiredRequests(ctx context.Context) error
}
