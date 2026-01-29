package handlers

import (
	"net/http"
	"strconv"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/models"
	"rule-based-approval-engine/internal/pkg/apperrors"
	"rule-based-approval-engine/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// RuleHandler handles rule-related HTTP requests
type RuleHandler struct {
	ruleService *services.RuleService
}

// NewRuleHandler creates a new RuleHandler instance
func NewRuleHandler(ruleService *services.RuleService) *RuleHandler {
	return &RuleHandler{ruleService: ruleService}
}

func (h *RuleHandler) CreateRule(c *gin.Context) {
	role := c.GetString("role")

	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := h.ruleService.CreateRule(ctx, role, rule); err != nil {
		handleRuleError(c, err, "Failed to create rule")
		return
	}

	response.Created(c, "Rule created successfully", nil)
}

func (h *RuleHandler) GetRules(c *gin.Context) {
	role := c.GetString("role")

	ctx := c.Request.Context()
	rules, err := h.ruleService.GetRules(ctx, role)
	if err != nil {
		handleRuleError(c, err, "Failed to fetch rules")
		return
	}

	response.Success(c, "Rules fetched successfully", rules)
}

func (h *RuleHandler) UpdateRule(c *gin.Context) {
	role := c.GetString("role")

	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid rule ID", err.Error())
		return
	}

	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := h.ruleService.UpdateRule(ctx, role, ruleID, rule); err != nil {
		handleRuleError(c, err, "Failed to update rule")
		return
	}

	response.Success(c, "Rule updated successfully", nil)
}

func (h *RuleHandler) DeleteRule(c *gin.Context) {
	role := c.GetString("role")

	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid rule ID", err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := h.ruleService.DeleteRule(ctx, role, ruleID); err != nil {
		handleRuleError(c, err, "Failed to delete rule")
		return
	}

	response.Success(c, "Rule deleted successfully", nil)
}

func handleRuleError(c *gin.Context, err error, message string) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrUnauthorized:
		status = http.StatusForbidden
	case apperrors.ErrNoRuleFound, apperrors.ErrRuleNotFoundForDelete:
		status = http.StatusNotFound
	case apperrors.ErrRequestTypeRequired, apperrors.ErrActionRequired,
		apperrors.ErrGradeIDRequired, apperrors.ErrConditionRequired,
		apperrors.ErrInvalidConditionJSON:
		status = http.StatusBadRequest
	}

	response.Error(c, status, message, err.Error())
}
