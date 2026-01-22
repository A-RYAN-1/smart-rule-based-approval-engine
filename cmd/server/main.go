package main

import (
	"log"
	"rule-based-approval-engine/internal/config"
	"rule-based-approval-engine/internal/database"
	"rule-based-approval-engine/internal/jobs"
	"rule-based-approval-engine/internal/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)

	router := gin.Default()
	routes.Register(router)
	// ✅ CORS CONFIGURATION
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// ⏱️ CRON SETUP
	loc, _ := time.LoadLocation("Asia/Kolkata")
	c := cron.New(cron.WithLocation(loc))

	// Run daily at 12:00 AM
	c.AddFunc("0 0 * * *", jobs.RunAutoRejectJob)

	c.Start()

	log.Println("🕛 Auto-reject cron scheduled for 12:00 AM daily")

	// log.Println("🚀 Server started on port", cfg.AppPort)
	router.Run(":" + cfg.AppPort)
	// hash, _ := bcrypt.GenerateFromPassword([]byte("Lee@123"), bcrypt.DefaultCost)
	// fmt.Println(string(hash))
	// log.Printf(string(hash))
}
