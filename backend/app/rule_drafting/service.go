// Package rule_drafting turns an administrator's sentence into a draft
// auto-approval rule. It only drafts: nothing here writes to the database. The
// administrator confirms a draft in the UI, and the confirmed rule is saved
// through the ordinary rules endpoint, which validates it again.
package rule_drafting

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
)

// MaxSentenceLength caps the input so a prompt cannot be padded.
const MaxSentenceLength = 300

var (
	ErrDisabled      = errors.New("rule drafting is not configured")
	ErrDailyLimit    = errors.New("daily rule drafting limit reached")
	ErrSentenceEmpty = errors.New("sentence is required")
	ErrSentenceLong  = errors.New("sentence is too long")
	ErrUnclear       = errors.New("could not turn that into one rule")
	ErrBadModelReply = errors.New("the model returned an unusable reply")
)

// Validator is the rules service's validation, reused so a draft can never be
// a rule the save endpoint would reject.
type Validator interface {
	ValidateRule(ctx context.Context, rule models.Rule) error
}

// GradeSource lists the grades that exist.
type GradeSource interface {
	ListGrades(ctx context.Context) ([]Grade, error)
}

// RuleLookup returns the current rule for a type and grade, or an error when none exists.
type RuleLookup interface {
	GetRule(ctx context.Context, requestType string, gradeID int64) (*models.Rule, error)
}

// Change is the one field of the diff shown to the administrator.
type Change struct {
	Field  string   `json:"field"`
	Before *float64 `json:"before"`
	After  float64  `json:"after"`
}

// Draft is what the endpoint returns. Rule is a proposal, never a saved row.
type Draft struct {
	Rule          models.Rule  `json:"rule"`
	GradeName     string       `json:"grade_name"`
	Existing      *models.Rule `json:"existing"`
	Change        Change       `json:"change"`
	PromptVersion string       `json:"prompt_version"`
}

type Service struct {
	provider  Provider
	validator Validator
	grades    GradeSource
	lookup    RuleLookup
	limiter   *dailyLimiter
}

// NewService builds the drafting service. A nil provider means the feature is
// switched off and Draft returns ErrDisabled.
func NewService(p Provider, v Validator, g GradeSource, l RuleLookup, dailyLimit int) *Service {
	return &Service{provider: p, validator: v, grades: g, lookup: l, limiter: newDailyLimiter(dailyLimit, time.Now)}
}

// Enabled reports whether a model is configured.
func (s *Service) Enabled() bool { return s.provider != nil }

// limitField maps each request type to the single condition field it uses.
var limitField = map[string]string{
	"LEAVE":    "max_days",
	"EXPENSE":  "max_amount",
	"DISCOUNT": "max_percent",
}

type modelReply struct {
	OK          bool    `json:"ok"`
	Reason      string  `json:"reason"`
	RequestType string  `json:"request_type"`
	GradeID     int64   `json:"grade_id"`
	Value       float64 `json:"value"`
}

// Draft asks the model for a rule and checks everything it says.
func (s *Service) Draft(ctx context.Context, role, sentence string) (*Draft, error) {
	if role != constants.RoleAdmin {
		return nil, apperrors.ErrAdminOnly
	}
	if s.provider == nil {
		return nil, ErrDisabled
	}
	sentence = strings.TrimSpace(sentence)
	if sentence == "" {
		return nil, ErrSentenceEmpty
	}
	if len([]rune(sentence)) > MaxSentenceLength {
		return nil, ErrSentenceLong
	}

	grades, err := s.grades.ListGrades(ctx)
	if err != nil {
		return nil, err
	}
	if !s.limiter.allow() {
		return nil, ErrDailyLimit
	}

	text, err := s.provider.Complete(ctx, SystemPrompt(grades), sentence)
	if err != nil {
		return nil, err
	}
	return s.build(ctx, grades, text)
}

// build parses and validates the model's reply. Every field is checked against
// data the server owns, so a wrong answer becomes an error, not a bad rule.
func (s *Service) build(ctx context.Context, grades []Grade, text string) (*Draft, error) {
	var reply modelReply
	dec := json.NewDecoder(strings.NewReader(text))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&reply); err != nil {
		return nil, ErrBadModelReply
	}
	if !reply.OK {
		return nil, ErrUnclear
	}

	field, known := limitField[reply.RequestType]
	if !known {
		return nil, ErrBadModelReply
	}
	var gradeName string
	for _, g := range grades {
		if g.ID == reply.GradeID {
			gradeName = g.Name
		}
	}
	if gradeName == "" {
		return nil, ErrBadModelReply
	}
	if math.IsNaN(reply.Value) || math.IsInf(reply.Value, 0) || reply.Value <= 0 {
		return nil, ErrBadModelReply
	}
	if reply.RequestType == "LEAVE" && reply.Value != math.Trunc(reply.Value) {
		return nil, ErrBadModelReply
	}

	rule := models.Rule{
		RequestType: reply.RequestType,
		Condition:   map[string]interface{}{field: reply.Value},
		Action:      constants.StatusAutoApprove,
		GradeID:     reply.GradeID,
		Active:      true,
	}
	if err := s.validator.ValidateRule(ctx, rule); err != nil {
		return nil, err
	}

	draft := &Draft{Rule: rule, GradeName: gradeName, PromptVersion: PromptVersion,
		Change: Change{Field: field, After: reply.Value}}
	if existing, err := s.lookup.GetRule(ctx, rule.RequestType, rule.GradeID); err == nil && existing != nil {
		draft.Existing = existing
		if before, ok := existing.Condition[field].(float64); ok {
			draft.Change.Before = &before
		}
	}
	return draft, nil
}

// dailyLimiter is a best-effort per-process cap on model calls per UTC day. On
// a serverless host each instance counts separately, so it is a cost guard,
// not an exact quota.
type dailyLimiter struct {
	mu    sync.Mutex
	limit int
	now   func() time.Time
	day   string
	count int
}

func newDailyLimiter(limit int, now func() time.Time) *dailyLimiter {
	return &dailyLimiter{limit: limit, now: now}
}

func (l *dailyLimiter) allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	today := l.now().UTC().Format("2006-01-02")
	if today != l.day {
		l.day, l.count = today, 0
	}
	if l.count >= l.limit {
		return false
	}
	l.count++
	return true
}
