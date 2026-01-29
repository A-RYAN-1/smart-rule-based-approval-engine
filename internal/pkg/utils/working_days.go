package utils

import (
	"time"
)

// CountWorkingDays counts business days between two dates, skipping weekends and holidays
func CountWorkingDays(from, to time.Time, isHoliday func(time.Time) bool) int {
	days := 0

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		if isHoliday != nil && isHoliday(d) {
			continue
		}
		days++
	}

	return days
}
