package routes

import (
	"context"
	"testing"

	apidocs "github.com/ankita-advitot/rule_based_approval_engine/api"
	"github.com/ankita-advitot/rule_based_approval_engine/api/speccheck"
	"github.com/gin-gonic/gin"
)

// serverLevel are described in the spec but registered by server.New, not by
// Register. The integration test in package server covers them.
var serverLevel = []string{
	"GET /health",
	"GET /api/health",
	"GET /api/cron/auto-reject",
	"GET /api/demo-logins",
}

// TestRoutesMatchSpec fails when a route is added, renamed or removed without
// api/swagger.yaml following, in either direction.
func TestRoutesMatchSpec(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	Register(context.Background(), router, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	var served []string
	for _, r := range router.Routes() {
		served = append(served, speccheck.FromGin(r.Method, r.Path))
	}
	described, err := speccheck.Operations(apidocs.Spec)
	if err != nil {
		t.Fatal(err)
	}
	undocumented, unserved := speccheck.Diff(described, served, serverLevel)
	for _, op := range undocumented {
		t.Errorf("served but not in api/swagger.yaml: %s", op)
	}
	for _, op := range unserved {
		t.Errorf("in api/swagger.yaml but not served: %s", op)
	}
}
