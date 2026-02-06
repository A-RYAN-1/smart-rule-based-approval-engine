package rules

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

// handles rule-related HTTP requests
type RuleHandler struct {
	ruleService interfaces.RuleService
}

// creates a new RuleHandler instance
func NewRuleHandler(ctx context.Context, ruleService interfaces.RuleService) *RuleHandler {
	return &RuleHandler{ruleService: ruleService}
}

func (h *RuleHandler) CreateRule(c *gin.Context) {
	role := c.GetString("role")

	if role != constants.RoleAdmin {
		handleRuleError(c, apperrors.ErrAdminOnly, nil)
		return
	}

	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		handleRuleError(c, apperrors.ErrInvalidRequestPayload, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.ruleService.CreateRule(ctx, role, rule); err != nil {
		handleRuleError(c, err, nil)
		return
	}

	response.Created(c, "Rule created successfully", nil)
}

func (h *RuleHandler) GetRules(c *gin.Context) {
	role := c.GetString("role")
	if role != constants.RoleAdmin {
		handleRuleError(c, apperrors.ErrAdminOnly, nil)
		return
	}

	ctx := c.Request.Context()
	rules, err := h.ruleService.GetRules(ctx, role)
	if err != nil {
		handleRuleError(c, err, nil)
		return
	}

	response.Success(c, "Rules fetched successfully", rules)
}

func (h *RuleHandler) UpdateRule(c *gin.Context) {
	role := c.GetString("role")

	if role != constants.RoleAdmin {
		handleRuleError(c, apperrors.ErrAdminOnly, nil)
		return
	}

	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleRuleError(c, apperrors.ErrInvalidID, err)
		return
	}

	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		handleRuleError(c, apperrors.ErrInvalidRequestPayload, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.ruleService.UpdateRule(ctx, role, ruleID, rule); err != nil {
		handleRuleError(c, err, nil)
		return
	}

	response.Success(c, "Rule updated successfully", nil)
}

func (h *RuleHandler) DeleteRule(c *gin.Context) {
	role := c.GetString("role")

	if role != constants.RoleAdmin {
		handleRuleError(c, apperrors.ErrAdminOnly, nil)
		return
	}

	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleRuleError(c, apperrors.ErrInvalidID, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.ruleService.DeleteRule(ctx, role, ruleID); err != nil {
		handleRuleError(c, err, nil)
		return
	}

	response.Success(c, "Rule deleted successfully", nil)
}

func handleRuleError(c *gin.Context, err error, detail error) {
	status := http.StatusInternalServerError
	switch err {
	case apperrors.ErrUnauthorized, apperrors.ErrAdminOnly:
		status = http.StatusForbidden
	case apperrors.ErrNoRuleFound, apperrors.ErrRuleNotFoundForDelete:
		status = http.StatusNotFound
	case apperrors.ErrRequestTypeRequired, apperrors.ErrActionRequired,
		apperrors.ErrGradeIDRequired, apperrors.ErrConditionRequired,
		apperrors.ErrInvalidConditionJSON, apperrors.ErrInvalidID,
		apperrors.ErrInvalidRequestPayload:
		status = http.StatusBadRequest
	}

	message := err.Error()
	var errDetail interface{}
	if detail != nil {
		errDetail = detail.Error()
	}

	response.Error(c, status, message, errDetail)
}
