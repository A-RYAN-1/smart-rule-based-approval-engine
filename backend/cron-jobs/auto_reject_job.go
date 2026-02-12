package jobs

import (
	"context"
	"log"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
)

func RunAutoRejectJob(ctx context.Context, service interfaces.AutoRejectService) {
	log.Println("Running Auto Reject Cron Job...")

	if err := service.AutoRejectExpiredRequests(ctx); err != nil {
		log.Printf("Error in Auto Reject Job: %v", err)
	}
}
