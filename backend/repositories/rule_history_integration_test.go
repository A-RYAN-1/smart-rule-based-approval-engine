package repositories

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// poolDB lets a pgx pool stand in for interfaces.DB; the history reader only calls Query.
type poolDB struct{ *pgxpool.Pool }

func (poolDB) Begin(context.Context) (interfaces.Tx, error) { panic("not used") }
func (p poolDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.Pool.QueryRow(ctx, sql, args...)
}
func (p poolDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.Pool.Query(ctx, sql, args...)
}
func (p poolDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.Pool.Exec(ctx, sql, args...)
}

// Runs against a real Postgres with the migrations applied:
//
//	RULE_HISTORY_TEST_DATABASE_URL=postgres://... go test ./repositories -run RuleHistory
//
// It is skipped when the variable is not set, so the ordinary unit run needs no database.
func TestRuleHistoryAgainstPostgres(t *testing.T) {
	url := os.Getenv("RULE_HISTORY_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("RULE_HISTORY_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	var userA, userB int64
	insertUser := func(email string, grade int64) int64 {
		var id int64
		err := pool.QueryRow(ctx,
			`INSERT INTO users (name, email, password_hash, grade_id, role) VALUES ($1, $2, 'x', $3, 'EMPLOYEE') RETURNING id`,
			"history test", email, grade).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	userA = insertUser("history-a@example.com", 1)
	userB = insertUser("history-b@example.com", 2)
	t.Cleanup(func() {
		for _, table := range []string{"leave_requests", "expense_requests", "discount_requests"} {
			_, _ = pool.Exec(ctx, "DELETE FROM "+table+" WHERE employee_id = ANY($1)", []int64{userA, userB})
		}
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = ANY($1)", []int64{userA, userB})
	})

	exec := func(sql string, args ...any) {
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	// Grade 1: two leaves counted, one cancelled. Grade 2: one leave that must not appear.
	leave := `INSERT INTO leave_requests (employee_id, from_date, to_date, reason, leave_type, status, days, created_at)
		VALUES ($1, '2026-01-05', '2026-01-07', 'r', 'CASUAL', $2, $3, $4)`
	exec(leave, userA, "PENDING", 3, "2026-01-01 10:00")
	exec(leave, userA, "PENDING", 5, "2026-01-02 10:00")
	exec(leave, userA, "CANCELLED", 9, "2026-01-03 10:00")
	exec(leave, userB, "PENDING", 7, "2026-01-04 10:00")
	expense := `INSERT INTO expense_requests (employee_id, amount, category, reason, status, created_at)
		VALUES ($1, $2, 'travel', 'r', $3, $4)`
	exec(expense, userA, 1200.50, "PENDING", "2026-01-01 10:00")
	exec(expense, userA, 99999, "CANCELLED", "2026-01-02 10:00")
	discount := `INSERT INTO discount_requests (employee_id, discount_percentage, reason, status, created_at)
		VALUES ($1, $2, 'r', $3, $4)`
	exec(discount, userA, 7.5, "PENDING", "2026-01-01 10:00")

	h := NewRuleHistory(poolDB{pool})
	cases := []struct {
		kind  string
		grade int64
		limit int
		want  []float64
	}{
		{"LEAVE", 1, 10, []float64{5, 3}}, // newest first, cancelled and other grade left out
		{"LEAVE", 1, 1, []float64{5}},     // the limit applies
		{"EXPENSE", 1, 10, []float64{1200.5}},
		{"DISCOUNT", 1, 10, []float64{7.5}},
		{"LEAVE", 2, 10, []float64{7}},
	}
	for _, tc := range cases {
		got, err := h.RecentValues(ctx, tc.kind, tc.grade, tc.limit)
		if err != nil {
			t.Fatalf("%s grade %d: %v", tc.kind, tc.grade, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s grade %d limit %d: got %v, want %v", tc.kind, tc.grade, tc.limit, got, tc.want)
		}
	}
	if _, err := h.RecentValues(ctx, "BOGUS", 1, 10); err == nil {
		t.Error("an unknown request type must be refused")
	}
}
