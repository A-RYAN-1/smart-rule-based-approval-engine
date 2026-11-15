package rule_dryrun

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"

	"github.com/gin-gonic/gin"
)

type validatorStub struct {
	err   error
	calls int
}

func (v *validatorStub) ValidateRule(context.Context, models.Rule) error {
	v.calls++
	return v.err
}

type lookupStub struct {
	rule *models.Rule
	err  error
}

func (l lookupStub) GetRule(context.Context, string, int64) (*models.Rule, error) {
	return l.rule, l.err
}

type historyStub struct {
	values []float64
	err    error
	asked  struct {
		requestType string
		gradeID     int64
		limit       int
	}
}

func (h *historyStub) RecentValues(_ context.Context, requestType string, gradeID int64, limit int) ([]float64, error) {
	h.asked.requestType, h.asked.gradeID, h.asked.limit = requestType, gradeID, limit
	return h.values, h.err
}

func expenseRule(limit float64) map[string]interface{} {
	return map[string]interface{}{"max_amount": limit}
}

func newService(saved *models.Rule, lookupErr error, history *historyStub) (*Service, *validatorStub) {
	v := &validatorStub{}
	return NewService(v, lookupStub{rule: saved, err: lookupErr}, history), v
}

func TestRunDecidesSamplesAndCountsHistory(t *testing.T) {
	saved := &models.Rule{Condition: expenseRule(5000)}
	history := &historyStub{values: []float64{1000, 4000, 6000, 8000, 12000}}
	svc, v := newService(saved, nil, history)

	res, err := svc.Run(context.Background(), "ADMIN", Request{
		RequestType: "EXPENSE", GradeID: 1, Condition: expenseRule(8000), Values: []float64{5000, 7000, 9000},
	})
	if err != nil {
		t.Fatal(err)
	}
	if v.calls != 1 {
		t.Fatalf("draft validated %d times, want 1", v.calls)
	}
	want := []Row{
		{5000, SourceSample, AutoApproved, AutoApproved},
		{7000, SourceSample, NeedsApproval, AutoApproved},
		{9000, SourceSample, NeedsApproval, NeedsApproval},
	}
	if len(res.Rows) != len(want) {
		t.Fatalf("rows %v", res.Rows)
	}
	for i, w := range want {
		if res.Rows[i] != w {
			t.Errorf("row %d: got %+v, want %+v", i, res.Rows[i], w)
		}
	}
	// 5 past requests: saved rule approves 1000 and 4000; the draft also approves 6000 and 8000.
	got := res.Summary
	if got != (Summary{HistoryTotal: 5, CurrentAuto: 2, DraftAuto: 4, NewlyAuto: 2, NoLongerAuto: 0}) {
		t.Fatalf("summary %+v", got)
	}
	if history.asked.requestType != "EXPENSE" || history.asked.gradeID != 1 || history.asked.limit != HistorySize {
		t.Fatalf("history asked for %+v", history.asked)
	}
}

