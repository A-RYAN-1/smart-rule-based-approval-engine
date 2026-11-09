package rule_drafting

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
)

//go:embed testdata/phrasings.json
var phrasingsJSON []byte

// EvalGrade is a grade with the limits the validator enforces.
type EvalGrade struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Leave    int     `json:"leave"`
	Expense  float64 `json:"expense"`
	Discount float64 `json:"discount"`
}

// Expectation is the rule a sentence must produce, or a rejection.
type Expectation struct {
	Reject      bool    `json:"reject"`
	RequestType string  `json:"request_type"`
	GradeID     int64   `json:"grade_id"`
	Value       float64 `json:"value"`
}

// EvalCase is one hand-written phrasing and its expected outcome.
type EvalCase struct {
	Sentence string      `json:"sentence"`
	Expect   Expectation `json:"expect"`
	Why      string      `json:"why,omitempty"`
}

// EvalSet is the checked-in evaluation set.
type EvalSet struct {
	PromptVersion string      `json:"prompt_version"`
	Grades        []EvalGrade `json:"grades"`
	Cases         []EvalCase  `json:"cases"`
}

// LoadEvalSet returns the embedded evaluation set.
func LoadEvalSet() (*EvalSet, error) {
	var set EvalSet
	if err := json.Unmarshal(phrasingsJSON, &set); err != nil {
		return nil, err
	}
	return &set, nil
}

// limitsValidator applies the same limit checks as the rules service, using
// the grade limits stored in the evaluation set instead of the database.
type limitsValidator struct{ grades map[int64]EvalGrade }

func (v limitsValidator) ValidateRule(_ context.Context, rule models.Rule) error {
	g, ok := v.grades[rule.GradeID]
	if !ok {
		return apperrors.ErrGradeIDRequired
	}
	for key, raw := range rule.Condition {
		val, ok := raw.(float64)
		if !ok {
			continue
		}
		if val < 0 {
			return apperrors.ErrNegativeValue
		}
		over := (rule.RequestType == "LEAVE" && key == "max_days" && val > float64(g.Leave)) ||
			(rule.RequestType == "EXPENSE" && key == "max_amount" && val > g.Expense) ||
			(rule.RequestType == "DISCOUNT" && key == "max_percent" && val > g.Discount)
		if over {
			return apperrors.ErrQuotaExceeded
		}
	}
	return nil
}

type staticGrades []Grade

func (s staticGrades) ListGrades(context.Context) ([]Grade, error) { return s, nil }

type noRules struct{}

func (noRules) GetRule(context.Context, string, int64) (*models.Rule, error) {
	return nil, apperrors.ErrNoRuleFound
}

// CaseResult is the outcome of one case.
type CaseResult struct {
	Sentence string `json:"sentence"`
	Pass     bool   `json:"pass"`
	Got      string `json:"got"`
	Want     string `json:"want"`
}

// Report is the measured result of one run.
type Report struct {
	PromptVersion string       `json:"prompt_version"`
	Total         int          `json:"total"`
	ExactMatch    int          `json:"exact_match"`
	Rules         int          `json:"rule_cases"`
	RulesMatched  int          `json:"rule_cases_matched"`
	Rejects       int          `json:"reject_cases"`
	RejectsKept   int          `json:"reject_cases_rejected"`
	Results       []CaseResult `json:"results"`
}

// Evaluate runs every case through the same Service code path the endpoint
// uses (parsing, grade check, validator) and compares the draft to the
// expectation. A case passes only when every field matches exactly.
func Evaluate(ctx context.Context, p Provider, set *EvalSet) (*Report, error) {
	grades := make([]Grade, 0, len(set.Grades))
	byID := make(map[int64]EvalGrade, len(set.Grades))
	for _, g := range set.Grades {
		grades = append(grades, Grade{ID: g.ID, Name: g.Name})
		byID[g.ID] = g
	}
	svc := NewService(p, limitsValidator{byID}, staticGrades(grades), noRules{}, len(set.Cases)+1)

	rep := &Report{PromptVersion: PromptVersion, Total: len(set.Cases)}
	for _, c := range set.Cases {
		res := CaseResult{Sentence: c.Sentence}
		draft, err := svc.Draft(ctx, "ADMIN", c.Sentence)
		if errors.Is(err, ErrProviderQuota) {
			return rep, err
		}
		if c.Expect.Reject {
			rep.Rejects++
			res.Want = "reject"
			res.Pass = err != nil
			res.Got = "reject"
			if err == nil {
				res.Got = describe(draft)
			} else {
				rep.RejectsKept++
			}
		} else {
			rep.Rules++
			res.Want = fmt.Sprintf("%s grade %d value %g", c.Expect.RequestType, c.Expect.GradeID, c.Expect.Value)
			if err != nil {
				res.Got = "error: " + err.Error()
			} else {
				res.Got = describe(draft)
				res.Pass = res.Got == res.Want
				if res.Pass {
					rep.RulesMatched++
				}
			}
		}
		if res.Pass {
			rep.ExactMatch++
		}
		rep.Results = append(rep.Results, res)
	}
	return rep, nil
}

func describe(d *Draft) string {
	return fmt.Sprintf("%s grade %d value %g", d.Rule.RequestType, d.Rule.GradeID, d.Change.After)
}
