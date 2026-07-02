package handlers

import "time"

// monthRange parses a "YYYY-MM" string (defaulting to the current month when
// empty) and returns the normalized month string plus [from, to) bounds.
func monthRange(month string) (normalized string, from, to time.Time, ok bool) {
	now := time.Now()
	if month == "" {
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	} else {
		t, err := time.ParseInLocation("2006-01", month, time.Local)
		if err != nil {
			return "", time.Time{}, time.Time{}, false
		}
		from = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	}
	to = from.AddDate(0, 1, 0)
	return from.Format("2006-01"), from, to, true
}
