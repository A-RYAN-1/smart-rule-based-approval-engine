// Package handler serves the Go API as a Vercel Function. vercel.json routes
// /api/* and /health here; the app is built once per instance and reused.
package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/ankita-advitot/rule_based_approval_engine/config"
	"github.com/ankita-advitot/rule_based_approval_engine/server"
)

var (
	once    sync.Once
	app     *server.App
	initErr error
)

func build() {
	cfg, err := config.FromEnv(os.Getenv)
	if err != nil {
		initErr = err
		return
	}
	// Vercel Cron calls /api/cron/auto-reject; no in-process scheduler here.
	cfg.RunCron = false
	app, initErr = server.New(context.Background(), cfg)
}

// Handler is the Vercel Function entrypoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(build)
	if initErr != nil {
		log.Printf("startup failed: %v", initErr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"success":false,"message":"service is starting or misconfigured, please try again","data":null,"error":null}`))
		return
	}
	app.Router.ServeHTTP(w, r)
}
