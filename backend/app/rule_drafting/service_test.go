package rule_drafting

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
)

type fakeProvider struct {
	reply string
	err   error
	calls int
	got   struct{ system, user string }
}

func (f *fakeProvider) Complete(_ context.Context, system, user string) (string, error) {
	f.calls++
	f.got.system, f.got.user = system, user
	return f.reply, f.err
}

type lookupStub struct{ rule *models.Rule }

func (l lookupStub) GetRule(context.Context, string, int64) (*models.Rule, error) {
	if l.rule == nil {
		return nil, apperrors.ErrNoRuleFound
	}
	return l.rule, nil
}

func newTestService(t *testing.T, p Provider, existing *models.Rule) *Service {
	t.Helper()
	set, err := LoadEvalSet()
	if err != nil {
		t.Fatal(err)
	}
	byID := map[int64]EvalGrade{}
	var grades []Grade
	for _, g := range set.Grades {
		byID[g.ID] = g
		grades = append(grades, Grade{ID: g.ID, Name: g.Name})
	}
	return NewService(p, limitsValidator{byID}, staticGrades(grades), lookupStub{existing}, 100)
}

func TestDraftBuildsRuleAndDiff(t *testing.T) {
	fp := &fakeProvider{reply: `{"ok":true,"request_type":"EXPENSE","grade_id":2,"value":8000}`}
	existing := &models.Rule{ID: 7, Condition: map[string]interface{}{"max_amount": float64(5000)}}
	d, err := newTestService(t, fp, existing).Draft(context.Background(), "ADMIN", "  expenses up to 8k for grade 2 ")
	if err != nil {
		t.Fatal(err)
	}
	if d.Rule.RequestType != "EXPENSE" || d.Rule.GradeID != 2 || d.Rule.Action != "AUTO_APPROVE" || !d.Rule.Active {
		t.Fatalf("unexpected rule %+v", d.Rule)
	}
	if d.Rule.Condition["max_amount"] != float64(8000) {
		t.Fatalf("condition %v", d.Rule.Condition)
	}
	if d.GradeName != "Grade 2" || d.Change.Field != "max_amount" || d.Change.Before == nil || *d.Change.Before != 5000 || d.Change.After != 8000 {
		t.Fatalf("unexpected diff %+v", d.Change)
	}
	if fp.got.user != "expenses up to 8k for grade 2" {
		t.Fatalf("sentence was not trimmed: %q", fp.got.user)
	}
	if !strings.Contains(fp.got.system, `id 2: "Grade 2"`) {
		t.Fatal("prompt does not list the grades")
	}
}

func TestDraftWithoutExistingRuleHasNoBefore(t *testing.T) {
	fp := &fakeProvider{reply: `{"ok":true,"request_type":"LEAVE","grade_id":1,"value":3}`}
	d, err := newTestService(t, fp, nil).Draft(context.Background(), "ADMIN", "leave 3 days grade 1")
	if err != nil {
		t.Fatal(err)
	}
	if d.Change.Before != nil {
		t.Fatalf("expected no current value, got %+v", d)
	}
}

