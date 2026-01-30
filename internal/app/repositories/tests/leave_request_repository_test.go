// File: internal/app/repositories/tests/leave_request_repository_test.go
package tests

import (
	"context"
	"testing"
	"time"

	"rule-based-approval-engine/internal/app/repositories"
	"rule-based-approval-engine/internal/models"
	"rule-based-approval-engine/internal/pkg/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeaveRequestRepository_Create_Success(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mockTx.Rollback(context.Background())

	repo := repositories.NewLeaveRequestRepository(nil)
	req := &models.LeaveRequest{
		EmployeeID: 1,
		FromDate:   time.Now(),
		ToDate:     time.Now().Add(24 * time.Hour),
		Reason:     "Vacation",
		LeaveType:  "ANNUAL",
		Status:     "PENDING",
	}

	mockTx.ExpectExec("INSERT INTO leave_requests").
		WithArgs(req.EmployeeID, req.FromDate, req.ToDate, req.Reason, req.LeaveType, req.Status, req.RuleID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), mockTx, req)
	assert.NoError(t, err)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestLeaveRequestRepository_GetByID_Success(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mockTx.Rollback(context.Background())

	repo := repositories.NewLeaveRequestRepository(nil)
	reqID := int64(123)

	mockTx.ExpectQuery("SELECT employee_id, status, from_date, to_date FROM leave_requests WHERE id=\\$1").
		WithArgs(reqID).
		WillReturnRows(pgxmock.NewRows([]string{"employee_id", "status", "from_date", "to_date"}).
			AddRow(int64(1), "APPROVED", time.Now(), time.Now()))

	req, err := repo.GetByID(context.Background(), mockTx, reqID)
	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, int64(1), req.EmployeeID)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestLeaveRequestRepository_GetByID_NotFound(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)

	repo := repositories.NewLeaveRequestRepository(nil)
	mockTx.ExpectQuery("SELECT").WithArgs(int64(1)).WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(context.Background(), mockTx, 1)
	assert.Equal(t, apperrors.ErrLeaveRequestNotFound, err)
}

func TestLeaveRequestRepository_UpdateStatus_Success(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)

	repo := repositories.NewLeaveRequestRepository(nil)
	mockTx.ExpectExec("UPDATE leave_requests").
		WithArgs("APPROVED", int64(2), "Good", int64(123)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateStatus(context.Background(), mockTx, 123, "APPROVED", 2, "Good")
	assert.NoError(t, err)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestLeaveRequestRepository_Cancel_Success(t *testing.T) {
	mockTx, err := pgxmock.NewConn()
	require.NoError(t, err)

	repo := repositories.NewLeaveRequestRepository(nil)
	mockTx.ExpectExec(`UPDATE leave_requests SET status='CANCELLED' WHERE id=\$1`).
		WithArgs(int64(123)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.Cancel(context.Background(), mockTx, 123)
	assert.NoError(t, err)
}
