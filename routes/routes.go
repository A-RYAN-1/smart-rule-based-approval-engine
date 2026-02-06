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
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
		// Leave routes
		protected.POST("/leave/apply", leaveHandler.ApplyLeave)
		protected.POST("/leave/cancel/:id", leaveHandler.CancelLeave)
		protected.GET("/leave/pending", leaveApprovalHandler.GetPendingLeaves)
		protected.POST("/leave/approve/:id", leaveApprovalHandler.ApproveLeave)
		protected.POST("/leave/reject/:id", leaveApprovalHandler.RejectLeave)

		// Expense routes
		protected.POST("/expense/apply", expenseHandler.ApplyExpense)
		protected.POST("/expense/cancel/:id", expenseHandler.CancelExpense)
		protected.GET("/expense/pending", expenseApprovalHandler.GetPendingExpenses)
		protected.POST("/expense/approve/:id", expenseApprovalHandler.ApproveExpense)
		protected.POST("/expense/reject/:id", expenseApprovalHandler.RejectExpense)

		// Rule routes
		protected.POST("/rules", ruleHandler.CreateRule)
		protected.GET("/rules", ruleHandler.GetRules)
		protected.PUT("/rules/:id", ruleHandler.UpdateRule)
		protected.DELETE("/rules/:id", ruleHandler.DeleteRule)

		// My Requests routes
		protected.GET("/my-requests", myRequestsHandler.GetMyRequests)
		protected.GET("/my-requests/all", myRequestsHandler.GetMyAllRequests)

		// Holiday routes
		protected.POST("/holidays", holidayHandler.AddHoliday)
		protected.GET("/holidays", holidayHandler.GetHolidays)
		protected.DELETE("/holidays/:id", holidayHandler.DeleteHoliday)

		// Report routes
		protected.GET("/reports/dashboard", reportHandler.GetDashboardSummary)

		// Balance routes
		protected.GET("/balances", balanceHandler.GetBalances)

		// Discount routes
		protected.POST("/discount/apply", discountHandler.ApplyDiscount)
		protected.POST("/discount/cancel/:id", discountHandler.CancelDiscount)
		protected.GET("/discount/pending", discountApprovalHandler.GetPendingDiscounts)
		protected.POST("/discount/approve/:id", discountApprovalHandler.ApproveDiscount)
		protected.POST("/discount/reject/:id", discountApprovalHandler.RejectDiscount)
	}
}
