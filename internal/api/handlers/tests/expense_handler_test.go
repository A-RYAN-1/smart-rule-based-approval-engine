// testcases for expense handler
package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"rule-based-approval-engine/internal/api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExpenseHandler_ApplyExpense_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewExpenseHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.ApplyExpense(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExpenseHandler_ApplyExpense_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewExpenseHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(1))

	c.Request = httptest.NewRequest("POST", "/expense", bytes.NewBufferString("invalid json"))
	handler.ApplyExpense(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
