package handlers

import (
	"net/http"
	"strconv"
	"time"

	"rule-based-approval-engine/internal/app/services"
	"rule-based-approval-engine/internal/pkg/apperrors"
	"rule-based-approval-engine/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type HolidayRequest struct {
	Date        string `json:"date"`
	Description string `json:"description"`
}

type HolidayHandler struct {
	holidayService *services.HolidayService
}

func NewHolidayHandler(holidayService *services.HolidayService) *HolidayHandler {
	return &HolidayHandler{holidayService: holidayService}
}

func (h *HolidayHandler) AddHoliday(c *gin.Context) {
	role := c.GetString("role")
	adminID := c.GetInt64("user_id")

	var req HolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid input",
			err.Error(),
		)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid date format (YYYY-MM-DD)",
			nil,
		)
		return
	}

	ctx := c.Request.Context()
	err = h.holidayService.AddHoliday(ctx, role, adminID, date, req.Description)
	if err != nil {
		handleHolidayError(c, err, "unable to add holiday")
		return
	}

	response.Created(
		c,
		"holiday added successfully",
		nil,
	)
}

func (h *HolidayHandler) GetHolidays(c *gin.Context) {
	role := c.GetString("role")
	ctx := c.Request.Context()

	holidays, err := h.holidayService.GetHolidays(ctx, role)
	if err != nil {
		handleHolidayError(c, err, "failed to fetch holidays")
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
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid holiday id",
			nil,
		)
		return
	}

	ctx := c.Request.Context()
	err = h.holidayService.DeleteHoliday(ctx, role, id)
	if err != nil {
		handleHolidayError(c, err, "unable to delete holiday")
		return
	}

	response.Success(
		c,
		"holiday removed successfully",
		nil,
	)
}

func handleHolidayError(c *gin.Context, err error, message string) {
	status := http.StatusInternalServerError

	if err == apperrors.ErrAdminOnly {
		status = http.StatusForbidden
	}

	response.Error(c, status, message, err.Error())
}
