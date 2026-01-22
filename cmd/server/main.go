package main

import (
	"rule-based-approval-engine/internal/config"
	"rule-based-approval-engine/internal/database"
	"rule-based-approval-engine/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)

	router := gin.Default()
	routes.Register(router)

	// log.Println("🚀 Server started on port", cfg.AppPort)
	router.Run(":" + cfg.AppPort)
	// hash, _ := bcrypt.GenerateFromPassword([]byte("Lee@123"), bcrypt.DefaultCost)
	// fmt.Println(string(hash))
	// log.Printf(string(hash))
}
