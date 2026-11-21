package utils

import (
	"fmt"
	"strconv"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
)

// Decision kinds returned by ExplainDecision.
const (
	DecisionAutoApproved  = "AUTO_APPROVED"
	DecisionSentToPerson  = "SENT_TO_PERSON"
	DecisionAutoRejected  = "AUTO_REJECTED"
	limitKeyLeave         = "max_days"
	limitKeyExpense       = "max_amount"
	limitKeyDiscount      = "max_percent"
	notRecordedSuffixText = " (its limit was not recorded)"
)

// RuleLimit returns the limit a rule applies to a request type, or nil when
// the rule has none. It is stored with the request so the explanation keeps the
// limit that applied at submission even if the rule is edited later.
func RuleLimit(requestType string, rule map[string]interface{}) *float64 {
	key := map[string]string{
		"LEAVE":    limitKeyLeave,
		"EXPENSE":  limitKeyExpense,
		"DISCOUNT": limitKeyDiscount,
	}[requestType]
	v, ok := rule[key].(float64)
	if !ok {
		return nil
	}
	return &v
}

// ExplainDecision says why a request has the status it has: which rule cleared
// it, why it went to a person, or that the scheduled job rejected it. It returns
// nil for a cancelled request, which has no routing to explain.
func ExplainDecision(requestType, status string, ruleID *int64, limit *float64, value float64) map[string]interface{} {
	var kind, text string
	switch status {
	case constants.StatusCancelled:
		return nil
	case constants.StatusAutoRejected:
		kind = DecisionAutoRejected
		text = "Rejected by the system: nobody decided it within 7 working days."
	case constants.StatusAutoApproved:
		kind = DecisionAutoApproved
		text = autoApprovedText(requestType, ruleID, limit, value)
	default:
		kind = DecisionSentToPerson
		text = sentToPersonText(requestType, ruleID, limit, value)
	}

	out := map[string]interface{}{"kind": kind, "text": text, "rule_id": nil, "limit": nil}
	if ruleID != nil {
		out["rule_id"] = *ruleID
	}
	if limit != nil {
		out["limit"] = *limit
	}
	return out
}

func autoApprovedText(requestType string, ruleID *int64, limit *float64, value float64) string {
	switch {
	case ruleID == nil:
		return "Approved by the system."
	case limit == nil:
		return fmt.Sprintf("Approved by the system under rule #%d%s.", *ruleID, notRecordedSuffixText)
	}
	return fmt.Sprintf("Approved by the system: %s is within the limit of %s set by rule #%d.",
		amountText(requestType, value), amountText(requestType, *limit), *ruleID)
}

func sentToPersonText(requestType string, ruleID *int64, limit *float64, value float64) string {
	switch {
	case ruleID == nil:
		return "Sent to a person for approval."
	case limit == nil:
		return fmt.Sprintf("Sent to a person for approval by rule #%d%s.", *ruleID, notRecordedSuffixText)
	}
	return fmt.Sprintf("Sent to a person: %s is over the limit of %s set by rule #%d.",
		amountText(requestType, value), amountText(requestType, *limit), *ruleID)
}

func amountText(requestType string, v float64) string {
	n := strconv.FormatFloat(v, 'f', -1, 64)
	switch requestType {
	case "LEAVE":
		if v == 1 {
			return n + " day"
		}
		return n + " days"
	case "DISCOUNT":
		return n + "%"
	}
	return n
}
