// File: internal/app/repositories/tests/expense_request_repository_test.go
package tests

import (
	"context"
	"testing"

	"rule-based-approval-engine/internal/app/repositories"
	"rule-based-approval-engine/internal/models"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpenseRequestRepository_Create_Success(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mockTx.Rollback(context.Background())

	repo := repositories.NewExpenseRequestRepository(nil)
	req := &models.ExpenseRequest{
		EmployeeID: 1,
		Amount:     500.0,
		Category:   "Travel",
		Reason:     "Client visit",
		Status:     "PENDING",
	}

	mockTx.ExpectExec("INSERT INTO expense_requests").
		WithArgs(req.EmployeeID, req.Amount, req.Category, req.Reason, req.Status, req.RuleID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), mockTx, req)
	assert.NoError(t, err)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestExpenseRequestRepository_GetByID_Success(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mockTx.Rollback(context.Background())

	repo := repositories.NewExpenseRequestRepository(nil)
	reqID := int64(123)

	mockTx.ExpectQuery("SELECT employee_id, status, amount FROM expense_requests WHERE id=\\$1").
		WithArgs(reqID).
		WillReturnRows(pgxmock.NewRows([]string{"employee_id", "status", "amount"}).
			AddRow(int64(1), "APPROVED", 500.0))

	req, err := repo.GetByID(context.Background(), mockTx, reqID)
	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, 500.0, req.Amount)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestExpenseRequestRepository_GetByID_NotFound(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)

	repo := repositories.NewExpenseRequestRepository(nil)
	mockTx.ExpectQuery("SELECT").WithArgs(int64(1)).WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(context.Background(), mockTx, 1)
	assert.Equal(t, apperrors.ErrExpenseRequestNotFound, err)
}

func TestExpenseRequestRepository_UpdateStatus_Success(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)

	repo := repositories.NewExpenseRequestRepository(nil)
	mockTx.ExpectExec("UPDATE expense_requests").
		WithArgs("APPROVED", int64(2), "Good", int64(123)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateStatus(context.Background(), mockTx, 123, "APPROVED", 2, "Good")
	assert.NoError(t, err)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestExpenseRequestRepository_Cancel_Success(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)

	repo := repositories.NewExpenseRequestRepository(nil)
	mockTx.ExpectExec(`UPDATE expense_requests SET status='CANCELLED' WHERE id=\$1`).
		WithArgs(int64(123)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.Cancel(context.Background(), mockTx, 123)
	assert.NoError(t, err)
}
