package tests

import (
	"testing"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMiscUtils_ApprovalRules(t *testing.T) {
	t.Run("ValidatePendingStatus", func(t *testing.T) {
		assert.NoError(t, utils.ValidatePendingStatus(constants.StatusPending))
		assert.ErrorIs(t, utils.ValidatePendingStatus(constants.StatusApproved), apperrors.ErrRequestNotPending)
	})

	t.Run("ValidateApproverRole", func(t *testing.T) {
		// Employee cannot approve
		assert.ErrorIs(t, utils.ValidateApproverRole(constants.RoleEmployee, constants.RoleEmployee), apperrors.ErrEmployeeCannotApprove)

		// Admin can approve anyone
		assert.NoError(t, utils.ValidateApproverRole(constants.RoleAdmin, constants.RoleEmployee))
		assert.NoError(t, utils.ValidateApproverRole(constants.RoleAdmin, constants.RoleManager))

		// Manager can approve employee
		assert.NoError(t, utils.ValidateApproverRole(constants.RoleManager, constants.RoleEmployee))

		// Manager cannot approve manager
		assert.ErrorIs(t, utils.ValidateApproverRole(constants.RoleManager, constants.RoleManager), apperrors.ErrManagerNeedsAdmin)

		// Manager cannot approve admin (safety net)
		assert.ErrorIs(t, utils.ValidateApproverRole(constants.RoleManager, constants.RoleAdmin), apperrors.ErrAdminRequestNotAllowed)

		// Invalid role combos
		assert.ErrorIs(t, utils.ValidateApproverRole("INVALID", constants.RoleEmployee), apperrors.ErrUnauthorizedApproval)
	})
}

func TestMiscUtils_DateAndWorkingDays(t *testing.T) {
	t.Run("CalculateLeaveDays", func(t *testing.T) {
		from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, 3, utils.CalculateLeaveDays(from, to))
	})

	t.Run("CountWorkingDays", func(t *testing.T) {
		from := time.Date(2023, 1, 6, 0, 0, 0, 0, time.UTC) // Friday
		to := time.Date(2023, 1, 9, 0, 0, 0, 0, time.UTC)   // Monday
		// Fri, Sat, Sun, Mon -> Working: Fri, Mon = 2

		assert.Equal(t, 2, utils.CountWorkingDays(from, to, nil))

		// With holiday
		isHoliday := func(d time.Time) bool {
			return d.Day() == 9 // Monday is holiday
		}
		assert.Equal(t, 1, utils.CountWorkingDays(from, to, isHoliday))
	})

	t.Run("CalculateLeaveDays - Edge Cases", func(t *testing.T) {
		// Same day
		d := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, 1, utils.CalculateLeaveDays(d, d))

		// Backward dates
		from := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
		to := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, 0, utils.CalculateLeaveDays(from, to))
	})

	t.Run("CountWorkingDays - Long Range", func(t *testing.T) {
		from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)   // Sunday
		to := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC) // 31 days
		// 31 days: 4 weeks (20 working) + 3 days (Sun, Mon, Tue -> 2 working) = 22 working days
		assert.Equal(t, 22, utils.CountWorkingDays(from, to, nil))
	})
}