func TestDraftRejectsBadReplies(t *testing.T) {
	cases := map[string]struct {
		reply string
		want  error
	}{
		"not json":              {`here is your rule`, ErrBadModelReply},
		"extra field":           {`{"ok":true,"request_type":"LEAVE","grade_id":1,"value":3,"active":false}`, ErrBadModelReply},
		"model refused":         {`{"ok":false,"reason":"no grade"}`, ErrUnclear},
		"unknown type":          {`{"ok":true,"request_type":"BONUS","grade_id":1,"value":3}`, ErrBadModelReply},
		"unknown grade":         {`{"ok":true,"request_type":"LEAVE","grade_id":9,"value":3}`, ErrBadModelReply},
		"zero value":            {`{"ok":true,"request_type":"LEAVE","grade_id":1,"value":0}`, ErrBadModelReply},
		"negative value":        {`{"ok":true,"request_type":"EXPENSE","grade_id":1,"value":-5}`, ErrBadModelReply},
		"fractional leave days": {`{"ok":true,"request_type":"LEAVE","grade_id":1,"value":2.5}`, ErrBadModelReply},
		"over the grade limit":  {`{"ok":true,"request_type":"EXPENSE","grade_id":2,"value":90000}`, apperrors.ErrQuotaExceeded},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := newTestService(t, &fakeProvider{reply: tc.reply}, nil).Draft(context.Background(), "ADMIN", "x")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestDraftGuards(t *testing.T) {
	ctx := context.Background()
	fp := &fakeProvider{reply: `{"ok":true,"request_type":"LEAVE","grade_id":1,"value":3}`}
	svc := newTestService(t, fp, nil)

	if _, err := svc.Draft(ctx, "EMPLOYEE", "leave 3"); !errors.Is(err, apperrors.ErrAdminOnly) {
		t.Fatalf("non-admin: %v", err)
	}
	if _, err := svc.Draft(ctx, "ADMIN", "   "); !errors.Is(err, ErrSentenceEmpty) {
		t.Fatalf("empty: %v", err)
	}
	if _, err := svc.Draft(ctx, "ADMIN", strings.Repeat("a", MaxSentenceLength+1)); !errors.Is(err, ErrSentenceLong) {
		t.Fatalf("long: %v", err)
	}
	if fp.calls != 0 {
		t.Fatalf("model was called %d times for invalid input", fp.calls)
	}
	if _, err := newTestService(t, nil, nil).Draft(ctx, "ADMIN", "leave 3"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("disabled: %v", err)
	}
}

func TestProviderErrorIsPassedUp(t *testing.T) {
	boom := errors.New("network down")
	_, err := newTestService(t, &fakeProvider{err: boom}, nil).Draft(context.Background(), "ADMIN", "x")
	if !errors.Is(err, boom) {
		t.Fatalf("got %v", err)
	}
}

func TestDailyLimiter(t *testing.T) {
	now := time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC)
	l := newDailyLimiter(2, func() time.Time { return now })
	if !l.allow() || !l.allow() {
		t.Fatal("first two calls must pass")
	}
	if l.allow() {
		t.Fatal("third call must be refused")
	}
	now = now.Add(24 * time.Hour)
	if !l.allow() {
		t.Fatal("a new UTC day resets the count")
	}
}

func TestDailyLimitStopsModelCalls(t *testing.T) {
	fp := &fakeProvider{reply: `{"ok":true,"request_type":"LEAVE","grade_id":1,"value":3}`}
	svc := newTestService(t, fp, nil)
	svc.limiter = newDailyLimiter(1, time.Now)
	if _, err := svc.Draft(context.Background(), "ADMIN", "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Draft(context.Background(), "ADMIN", "b"); !errors.Is(err, ErrDailyLimit) {
		t.Fatalf("got %v", err)
	}
	if fp.calls != 1 {
		t.Fatalf("model calls %d, want 1", fp.calls)
	}
}

func TestGeminiClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "k" || !strings.HasSuffix(r.URL.Path, "/m1:generateContent") {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.URL.RawQuery != "" {
			http.Error(w, "the key must not travel in the URL", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"ok\":false}"}]}}]}`))
	}))
	defer srv.Close()

	g := NewGemini("k", "m1")
	g.BaseURL = srv.URL + "/"
	got, err := g.Complete(context.Background(), "sys", "hi")
	if err != nil || got != `{"ok":false}` {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestGeminiQuotaAndFailure(t *testing.T) {
	for status, check := range map[int]func(error) bool{
		http.StatusTooManyRequests:     func(e error) bool { return errors.Is(e, ErrProviderQuota) },
		http.StatusInternalServerError: func(e error) bool { return e != nil && !errors.Is(e, ErrProviderQuota) },
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
		g := NewGemini("k", "m")
		g.BaseURL = srv.URL + "/"
		_, err := g.Complete(context.Background(), "s", "u")
		srv.Close()
		if !check(err) {
			t.Fatalf("status %d gave %v", status, err)
		}
	}
}

func TestEvalSetIsSound(t *testing.T) {
	set, err := LoadEvalSet()
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Cases) != 50 {
		t.Fatalf("%d cases, want 50", len(set.Cases))
	}
	if set.PromptVersion != PromptVersion {
		t.Fatalf("eval set is for prompt %s, code is %s", set.PromptVersion, PromptVersion)
	}
	seen := map[string]bool{}
	for _, c := range set.Cases {
		if seen[c.Sentence] {
			t.Errorf("duplicate sentence %q", c.Sentence)
		}
		seen[c.Sentence] = true
	}
}

// oracle answers every case with its expected outcome. Scoring it 50/50 proves
// the scorer and the validator agree with the dataset, so a real model's
// misses are the model's and not the harness's.
type oracle struct{ set *EvalSet }

func (o oracle) Complete(_ context.Context, _, user string) (string, error) {
	for _, c := range o.set.Cases {
		if c.Sentence != user {
			continue
		}
		if c.Expect.Reject {
			return `{"ok":false,"reason":"x"}`, nil
		}
		return `{"ok":true,"request_type":"` + c.Expect.RequestType +
			`","grade_id":` + strconv.FormatInt(c.Expect.GradeID, 10) +
			`,"value":` + strconv.FormatFloat(c.Expect.Value, 'f', -1, 64) + `}`, nil
	}
	return `{"ok":false}`, nil
}

func TestEvaluateScoresAnOracleFully(t *testing.T) {
	set, _ := LoadEvalSet()
	rep, err := Evaluate(context.Background(), oracle{set}, set)
	if err != nil {
		t.Fatal(err)
	}
	if rep.ExactMatch != rep.Total {
		for _, r := range rep.Results {
			if !r.Pass {
				t.Logf("%q want %s got %s", r.Sentence, r.Want, r.Got)
			}
		}
		t.Fatalf("oracle scored %d/%d", rep.ExactMatch, rep.Total)
	}
}

// A model that always refuses passes every reject case and fails every rule case.
func TestEvaluateCountsMisses(t *testing.T) {
	set, _ := LoadEvalSet()
	rep, err := Evaluate(context.Background(), &fakeProvider{reply: `{"ok":false}`}, set)
	if err != nil {
		t.Fatal(err)
	}
	if rep.RulesMatched != 0 || rep.RejectsKept != rep.Rejects || rep.ExactMatch != rep.Rejects {
		t.Fatalf("unexpected report %+v", *rep)
	}
}

// A provider failure must stop the run, never be scored as a refusal.
func TestEvaluateStopsOnProviderFailure(t *testing.T) {
	set, _ := LoadEvalSet()
	rep, err := Evaluate(context.Background(), &fakeProvider{err: errors.New("status 404")}, set)
	if err == nil {
		t.Fatal("expected the run to stop with an error")
	}
	if rep.ExactMatch != 0 || rep.RejectsKept != 0 {
		t.Fatalf("a failed call was scored: %+v", *rep)
	}
}
