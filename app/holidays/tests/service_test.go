package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/app/holidays"
	"github.com/ankita-advitot/rule_based_approval_engine/app/holidays/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHolidayService(t *testing.T) {
	ctx := context.Background()

	t.Run("AddHoliday - Repository Error", func(t *testing.T) {
		mockRepo := mocks.NewHolidayRepository(t)
		mockRepo.EXPECT().AddHoliday(ctx, mock.Anything, "Fail", int64(1)).Return(assert.AnError)

		service := holidays.NewHolidayService(ctx, mockRepo)
		err := service.AddHoliday(ctx, "ADMIN", 1, time.Now(), "Fail")

		assert.Error(t, err)
	})

	t.Run("GetHolidays - Repository Error", func(t *testing.T) {
		mockRepo := mocks.NewHolidayRepository(t)
		mockRepo.EXPECT().GetHolidays(ctx).Return(nil, assert.AnError)

		service := holidays.NewHolidayService(ctx, mockRepo)
		_, err := service.GetHolidays(ctx, "ADMIN")

		assert.Error(t, err)
	})

	t.Run("DeleteHoliday - Repository Error", func(t *testing.T) {
		mockRepo := mocks.NewHolidayRepository(t)
		mockRepo.EXPECT().DeleteHoliday(ctx, int64(99)).Return(assert.AnError)

		service := holidays.NewHolidayService(ctx, mockRepo)
		err := service.DeleteHoliday(ctx, "ADMIN", 99)

		assert.Error(t, err)
	})
}