func TestRunTightenedRuleShowsRequestsThatNoLongerAutoApprove(t *testing.T) {
	saved := &models.Rule{Condition: expenseRule(8000)}
	svc, _ := newService(saved, nil, &historyStub{values: []float64{1000, 4000, 8000}})
	res, err := svc.Run(context.Background(), "ADMIN", Request{
		RequestType: "EXPENSE", GradeID: 1, Condition: expenseRule(3000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.NoLongerAuto != 2 || res.Summary.NewlyAuto != 0 || res.Summary.CurrentAuto != 3 || res.Summary.DraftAuto != 1 {
		t.Fatalf("summary %+v", res.Summary)
	}
	if len(res.Rows) != 0 {
		t.Fatalf("history is summarised, not listed: %v", res.Rows)
	}
}

func TestRunWithNoSavedRuleSendsEverythingToApproval(t *testing.T) {
	svc, _ := newService(nil, apperrors.ErrNoRuleFound, &historyStub{values: []float64{100}})
	res, err := svc.Run(context.Background(), "ADMIN", Request{
		RequestType: "EXPENSE", GradeID: 1, Condition: expenseRule(500), Values: []float64{100},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.CurrentCondition != nil {
		t.Fatalf("current condition %v, want none", res.CurrentCondition)
	}
	if res.Rows[0].Current != NeedsApproval || res.Rows[0].Draft != AutoApproved {
		t.Fatalf("row %+v", res.Rows[0])
	}
	if res.Summary.NewlyAuto != 1 {
		t.Fatalf("summary %+v", res.Summary)
	}
}

func TestRunRefusals(t *testing.T) {
	good := Request{RequestType: "EXPENSE", GradeID: 1, Condition: expenseRule(500)}
	tooMany := good
	tooMany.Values = make([]float64, MaxSamples+1)
	for i := range tooMany.Values {
		tooMany.Values[i] = 1
	}
	zero := good
	zero.Values = []float64{0}
	negative := good
	negative.Values = []float64{-3}

	cases := []struct {
		name    string
		role    string
		req     Request
		valErr  error
		history *historyStub
		want    error
	}{
		{"employee", "EMPLOYEE", good, nil, &historyStub{}, apperrors.ErrAdminOnly},
		{"too many samples", "ADMIN", tooMany, nil, &historyStub{}, ErrTooManySamples},
		{"zero sample", "ADMIN", zero, nil, &historyStub{}, ErrBadSample},
		{"negative sample", "ADMIN", negative, nil, &historyStub{}, ErrBadSample},
		{"over the grade limit", "ADMIN", good, apperrors.ErrQuotaExceeded, &historyStub{}, apperrors.ErrQuotaExceeded},
		{"history unavailable", "ADMIN", good, nil, &historyStub{err: apperrors.ErrDatabase}, apperrors.ErrDatabase},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, v := newService(nil, apperrors.ErrNoRuleFound, tc.history)
			v.err = tc.valErr
			if _, err := svc.Run(context.Background(), tc.role, tc.req); !errors.Is(err, tc.want) {
				t.Fatalf("error %v, want %v", err, tc.want)
			}
		})
	}
}

func TestRunNeverWrites(t *testing.T) {
	// The service holds a validator, a lookup and a history reader. None has a write method,
	// so a dry run cannot save; this checks the lookup error path does not hide a real failure.
	svc, _ := newService(nil, apperrors.ErrDatabase, &historyStub{})
	_, err := svc.Run(context.Background(), "ADMIN", Request{RequestType: "EXPENSE", GradeID: 1, Condition: expenseRule(1)})
	if !errors.Is(err, apperrors.ErrDatabase) {
		t.Fatalf("a failing lookup must surface, got %v", err)
	}
}

func dryRunRouter(role string, svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("role", role) })
	r.POST("/dry-run", NewHandler(svc).DryRun)
	return r
}

func post(r *gin.Engine, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dry-run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestHandlerStatusCodes(t *testing.T) {
	body := `{"request_type":"EXPENSE","grade_id":1,"condition":{"max_amount":8000},"values":[7000]}`
	cases := []struct {
		name   string
		role   string
		valErr error
		body   string
		want   int
	}{
		{"admin", "ADMIN", nil, body, http.StatusOK},
		{"employee", "EMPLOYEE", nil, body, http.StatusForbidden},
		{"bad json", "ADMIN", nil, `nope`, http.StatusBadRequest},
		{"zero sample", "ADMIN", nil, `{"request_type":"EXPENSE","grade_id":1,"condition":{"max_amount":1},"values":[0]}`, http.StatusBadRequest},
		{"missing condition", "ADMIN", apperrors.ErrConditionRequired, `{"request_type":"EXPENSE","grade_id":1}`, http.StatusBadRequest},
		{"over limit", "ADMIN", apperrors.ErrQuotaExceeded, body, http.StatusUnprocessableEntity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, v := newService(nil, apperrors.ErrNoRuleFound, &historyStub{values: []float64{1, 2}})
			v.err = tc.valErr
			w := post(dryRunRouter(tc.role, svc), tc.body)
			if w.Code != tc.want {
				t.Fatalf("status %d, want %d: %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestHandlerReturnsTheEnvelopeWithRowsAndSummary(t *testing.T) {
	svc, _ := newService(&models.Rule{Condition: expenseRule(5000)}, nil, &historyStub{values: []float64{6000}})
	w := post(dryRunRouter("ADMIN", svc), `{"request_type":"EXPENSE","grade_id":1,"condition":{"max_amount":8000},"values":[6000]}`)
	var out struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    Result `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !out.Success || !strings.Contains(out.Message, "Nothing was saved") {
		t.Fatalf("envelope %s", w.Body.String())
	}
	if len(out.Data.Rows) != 1 || out.Data.Rows[0].Draft != AutoApproved || out.Data.Summary.NewlyAuto != 1 {
		t.Fatalf("data %+v", out.Data)
	}
}
