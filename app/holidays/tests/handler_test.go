package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/app/holidays"
	"github.com/ankita-advitot/rule_based_approval_engine/app/holidays/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHolidayHandler_AddHoliday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		userID         int64
		reqBody        interface{}
		mockSetup      func(s *mocks.HolidayService)
		expectedStatus int
	}{
		{
			name:   "Success",
			role:   "ADMIN",
			userID: 1,
			reqBody: holidays.HolidayRequest{
				Date:        "2026-01-01",
				Description: "New Year",
			},
			mockSetup: func(s *mocks.HolidayService) {
				date, _ := time.Parse("2006-01-02", "2026-01-01")
				s.EXPECT().AddHoliday(mock.Anything, "ADMIN", int64(1), date, "New Year").Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid Date Format",
			role:   "ADMIN",
			userID: 1,
			reqBody: holidays.HolidayRequest{
				Date: "01-01-2026",
			},
			mockSetup:      func(s *mocks.HolidayService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Unauthorized",
			role:    "EMPLOYEE",
			userID:  2,
			reqBody: holidays.HolidayRequest{Date: "2026-01-01"},
			mockSetup: func(s *mocks.HolidayService) {
				date, _ := time.Parse("2006-01-02", "2026-01-01")
				s.EXPECT().AddHoliday(mock.Anything, "EMPLOYEE", int64(2), date, "").Return(apperrors.ErrAdminOnly)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewHolidayService(t)
			tt.mockSetup(mockS)

			handler := holidays.NewHolidayHandler(context.Background(), mockS)
			r := gin.New()
			r.POST("/add", func(c *gin.Context) {
				c.Set("role", tt.role)
				c.Set("user_id", tt.userID)
				handler.AddHoliday(c)
			})

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHolidayHandler_GetHolidays(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockS := mocks.NewHolidayService(t)
		mockS.EXPECT().GetHolidays(mock.Anything, "ADMIN").Return([]map[string]interface{}{{"id": 1}}, nil)

		handler := holidays.NewHolidayHandler(context.Background(), mockS)
		r := gin.New()
		r.GET("/list", func(c *gin.Context) {
			c.Set("role", "ADMIN")
			handler.GetHolidays(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/list", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHolidayHandler_DeleteHoliday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		id             string
		mockSetup      func(s *mocks.HolidayService)
		expectedStatus int
	}{
		{
			name: "Success",
			role: "ADMIN",
			id:   "10",
			mockSetup: func(s *mocks.HolidayService) {
				s.EXPECT().DeleteHoliday(mock.Anything, "ADMIN", int64(10)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			role:           "ADMIN",
			id:             "abc",
			mockSetup:      func(s *mocks.HolidayService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS := mocks.NewHolidayService(t)
			tt.mockSetup(mockS)

			handler := holidays.NewHolidayHandler(context.Background(), mockS)
			r := gin.New()
			r.DELETE("/delete/:id", func(c *gin.Context) {
				c.Set("role", tt.role)
				handler.DeleteHoliday(c)
			})

			req := httptest.NewRequest(http.MethodDelete, "/delete/"+tt.id, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
