package services

import (
	"context"
	"encoding/json"
	"errors"

	"rule-based-approval-engine/internal/database"
	"rule-based-approval-engine/internal/models"
)

func GetRule(requestType string, gradeID int64) (*models.Rule, error) {
	var rule models.Rule
	var conditionJSON []byte

	err := database.DB.QueryRow(
		context.Background(),
		`SELECT id, condition, action 
		 FROM rules 
		 WHERE request_type=$1 AND grade_id=$2 AND active=true
		 LIMIT 1`,
		requestType, gradeID,
	).Scan(&rule.ID, &conditionJSON, &rule.Action)

	if err != nil {
		return nil, errors.New("no rule found")
	}

	err = json.Unmarshal(conditionJSON, &rule.Condition)
	if err != nil {
		return nil, err
	}

	return &rule, nil
}
