package rule_drafting

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func draftRouter(t *testing.T, role string, p Provider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := NewHandler(newTestService(t, p, nil))
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("role", role) })
	r.GET("/status", h.Status)
	r.POST("/draft", h.Draft)
	return r
}

func post(r *gin.Engine, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/draft", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestHandlerStatusCodes(t *testing.T) {
	good := &fakeProvider{reply: `{"ok":true,"request_type":"DISCOUNT","grade_id":3,"value":25}`}
	cases := []struct {
		name string
		role string
		p    Provider
		body string
		want int
	}{
		{"drafts for admin", "ADMIN", good, `{"sentence":"grade 3 discount 25%"}`, http.StatusOK},
		{"employee is forbidden", "EMPLOYEE", good, `{"sentence":"x"}`, http.StatusForbidden},
		{"bad json", "ADMIN", good, `not json`, http.StatusBadRequest},
		{"empty sentence", "ADMIN", good, `{"sentence":""}`, http.StatusBadRequest},
		{"model cannot draft", "ADMIN", &fakeProvider{reply: `{"ok":false}`}, `{"sentence":"x"}`, http.StatusUnprocessableEntity},
		{"over the limit", "ADMIN", &fakeProvider{reply: `{"ok":true,"request_type":"EXPENSE","grade_id":2,"value":99999}`}, `{"sentence":"x"}`, http.StatusUnprocessableEntity},
		{"provider down", "ADMIN", &fakeProvider{err: ErrBadModelReply}, `{"sentence":"x"}`, http.StatusBadGateway},
		{"quota used up", "ADMIN", &fakeProvider{err: ErrProviderQuota}, `{"sentence":"x"}`, http.StatusTooManyRequests},
		{"not configured", "ADMIN", nil, `{"sentence":"x"}`, http.StatusServiceUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if w := post(draftRouter(t, tc.role, tc.p), tc.body); w.Code != tc.want {
				t.Fatalf("status %d, want %d: %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestHandlerDoesNotLeakProviderDetail(t *testing.T) {
	w := post(draftRouter(t, "ADMIN", &fakeProvider{err: http.ErrHandlerTimeout}), `{"sentence":"x"}`)
	if strings.Contains(w.Body.String(), http.ErrHandlerTimeout.Error()) {
		t.Fatalf("provider error text reached the client: %s", w.Body.String())
	}
}

func TestStatusReportsEnabled(t *testing.T) {
	for _, tc := range []struct {
		p    Provider
		want bool
	}{{&fakeProvider{}, true}, {nil, false}} {
		w := httptest.NewRecorder()
		draftRouter(t, "ADMIN", tc.p).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/status", nil))
		var body struct {
			Data struct {
				Enabled bool `json:"enabled"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Data.Enabled != tc.want {
			t.Fatalf("enabled=%v want %v (%s)", body.Data.Enabled, tc.want, w.Body.String())
		}
	}
}
