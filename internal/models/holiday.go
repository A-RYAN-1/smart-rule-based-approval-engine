package models

import "time"

type Holiday struct {
	ID          int64
	HolidayDate time.Time
	Description string
	CreatedBy   int64
	CreatedAt   time.Time
}
