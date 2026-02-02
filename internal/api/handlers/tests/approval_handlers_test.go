// testcases for approval handlers
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
	// gin has multiple modes as tests debug release
	gin.SetMode(gin.TestMode)
	// creates an instance of handler by passing the nil params
	handler := handlers.NewLeaveApprovalHandler(nil)
	// http response writer
	w := httptest.NewRecorder()
	// fake gin context
	c, _ := gin.CreateTestContext(w)
	// set the params
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.ApproveLeave(c)
	//	expected status and actual status
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
