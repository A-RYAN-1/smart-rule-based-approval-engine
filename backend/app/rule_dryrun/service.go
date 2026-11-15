// Package rule_dryrun shows what a rule would decide before it is saved. It runs the same
// decision function the request flow uses, against values typed by the admin and against
// the last real requests of that type and grade. It never writes anything.
package rule_dryrun

import (
	"context"
	"errors"
	"math"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

const (
	// MaxSamples caps the values an admin can type in one call.
	MaxSamples = 20
	// HistorySize is how many recent real requests are replayed.
	HistorySize = 200

	// AutoApproved and NeedsApproval are the two outcomes a rule can give.
	AutoApproved  = "AUTO_APPROVED"
	NeedsApproval = "NEEDS_APPROVAL"

	SourceSample  = "sample"
	SourceHistory = "history"
)

var ErrTooManySamples = errors.New("too many sample values")
var ErrBadSample = errors.New("sample values must be positive numbers")

// Validator checks a rule against the grade limits, exactly as saving would.
type Validator interface {
	ValidateRule(ctx context.Context, rule models.Rule) error
}

// RuleLookup returns the rule that is saved today for a type and grade.
type RuleLookup interface {
	GetRule(ctx context.Context, requestType string, gradeID int64) (*models.Rule, error)
}

// History returns the values of the newest non-cancelled requests of a type for a grade:
// days for leave, amount for expenses, percent for discounts.
type History interface {
	RecentValues(ctx context.Context, requestType string, gradeID int64, limit int) ([]float64, error)
}

type Service struct {
	validator Validator
	lookup    RuleLookup
	history   History
}

func NewService(v Validator, l RuleLookup, h History) *Service {
	return &Service{validator: v, lookup: l, history: h}
}

// Request is the rule as it would be saved, plus optional values to try.
type Request struct {
	RequestType string                 `json:"request_type"`
	GradeID     int64                  `json:"grade_id"`
	Condition   map[string]interface{} `json:"condition"`
	Values      []float64              `json:"values"`
}

// Row is one value and what the saved rule and the draft rule decide for it.
type Row struct {
	Value   float64 `json:"value"`
	Source  string  `json:"source"`
	Current string  `json:"current"`
	Draft   string  `json:"draft"`
}

// Summary counts the replayed real requests only, never the typed samples.
type Summary struct {
	HistoryTotal int `json:"history_total"`
	CurrentAuto  int `json:"current_auto"`
	DraftAuto    int `json:"draft_auto"`
	NewlyAuto    int `json:"newly_auto"`
	NoLongerAuto int `json:"no_longer_auto"`
}

type Result struct {
	CurrentCondition map[string]interface{} `json:"current_condition"`
	DraftCondition   map[string]interface{} `json:"draft_condition"`
	Rows             []Row                  `json:"rows"`
	Summary          Summary                `json:"summary"`
}

func decide(condition map[string]interface{}, requestType string, value float64) string {
	if condition != nil && utils.Decide(requestType, condition, value) == constants.StatusAutoApprove {
		return AutoApproved
	}
	return NeedsApproval
}

// Run validates the draft, then decides every value under the saved rule and the draft.
func (s *Service) Run(ctx context.Context, role string, req Request) (*Result, error) {
	if role != constants.RoleAdmin {
		return nil, apperrors.ErrAdminOnly
	}
	if len(req.Values) > MaxSamples {
		return nil, ErrTooManySamples
	}
	for _, v := range req.Values {
		if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, ErrBadSample
		}
	}
	draft := models.Rule{
		RequestType: req.RequestType,
		GradeID:     req.GradeID,
		Condition:   req.Condition,
		Action:      constants.StatusAutoApprove,
		Active:      true,
	}
	if err := s.validator.ValidateRule(ctx, draft); err != nil {
		return nil, err
	}

	var current map[string]interface{}
	saved, err := s.lookup.GetRule(ctx, req.RequestType, req.GradeID)
	switch {
	case err == nil && saved != nil:
		current = saved.Condition
	case err != nil && !errors.Is(err, apperrors.ErrNoRuleFound):
		return nil, err
	}

	past, err := s.history.RecentValues(ctx, req.RequestType, req.GradeID, HistorySize)
	if err != nil {
		return nil, err
	}

	res := &Result{CurrentCondition: current, DraftCondition: req.Condition, Rows: []Row{}}
	row := func(value float64, source string) Row {
		return Row{
			Value:   value,
			Source:  source,
			Current: decide(current, req.RequestType, value),
			Draft:   decide(req.Condition, req.RequestType, value),
		}
	}
	for _, v := range req.Values {
		res.Rows = append(res.Rows, row(v, SourceSample))
	}
	for _, v := range past {
		r := row(v, SourceHistory)
		res.Summary.HistoryTotal++
		if r.Current == AutoApproved {
			res.Summary.CurrentAuto++
		}
		if r.Draft == AutoApproved {
			res.Summary.DraftAuto++
		}
		switch {
		case r.Current != AutoApproved && r.Draft == AutoApproved:
			res.Summary.NewlyAuto++
		case r.Current == AutoApproved && r.Draft != AutoApproved:
			res.Summary.NoLongerAuto++
		}
	}
	return res, nil
}
