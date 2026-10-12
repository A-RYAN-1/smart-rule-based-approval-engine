package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/config"
	jobs "github.com/ankita-advitot/rule_based_approval_engine/cron-jobs"
	"github.com/ankita-advitot/rule_based_approval_engine/database"
	"github.com/ankita-advitot/rule_based_approval_engine/repositories/migrations"
	"github.com/ankita-advitot/rule_based_approval_engine/server"

	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// `server migrate up` or `server migrate create <name>` manages
	// migrations and exits.
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		pool, err := database.Connect(ctx, cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()
		if err := migrations.HandleMigrateCommand(ctx, database.DB, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	app, err := server.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Pool.Close()

	// Nightly auto-reject of expired requests (midnight, India time).
	if cfg.RunCron {
		loc, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			loc = time.FixedZone("IST", 5*60*60+30*60)
		}
		c := cron.New(cron.WithLocation(loc))
		if _, err := c.AddFunc("0 0 * * *", func() { jobs.RunAutoRejectJob(ctx, app.AutoReject) }); err != nil {
			log.Fatal(err)
		}
		c.Start()
		defer c.Stop()
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Println("Server started on port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
