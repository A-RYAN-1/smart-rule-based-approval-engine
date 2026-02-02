// testcases for discount handler
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

func TestDiscountHandler_ApplyDiscount_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewDiscountHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.ApplyDiscount(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDiscountHandler_ApplyDiscount_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewDiscountHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(1))

	c.Request = httptest.NewRequest("POST", "/discount", bytes.NewBufferString("invalid json"))
	handler.ApplyDiscount(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
