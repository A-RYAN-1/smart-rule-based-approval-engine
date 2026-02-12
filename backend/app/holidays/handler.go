package holidays

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/response"

	"github.com/gin-gonic/gin"
)

type HolidayHandler struct {
	holidayService interfaces.HolidayService
}

func NewHolidayHandler(ctx context.Context, holidayService interfaces.HolidayService) *HolidayHandler {
	return &HolidayHandler{holidayService: holidayService}
}

func (h *HolidayHandler) AddHoliday(c *gin.Context) {
	role := c.GetString("role")
	adminID := c.GetInt64("user_id")

	var req HolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleHolidayError(c, apperrors.ErrInvalidInput)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		handleHolidayError(c, apperrors.ErrInvalidDateFormat)
		return
	}

	ctx := c.Request.Context()
	err = h.holidayService.AddHoliday(ctx, role, adminID, date, req.Description)
	if err != nil {
		handleHolidayError(c, err)
		return
	}

	response.Created(c, "holiday added successfully", nil)
}

func (h *HolidayHandler) GetHolidays(c *gin.Context) {
	role := c.GetString("role")
	ctx := c.Request.Context()

	holidays, err := h.holidayService.GetHolidays(ctx, role)
	if err != nil {
		handleHolidayError(c, err)
		return
	}

	response.Success(
		c,
		"holidays fetched successfully",
		holidays,
	)
}

func (h *HolidayHandler) DeleteHoliday(c *gin.Context) {
	role := c.GetString("role")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		handleHolidayError(c, apperrors.ErrInvalidID)
		return
	}

	ctx := c.Request.Context()
	err = h.holidayService.DeleteHoliday(ctx, role, id)
	if err != nil {
		handleHolidayError(c, err)
		return
	}

	response.Success(c, "holiday removed successfully", nil)
}

func handleHolidayError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch err {
	case apperrors.ErrAdminOnly:
		status = http.StatusForbidden
	case apperrors.ErrInvalidDateFormat, apperrors.ErrInvalidInput, apperrors.ErrInvalidID:
		status = http.StatusBadRequest
	}

	response.Error(c, status, err.Error(), nil)
}
