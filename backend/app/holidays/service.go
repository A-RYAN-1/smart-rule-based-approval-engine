package holidays

import (
	"context"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/constants"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/apperrors"
)

type HolidayService struct {
	holidayRepo interfaces.HolidayRepository
}

func NewHolidayService(ctx context.Context, holidayRepo interfaces.HolidayRepository) interfaces.HolidayService {
	return &HolidayService{holidayRepo: holidayRepo}
}

func (s *HolidayService) ensureAdmin(role string) error {
	if role != constants.RoleAdmin {
		return apperrors.ErrAdminOnly
	}
	return nil
}

func (s *HolidayService) AddHoliday(ctx context.Context, role string, adminID int64, date time.Time, desc string) error {
	if err := s.ensureAdmin(role); err != nil {
		return err
	}
	return s.holidayRepo.AddHoliday(ctx, date, desc, adminID)
}

func (s *HolidayService) GetHolidays(ctx context.Context, role string) ([]map[string]interface{}, error) {
	if err := s.ensureAdmin(role); err != nil {
		return nil, err
	}
	return s.holidayRepo.GetHolidays(ctx)
}

func (s *HolidayService) DeleteHoliday(ctx context.Context, role string, holidayID int64) error {
	if err := s.ensureAdmin(role); err != nil {
		return err
	}
	return s.holidayRepo.DeleteHoliday(ctx, holidayID)
}
