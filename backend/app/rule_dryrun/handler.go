package rule_dryrun

import (
	"errors"
	"net/http"

	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler serves POST /api/admin/rules/dry-run.
type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// DryRun decides the sample values and recent real requests under a draft rule. Nothing is saved.
func (h *Handler) DryRun(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidRequestPayload.Error(), nil)
		return
	}
	res, err := h.svc.Run(c.Request.Context(), c.GetString("role"), req)
	if err != nil {
		response.Error(c, statusFor(err), err.Error(), nil)
		return
	}
	response.Success(c, "Dry run only. Nothing was saved.", res)
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, apperrors.ErrAdminOnly), errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusForbidden
	case errors.Is(err, apperrors.ErrQuotaExceeded), errors.Is(err, apperrors.ErrNegativeValue):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ErrTooManySamples), errors.Is(err, ErrBadSample),
		errors.Is(err, apperrors.ErrRequestTypeRequired), errors.Is(err, apperrors.ErrActionRequired),
		errors.Is(err, apperrors.ErrGradeIDRequired), errors.Is(err, apperrors.ErrConditionRequired):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
