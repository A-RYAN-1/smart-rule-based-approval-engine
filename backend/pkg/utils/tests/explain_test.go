package tests

import (
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func ptrF(v float64) *float64 { return &v }
func ptrI(v int64) *int64     { return &v }

func TestRuleLimit(t *testing.T) {
	assert.Equal(t, ptrF(3), utils.RuleLimit("LEAVE", map[string]interface{}{"max_days": 3.0}))
	assert.Equal(t, ptrF(5000), utils.RuleLimit("EXPENSE", map[string]interface{}{"max_amount": 5000.0}))
	assert.Equal(t, ptrF(10), utils.RuleLimit("DISCOUNT", map[string]interface{}{"max_percent": 10.0}))
	assert.Nil(t, utils.RuleLimit("LEAVE", map[string]interface{}{"max_amount": 5.0}))
	assert.Nil(t, utils.RuleLimit("LEAVE", nil))
	assert.Nil(t, utils.RuleLimit("OTHER", map[string]interface{}{"max_days": 3.0}))
}

func TestExplainDecision(t *testing.T) {
	tests := []struct {
		name   string
		typ    string
		status string
		rule   *int64
		limit  *float64
		value  float64
		kind   string
		text   string
	}{
		{"auto approved leave", "LEAVE", "AUTO_APPROVED", ptrI(7), ptrF(3), 2, "AUTO_APPROVED",
			"Approved by the system: 2 days is within the limit of 3 days set by rule #7."},
		{"one day is singular", "LEAVE", "AUTO_APPROVED", ptrI(7), ptrF(3), 1, "AUTO_APPROVED",
			"Approved by the system: 1 day is within the limit of 3 days set by rule #7."},
		{"expense over limit pending", "EXPENSE", "PENDING", ptrI(8), ptrF(5000), 7500.5, "SENT_TO_PERSON",
			"Sent to a person: 7500.5 is over the limit of 5000 set by rule #8."},
		{"discount decided later by a manager", "DISCOUNT", "APPROVED", ptrI(9), ptrF(10), 15, "SENT_TO_PERSON",
			"Sent to a person: 15% is over the limit of 10% set by rule #9."},
		{"old rejected row without limit", "LEAVE", "REJECTED", ptrI(7), nil, 5, "SENT_TO_PERSON",
			"Sent to a person for approval by rule #7 (its limit was not recorded)."},
		{"old auto approved row without limit", "LEAVE", "AUTO_APPROVED", ptrI(7), nil, 2, "AUTO_APPROVED",
			"Approved by the system under rule #7 (its limit was not recorded)."},
		{"no rule", "LEAVE", "PENDING", nil, nil, 2, "SENT_TO_PERSON", "Sent to a person for approval."},
		{"auto rejected", "EXPENSE", "AUTO_REJECTED", ptrI(8), ptrF(5000), 9000, "AUTO_REJECTED",
			"Rejected by the system: nobody decided it within 7 working days."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.ExplainDecision(tc.typ, tc.status, tc.rule, tc.limit, tc.value)
			assert.Equal(t, tc.kind, got["kind"])
			assert.Equal(t, tc.text, got["text"])
		})
	}
}

func TestExplainDecisionCancelledHasNone(t *testing.T) {
	assert.Nil(t, utils.ExplainDecision("LEAVE", "CANCELLED", ptrI(1), ptrF(3), 2))
}

func TestExplainDecisionCarriesRuleAndLimit(t *testing.T) {
	got := utils.ExplainDecision("EXPENSE", "AUTO_APPROVED", ptrI(8), ptrF(5000), 100)
	assert.Equal(t, int64(8), got["rule_id"])
	assert.Equal(t, float64(5000), got["limit"])

	none := utils.ExplainDecision("EXPENSE", "PENDING", nil, nil, 100)
	assert.Nil(t, none["rule_id"])
	assert.Nil(t, none["limit"])
}
