package rule_drafting

import (
	"errors"
	"net/http"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler serves the admin rule-drafting endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type draftRequest struct {
	Sentence string `json:"sentence"`
}

// Status tells the UI whether to show the drafting box.
func (h *Handler) Status(c *gin.Context) {
	if c.GetString("role") != constants.RoleAdmin {
		response.Error(c, http.StatusForbidden, apperrors.ErrAdminOnly.Error(), nil)
		return
	}
	response.Success(c, "Rule drafting status", gin.H{"enabled": h.svc.Enabled()})
}

// Draft returns a proposed rule for a sentence. It never saves anything.
func (h *Handler) Draft(c *gin.Context) {
	var req draftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidRequestPayload.Error(), nil)
		return
	}
	draft, err := h.svc.Draft(c.Request.Context(), c.GetString("role"), req.Sentence)
	if err != nil {
		status := statusFor(err)
		if status == http.StatusBadGateway {
			response.Error(c, status, "rule drafting is unavailable right now", err.Error())
			return
		}
		response.Error(c, status, err.Error(), nil)
		return
	}
	response.Success(c, "Rule drafted. Review it, then confirm to save.", draft)
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, apperrors.ErrAdminOnly):
		return http.StatusForbidden
	case errors.Is(err, ErrDisabled):
		return http.StatusServiceUnavailable
	case errors.Is(err, ErrDailyLimit), errors.Is(err, ErrProviderQuota):
		return http.StatusTooManyRequests
	case errors.Is(err, ErrSentenceEmpty), errors.Is(err, ErrSentenceLong):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnclear), errors.Is(err, apperrors.ErrQuotaExceeded), errors.Is(err, apperrors.ErrNegativeValue):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusBadGateway
	}
}
