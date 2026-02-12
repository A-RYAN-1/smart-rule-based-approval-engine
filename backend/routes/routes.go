package routes

import (
	"context"

	"github.com/ankita-advitot/rule_based_approval_engine/app/auth"
	"github.com/ankita-advitot/rule_based_approval_engine/app/domain_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/expense_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/holidays"
	"github.com/ankita-advitot/rule_based_approval_engine/app/leave_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/my_requests"
	"github.com/ankita-advitot/rule_based_approval_engine/app/reports"
	"github.com/ankita-advitot/rule_based_approval_engine/app/rules"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func Register(
	ctx context.Context,
	router *gin.Engine,
	authService interfaces.AuthService,
	leaveService interfaces.LeaveService,
	expenseService interfaces.ExpenseService,
	leaveApprovalService interfaces.LeaveApprovalService,
	expenseApprovalService interfaces.ExpenseApprovalService,
	ruleService interfaces.RuleService,
	myRequestsService interfaces.MyRequestsService,
	holidayService interfaces.HolidayService,
	reportService interfaces.ReportService,
	balanceService interfaces.BalanceService,
	autoRejectService interfaces.AutoRejectService,
	discountService interfaces.DiscountService,
	discountApprovalService interfaces.DiscountApprovalService,
) {
	// Initialize handlers
	authHandler := auth.NewAuthHandler(ctx, authService)
	leaveHandler := leave_service.NewLeaveHandler(ctx, leaveService)
	leaveApprovalHandler := leave_service.NewLeaveApprovalHandler(ctx, leaveApprovalService)
	expenseHandler := expense_service.NewExpenseHandler(ctx, expenseService)
	expenseApprovalHandler := expense_service.NewExpenseApprovalHandler(ctx, expenseApprovalService)
	ruleHandler := rules.NewRuleHandler(ctx, ruleService)
	myRequestsHandler := my_requests.NewMyRequestsHandler(ctx, myRequestsService)
	holidayHandler := holidays.NewHolidayHandler(ctx, holidayService)
	reportHandler := reports.NewReportHandler(ctx, reportService)
	balanceHandler := domain_service.NewBalanceHandler(ctx, balanceService)
	discountHandler := domain_service.NewDiscountHandler(ctx, discountService)
	discountApprovalHandler := domain_service.NewDiscountApprovalHandler(ctx, discountApprovalService)

	// Public routes
	public := router.Group("/api")
	{
		// Auth routes (some frontends use /auth/login, some /api/login, supporting /auth via alias if needed, but here keeping /api for now as per original, but issue says frontend uses /auth/login. Wait, issue says "The backend only provides /auth/login". 
        // Actually the original code had public.POST("/login", ...) under /api group, so it was /api/login.
        // Frontend auth.service.ts calls /auth/login.
        // So I need to move these or alias them.
        // I will create a separate base group for /auth if it's at root, or /api/auth if strictly under api.
        // Frontend: `api.post<any>('/auth/login', ...)` with baseURL `VITE_API_URL || '/api'`.
        // If VITE_API_URL is empty, it calls `/api/auth/login`. If VITE_API_URL is `http://localhost:8080`, it calls `.../auth/login`.
        // The safest/most standard way if `api` object has baseURL `/api`:
        // Request is `/api/auth/login`.
        // So I should register `public.Group("/auth")` inside `/api`.
	}
    
    // Auth routes under /api/auth to match frontend's `/api/auth/login` (assuming standard axios baseURL setup)
    authGroup := public.Group("/auth")
    {
        authGroup.POST("/register", authHandler.Register)
        authGroup.POST("/login", authHandler.Login)
        authGroup.POST("/logout", authHandler.Logout) // Added logout
    }

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
        // User Info
        protected.GET("/me", authHandler.GetMe)

		// Leave routes
        // Singular (keep for backward compat if needed, or just replace)
		protected.POST("/leave/apply", leaveHandler.ApplyLeave)
        
        // Plural (Frontend expects these)
        leaves := protected.Group("/leaves")
        {
            leaves.POST("/request", leaveHandler.ApplyLeave) // Alias for apply
            leaves.POST("/:id/cancel", leaveHandler.CancelLeave)
            leaves.GET("/my", myRequestsHandler.GetMyLeaves)
            
            // Approval routes
            leaves.GET("/pending", leaveApprovalHandler.GetPendingLeaves)
            leaves.POST("/:id/approve", leaveApprovalHandler.ApproveLeave)
            leaves.POST("/:id/reject", leaveApprovalHandler.RejectLeave)
        }

		// Expense routes
        expenses := protected.Group("/expenses")
        {
            expenses.POST("/request", expenseHandler.ApplyExpense)
            expenses.POST("/:id/cancel", expenseHandler.CancelExpense)
            expenses.GET("/my", myRequestsHandler.GetMyExpenses)
            
            expenses.GET("/pending", expenseApprovalHandler.GetPendingExpenses)
            expenses.POST("/:id/approve", expenseApprovalHandler.ApproveExpense)
            expenses.POST("/:id/reject", expenseApprovalHandler.RejectExpense)
        }

        // Discount routes
        discounts := protected.Group("/discounts")
        {
             discounts.POST("/request", discountHandler.ApplyDiscount)
             discounts.POST("/:id/cancel", discountHandler.CancelDiscount)
             discounts.GET("/my", myRequestsHandler.GetMyDiscounts)

             discounts.GET("/pending", discountApprovalHandler.GetPendingDiscounts)
             discounts.POST("/:id/approve", discountApprovalHandler.ApproveDiscount)
             discounts.POST("/:id/reject", discountApprovalHandler.RejectDiscount)
        }

		// Rule routes
        // Moved to admin group, but keeping aliases or just moving?
        // Issue #5: "Frontend calls /admin/rules"
        admin := protected.Group("/admin")
        {
            admin.POST("/rules", ruleHandler.CreateRule)
            admin.GET("/rules", ruleHandler.GetRules)
            admin.PUT("/rules/:id", ruleHandler.UpdateRule)
            admin.DELETE("/rules/:id", ruleHandler.DeleteRule)

            admin.POST("/holidays", holidayHandler.AddHoliday)
            admin.GET("/holidays", holidayHandler.GetHolidays)
            admin.DELETE("/holidays/:id", holidayHandler.DeleteHoliday)

            // Admin Reports
            admin.GET("/reports/request-status-distribution", reportHandler.GetRequestStatusDistribution)
            admin.GET("/reports/requests-by-type", reportHandler.GetRequestsByType)
        }

		// My Requests routes (Legacy/General)
		protected.GET("/my-requests", myRequestsHandler.GetMyRequests)
		protected.GET("/my-requests/all", myRequestsHandler.GetMyAllRequests)
		protected.GET("/pending/all", myRequestsHandler.GetPendingAllRequests)

		// Report routes
		protected.GET("/reports/dashboard", reportHandler.GetDashboardSummary)

		// Balance routes
		protected.GET("/balances", balanceHandler.GetBalances)
	}
}
