package main

import (
	"log"
	"time"

	"rule-based-approval-engine/internal/app/repositories"
	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/config"
	jobs "rule-based-approval-engine/internal/cron-jobs"
	"rule-based-approval-engine/internal/database"
	"rule-based-approval-engine/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)

	// Initialize Repositories
	userRepo := repositories.NewUserRepository(database.DB)
	leaveReqRepo := repositories.NewLeaveRequestRepository(database.DB)
	expenseReqRepo := repositories.NewExpenseRequestRepository(database.DB)
	ruleRepo := repositories.NewRuleRepository(database.DB)
	balanceRepo := repositories.NewBalanceRepository(database.DB)
	myRequestsRepo := repositories.NewMyRequestsRepository(database.DB)
	holidayRepo := repositories.NewHolidayRepository(database.DB)
	reportRepo := repositories.NewReportRepository(database.DB)

	// Initialize Services
	ruleService := services.NewRuleService(ruleRepo)
	authService := services.NewAuthService(userRepo, balanceRepo, database.DB)
	leaveService := services.NewLeaveService(leaveReqRepo, balanceRepo, ruleService, userRepo, database.DB)
	expenseService := services.NewExpenseService(expenseReqRepo, balanceRepo, ruleService, userRepo, database.DB)
	leaveApprovalService := services.NewLeaveApprovalService(leaveReqRepo, balanceRepo, userRepo, database.DB)
	expenseApprovalService := services.NewExpenseApprovalService(expenseReqRepo, balanceRepo, userRepo, database.DB)
	myRequestsService := services.NewMyRequestsService(myRequestsRepo)
	holidayService := services.NewHolidayService(holidayRepo)
	reportService := services.NewReportService(reportRepo)
	balanceService := services.NewBalanceService(balanceRepo, database.DB)
	autoRejectService := services.NewAutoRejectService(leaveReqRepo, expenseReqRepo, holidayRepo, database.DB)
	discountService := services.NewDiscountService(ruleService, userRepo, database.DB)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register routes with services
	routes.Register(
		router,
		authService,
		leaveService,
		expenseService,
		leaveApprovalService,
		expenseApprovalService,
		ruleService,
		myRequestsService,
		holidayService,
		reportService,
		balanceService,
		autoRejectService,
		discountService,
	)

	//CRON SETUP
	loc, _ := time.LoadLocation("Asia/Kolkata")
	c := cron.New(cron.WithLocation(loc))
	c.AddFunc("0 0 * * *", func() { jobs.RunAutoRejectJob(autoRejectService) })
	// c.AddFunc("every @1m", func() { jobs.RunAutoRejectJob(autoRejectService) })

	c.Start()

	log.Println("🚀 Server started on port", cfg.AppPort)
	router.Run(":" + cfg.AppPort)
}
