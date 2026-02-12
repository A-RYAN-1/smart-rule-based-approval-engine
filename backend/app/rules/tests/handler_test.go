package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/app/rules"
	"github.com/ankita-advitot/rule_based_approval_engine/app/rules/mocks"
	"github.com/ankita-advitot/rule_based_approval_engine/models"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRuleHandler_CreateRule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		reqBody        interface{}
		mockSetup      func(s *mocks.RuleService)
		expectedStatus int
	}{
		{
			name: "Success",
			role: "ADMIN",
			reqBody: models.Rule{
				RequestType: "LEAVE",
				Action:      "AUTO_APPROVE",
				GradeID:     1,
				Condition:   map[string]interface{}{"max_days": 5.0},
			},
			mockSetup: func(s *mocks.RuleService) {
				s.EXPECT().CreateRule(mock.Anything, "ADMIN", mock.AnythingOfType("models.Rule")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Unauthorized - Not Admin",
			role:           "EMPLOYEE",
			reqBody:        models.Rule{},
			mockSetup:      func(s *mocks.RuleService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid JSON",
			role:           "ADMIN",
			reqBody:        "{ invalid",
			mockSetup:      func(s *mocks.RuleService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			role: "ADMIN",
			reqBody: models.Rule{
				RequestType: "LEAVE",
				Action:      "AUTO_APPROVE",
				GradeID:     1,
			},
			mockSetup: func(s *mocks.RuleService) {
				s.EXPECT().CreateRule(mock.Anything, "ADMIN", mock.Anything).Return(apperrors.ErrDatabase)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewRuleService(t)
			tt.mockSetup(mockService)

			handler := rules.NewRuleHandler(nil, mockService)
			r := gin.New()
			r.POST("/rules", func(c *gin.Context) {
				c.Set("role", tt.role)
				handler.CreateRule(c)
			})

			var body []byte
			if s, ok := tt.reqBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}
			req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRuleHandler_UpdateRule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		ruleID         string
		reqBody        interface{}
		mockSetup      func(s *mocks.RuleService)
		expectedStatus int
	}{
		{
			name:   "Success",
			role:   "ADMIN",
			ruleID: "1",
			reqBody: models.Rule{
				Action: "MANUAL",
			},
			mockSetup: func(s *mocks.RuleService) {
				s.EXPECT().UpdateRule(mock.Anything, "ADMIN", int64(1), mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			role:           "ADMIN",
			ruleID:         "abc",
			reqBody:        models.Rule{},
			mockSetup:      func(s *mocks.RuleService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewRuleService(t)
			tt.mockSetup(mockService)

			handler := rules.NewRuleHandler(nil, mockService)
			r := gin.New()
			r.PUT("/rules/:id", func(c *gin.Context) {
				c.Set("role", tt.role)
				handler.UpdateRule(c)
			})

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, "/rules/"+tt.ruleID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRuleHandler_GetRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		mockSetup      func(s *mocks.RuleService)
		expectedStatus int
	}{
		{
			name: "Success",
			role: "ADMIN",
			mockSetup: func(s *mocks.RuleService) {
				s.EXPECT().GetRules(mock.Anything, "ADMIN").Return([]models.Rule{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized",
			role:           "MANAGER",
			mockSetup:      func(s *mocks.RuleService) {},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewRuleService(t)
			tt.mockSetup(mockService)

			handler := rules.NewRuleHandler(nil, mockService)
			r := gin.New()
			r.GET("/rules", func(c *gin.Context) {
				c.Set("role", tt.role)
				handler.GetRules(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/rules", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRuleHandler_DeleteRule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		ruleID         string
		mockSetup      func(s *mocks.RuleService)
		expectedStatus int
	}{
		{
			name:   "Success",
			role:   "ADMIN",
			ruleID: "1",
			mockSetup: func(s *mocks.RuleService) {
				s.EXPECT().DeleteRule(mock.Anything, "ADMIN", int64(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Rule Not Found",
			role:   "ADMIN",
			ruleID: "99",
			mockSetup: func(s *mocks.RuleService) {
				s.EXPECT().DeleteRule(mock.Anything, "ADMIN", int64(99)).Return(apperrors.ErrRuleNotFoundForDelete)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewRuleService(t)
			tt.mockSetup(mockService)

			handler := rules.NewRuleHandler(nil, mockService)
			r := gin.New()
			r.DELETE("/rules/:id", func(c *gin.Context) {
				c.Set("role", tt.role)
				handler.DeleteRule(c)
			})

			req := httptest.NewRequest(http.MethodDelete, "/rules/"+tt.ruleID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
