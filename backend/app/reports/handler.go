package reports

import (
	"context"
	"net/http"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService interfaces.ReportService
}

func NewReportHandler(ctx context.Context, reportService interfaces.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) GetDashboardSummary(c *gin.Context) {
	role := c.GetString("role")
	ctx := c.Request.Context()
	data, err := h.reportService.GetDashboardSummary(ctx, role)
	if err != nil {
		handleReportError(c, err)
		return
	}

	response.Success(
		c,
		"Dashboard summary fetched successfully",
		data,
	)
}

func handleReportError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	if err == apperrors.ErrAdminOnly || err == apperrors.ErrUnauthorized {
		status = http.StatusForbidden
	}

	response.Error(c, status, err.Error(), nil)
}
