// File: internal/api/handlers/tests/approval_handlers_test.go
package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"rule-based-approval-engine/internal/api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLeaveApprovalHandler_Approve_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewLeaveApprovalHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.ApproveLeave(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExpenseApprovalHandler_Approve_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewExpenseApprovalHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.ApproveExpense(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDiscountApprovalHandler_Approve_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewDiscountApprovalHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.ApproveDiscount(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
