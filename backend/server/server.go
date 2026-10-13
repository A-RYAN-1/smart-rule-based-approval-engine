// Package server builds the whole application (database, migrations, seed
// accounts, services, routes) from a Config. cmd/server runs it as a
// long-running process; api/index.go serves it as a Vercel function.
package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/app/auth"
	"github.com/ankita-advitot/rule_based_approval_engine/app/auto_reject"
	"github.com/ankita-advitot/rule_based_approval_engine/app/domain_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/expense_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/holidays"
	"github.com/ankita-advitot/rule_based_approval_engine/app/leave_service"
	"github.com/ankita-advitot/rule_based_approval_engine/app/my_requests"
	"github.com/ankita-advitot/rule_based_approval_engine/app/reports"
	"github.com/ankita-advitot/rule_based_approval_engine/app/rules"
	"github.com/ankita-advitot/rule_based_approval_engine/config"
	jobs "github.com/ankita-advitot/rule_based_approval_engine/cron-jobs"
	"github.com/ankita-advitot/rule_based_approval_engine/database"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
	"github.com/ankita-advitot/rule_based_approval_engine/repositories"
	"github.com/ankita-advitot/rule_based_approval_engine/repositories/migrations"
	"github.com/ankita-advitot/rule_based_approval_engine/routes"
	"github.com/ankita-advitot/rule_based_approval_engine/seed"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App is the built application.
type App struct {
	Router     *gin.Engine
	AutoReject interfaces.AutoRejectService
	Pool       *pgxpool.Pool
}

// New connects to the database, applies migrations, seeds the admin and
// manager accounts and registers every route.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	utils.SetJWTSecret(cfg.JWTSecret)
	utils.SetCookieSecure(cfg.CookieSecure)
	gin.SetMode(cfg.GinMode)

	pool, err := database.Connect(ctx, cfg)
	if err != nil {
		return nil, err
	}
	db := database.DB

	if err := migrations.Up(ctx, db); err != nil {
		pool.Close()
		return nil, err
	}

	// Repositories
	userRepo := repositories.NewUserRepository(ctx, db)
	balanceRepo := repositories.NewBalanceRepository(ctx, db)
	ruleRepo := repositories.NewRuleRepository(ctx, db)
	leaveRepo := repositories.NewLeaveRequestRepository(ctx, db)
	expenseRepo := repositories.NewExpenseRequestRepository(ctx, db)
	discountRepo := repositories.NewDiscountRequestRepository(ctx, db)
	holidayRepo := repositories.NewHolidayRepository(ctx, db)
	reportRepo := repositories.NewReportRepository(ctx, db)
	gradeRepo := repositories.NewGradeRepository(ctx, db)
	myRequestsRepo := repositories.NewAggregatedRepository(ctx, db)

	if err := seed.Accounts(ctx, db, balanceRepo, cfg); err != nil {
		pool.Close()
		return nil, err
	}

	// Services
	authService := auth.NewAuthService(ctx, userRepo, balanceRepo, db)
	ruleService := rules.NewRuleService(ctx, ruleRepo, gradeRepo, db)
	leaveService := leave_service.NewLeaveService(ctx, leaveRepo, balanceRepo, ruleService, userRepo, db)
	leaveApprovalService := leave_service.NewLeaveApprovalService(ctx, leaveRepo, balanceRepo, userRepo, db)
	expenseService := expense_service.NewExpenseService(ctx, expenseRepo, balanceRepo, ruleService, userRepo, db)
	expenseApprovalService := expense_service.NewExpenseApprovalService(ctx, expenseRepo, balanceRepo, userRepo, db)
	holidayService := holidays.NewHolidayService(ctx, holidayRepo)
	reportService := reports.NewReportService(ctx, reportRepo)
	balanceService := domain_service.NewBalanceService(ctx, balanceRepo, db)
	discountService := domain_service.NewDiscountService(ctx, discountRepo, balanceRepo, ruleService, userRepo, db)
	discountApprovalService := domain_service.NewDiscountApprovalService(ctx, discountRepo, balanceRepo, userRepo, db)
	autoRejectService := auto_reject.NewAutoRejectService(ctx, leaveRepo, expenseRepo, discountRepo, holidayRepo, db)
	myRequestsService := my_requests.NewMyRequestsService(ctx, myRequestsRepo)

	router := gin.New()
	router.Use(requestID(), requestLog(), gin.CustomRecovery(recovered))
	if len(cfg.CORSOrigins) > 0 {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.CORSOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	// Health also pings the database, so the uptime check notices when the
	// database is unreachable. Served at /health and /api/health.
	health := func(c *gin.Context) {
		status, state, dbState := http.StatusOK, "healthy", "ok"
		pingCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := pool.Ping(pingCtx); err != nil {
			status, state, dbState = http.StatusServiceUnavailable, "degraded", "unreachable"
			log.Printf("request_id=%s health: database ping failed: %v", c.GetString("request_id"), err)
		}
		c.JSON(status, gin.H{
			"status":    state,
			"database":  dbState,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "approval-engine",
		})
	}
	router.GET("/health", health)
	router.GET("/api/health", health)

	// Scheduled trigger for the auto-reject job. Vercel Cron sends
	// "Authorization: Bearer <CRON_SECRET>". Disabled without CRON_SECRET.
	router.GET("/api/cron/auto-reject", func(c *gin.Context) {
		want := []byte("Bearer " + cfg.CronSecret)
		got := []byte(c.GetHeader("Authorization"))
		if cfg.CronSecret == "" || subtle.ConstantTimeCompare(got, want) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		jobs.RunAutoRejectJob(c.Request.Context(), autoRejectService)
		if cfg.DemoMode {
			if err := seed.ResetDemoRules(c.Request.Context(), db); err != nil {
				log.Printf("request_id=%s demo rule reset failed: %v", c.GetString("request_id"), err)
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "done"})
	})

	// Demo accounts for the login page. 404 unless DEMO_MODE is on.
	router.GET("/api/demo-logins", func(c *gin.Context) {
		if !cfg.DemoMode {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"logins": seed.DemoLogins(cfg)})
	})

	routes.Register(ctx, router,
		authService, leaveService, expenseService, leaveApprovalService, expenseApprovalService,
		ruleService, myRequestsService, holidayService, reportService, balanceService,
		autoRejectService, discountService, discountApprovalService,
	)

	return &App{Router: router, AutoReject: autoRejectService, Pool: pool}, nil
}

// requestID tags each request with an id (from X-Request-Id, or a new one),
// returns it in the response header and puts it in every log line.
func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set("request_id", id)
		c.Header("X-Request-Id", id)
		c.Next()
	}
}

// requestLog writes one line per request. No query strings, bodies or
// cookies, so no personal data ends up in the logs.
func requestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("request_id=%s method=%s path=%s status=%d duration=%s",
			c.GetString("request_id"), c.Request.Method, c.Request.URL.Path,
			c.Writer.Status(), time.Since(start).Round(time.Millisecond))
	}
}

func recovered(c *gin.Context, err any) {
	log.Printf("request_id=%s panic: %v", c.GetString("request_id"), err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"message": "something went wrong, please try again",
		"data":    nil,
		"error":   nil,
	})
}
