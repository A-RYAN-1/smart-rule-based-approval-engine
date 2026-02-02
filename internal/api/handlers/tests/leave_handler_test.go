// testcases for leave handler
package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rule-based-approval-engine/internal/api/handlers"
	"rule-based-approval-engine/internal/app/repositories/mocks"
	"rule-based-approval-engine/internal/app/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLeaveHandler_ApplyLeave_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewLeaveHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// No user_id in context
	handler.ApplyLeave(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLeaveHandler_ApplyLeave_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewLeaveHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(1))

	c.Request = httptest.NewRequest("POST", "/apply", bytes.NewBufferString("invalid json"))
	handler.ApplyLeave(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeaveHandler_ApplyLeave_InvalidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewLeaveHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(1))

	body, _ := json.Marshal(map[string]string{
		"from_date":  "invalid",
		"to_date":    "2026-01-01",
		"leave_type": "SICK",
	})

	c.Request = httptest.NewRequest("POST", "/apply", bytes.NewBuffer(body))
	handler.ApplyLeave(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeaveHandler_ApplyLeave_ValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create service with mocked repo to test validation flow in handler
	mockLeaveRepo := mocks.NewLeaveRequestRepository(t)
	service := services.NewLeaveService(mockLeaveRepo, nil, nil, nil, nil)
	handler := handlers.NewLeaveHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(1))

	// Past date should fail validation in service
	pastDate := time.Now().Add(-48 * time.Hour).Format("2006-01-02")
	futureDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	body, _ := json.Marshal(map[string]string{
		"from_date":  pastDate,
		"to_date":    futureDate,
		"leave_type": "SICK",
	})

	c.Request = httptest.NewRequest("POST", "/apply", bytes.NewBuffer(body))
	handler.ApplyLeave(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "cannot be in the past")
}

func TestLeaveHandler_CancelLeave_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewLeaveHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.CancelLeave(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
