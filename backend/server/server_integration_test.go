package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/config"
)

// TestMainFlow runs against a real, empty PostgreSQL database named by
// TEST_DATABASE_URL (CI starts one; locally see the README). It is skipped
// when the variable is not set.
func TestMainFlow(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	cfg, err := config.FromEnv(func(k string) string {
		return map[string]string{
			"DATABASE_URL":           dbURL,
			"JWT_SECRET":             "integration-test-secret-0123456789abcdef",
			"GIN_MODE":               "test",
			"COOKIE_SECURE":          "false",
			"CRON_SECRET":            "cron-test-secret",
			"SEED_ADMIN_PASSWORD":    "Admin@12345",
			"SEED_MANAGER_PASSWORD":  "Manager@12345",
			"SEED_EMPLOYEE_PASSWORD": "Employee@12345",
			"DEMO_MODE":              "true",
		}[k]
	})
	if err != nil {
		t.Fatal(err)
	}

	app, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Pool.Close)

	// Building a second time must be a no-op for migrations and seeding.
	again, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("second startup: %v", err)
	}
	again.Pool.Close()

	call := func(method, path, token string, body any) (int, map[string]any) {
		t.Helper()
		var buf bytes.Buffer
		if body != nil {
			_ = json.NewEncoder(&buf).Encode(body)
		}
		req := httptest.NewRequest(method, path, &buf)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		app.Router.ServeHTTP(rec, req)
		var out map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}
	login := func(email, password string) string {
		t.Helper()
		code, out := call("POST", "/api/auth/login", "", map[string]string{"email": email, "password": password})
		if code != http.StatusOK {
			t.Fatalf("login %s: %d %v", email, code, out)
		}
		return out["data"].(map[string]any)["token"].(string)
	}
	status := func(out map[string]any) string {
		data, _ := out["data"].(map[string]any)
		return fmt.Sprint(data["status"])
	}

	if code, out := call("GET", "/api/health", "", nil); code != http.StatusOK || out["database"] != "ok" {
		t.Fatalf("health: %d %v", code, out)
	}

	email := "employee@example.com"
	if code, out := call("POST", "/api/auth/register", "", map[string]string{
		"name": "Employee", "email": email, "password": "Employee@123",
	}); code != http.StatusCreated {
		t.Fatalf("register: %d %v", code, out)
	}
	if code, _ := call("POST", "/api/auth/register", "", map[string]string{
		"name": "Mallory", "email": "admin@company.com", "password": "Mallory@123",
	}); code != http.StatusConflict {
		t.Fatalf("sign-up with the admin email must be refused, got %d", code)
	}

	employee := login(email, "Employee@123")
	manager := login("manager@company.com", "Manager@12345")

	// Grade 1 rule: up to 3 days is approved by the system.
	if _, out := call("POST", "/api/leaves/request", employee, map[string]string{
		"from_date": "2031-03-03", "to_date": "2031-03-04", "leave_type": "CASUAL", "reason": "family",
	}); status(out) != "AUTO_APPROVED" {
		t.Fatalf("short leave should be auto-approved: %v", out)
	}
	// Longer leave waits for the manager.
	if _, out := call("POST", "/api/leaves/request", employee, map[string]string{
		"from_date": "2031-04-07", "to_date": "2031-04-18", "leave_type": "CASUAL", "reason": "trip",
	}); status(out) != "PENDING" {
		t.Fatalf("long leave should be pending: %v", out)
	}

	code, out := call("GET", "/api/leaves/pending", manager, nil)
	if code != http.StatusOK {
		t.Fatalf("pending: %d %v", code, out)
	}
	requests := out["data"].(map[string]any)["requests"].([]any)
	if len(requests) != 1 {
		t.Fatalf("manager should see 1 pending leave, got %d", len(requests))
	}
	id := int64(requests[0].(map[string]any)["id"].(float64))
	if code, out := call("POST", fmt.Sprintf("/api/leaves/%d/approve", id), manager, map[string]string{"comment": "ok"}); code != http.StatusOK {
		t.Fatalf("approve: %d %v", code, out)
	}

	if code, _ := call("GET", "/api/admin/rules", employee, nil); code == http.StatusOK {
		t.Fatal("an employee must not read the admin rules")
	}
	if code, _ := call("GET", "/api/me", "", nil); code != http.StatusUnauthorized {
		t.Fatalf("no token should be 401, got %d", code)
	}
	if code, _ := call("GET", "/api/cron/auto-reject", "", nil); code != http.StatusUnauthorized {
		t.Fatalf("cron without the secret should be 401, got %d", code)
	}
	if code, _ := call("GET", "/api/cron/auto-reject", "cron-test-secret", nil); code != http.StatusOK {
		t.Fatalf("cron with the secret should run, got %d", code)
	}

	// Demo mode: the three seed logins are published and work, and the
	// cron run undoes rule changes made with the public admin login.
	code, out = call("GET", "/api/demo-logins", "", nil)
	if code != http.StatusOK || len(out["logins"].([]any)) != 3 {
		t.Fatalf("demo logins: %d %v", code, out)
	}
	demoEmployee := login("employee@company.com", "Employee@12345")
	admin := login("admin@company.com", "Admin@12345")
	if code, _ := call("GET", "/api/me", demoEmployee, nil); code != http.StatusOK {
		t.Fatalf("demo employee /me: %d", code)
	}
	code, out = call("GET", "/api/admin/rules", admin, nil)
	if code != http.StatusOK {
		t.Fatalf("admin rules: %d %v", code, out)
	}
	countRules := func() int {
		var n int
		if err := app.Pool.QueryRow(context.Background(), "SELECT count(*) FROM rules").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	before := countRules()
	if _, err := app.Pool.Exec(context.Background(), "DELETE FROM rules"); err != nil {
		t.Fatal(err)
	}
	if code, _ := call("GET", "/api/cron/auto-reject", "cron-test-secret", nil); code != http.StatusOK {
		t.Fatalf("cron: %d", code)
	}
	if after := countRules(); after != before || after != 9 {
		t.Fatalf("rules after demo reset = %d, want %d (9 defaults)", after, before)
	}
}
