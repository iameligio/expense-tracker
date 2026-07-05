package handlers

import (
	"net/http"
	"time"

	"expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/services"
)

// maxTrendMonths bounds a trend query so a custom range can't fan out into an
// unbounded number of per-month sums.
const maxTrendMonths = 24

// TrendsHandler serves the multi-month income-vs-expenses series.
type TrendsHandler struct {
	svc *services.DashboardService
}

// NewTrendsHandler builds a TrendsHandler.
func NewTrendsHandler(s *services.DashboardService) *TrendsHandler {
	return &TrendsHandler{svc: s}
}

// parseMonthStart parses "YYYY-MM" into the first day of that month (Local).
func parseMonthStart(s string) (time.Time, bool) {
	t, err := time.ParseInLocation("2006-01", s, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local), true
}

// monthSpan returns the inclusive number of months from a to b.
func monthSpan(a, b time.Time) int {
	return (int(b.Year())*12 + int(b.Month())) - (int(a.Year())*12 + int(a.Month())) + 1
}

// Get returns per-month income/expense/net for ?from=YYYY-MM&to=YYYY-MM.
// Defaults: `to` = current month, `from` = five months earlier (a 6-month view).
func (h *TrendsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	now := time.Now()
	current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	to := current
	if q := r.URL.Query().Get("to"); q != "" {
		parsed, ok := parseMonthStart(q)
		if !ok {
			writeError(w, http.StatusBadRequest, "to must be in YYYY-MM format")
			return
		}
		to = parsed
	}

	from := to.AddDate(0, -5, 0) // default: 6-month window ending at `to`
	if q := r.URL.Query().Get("from"); q != "" {
		parsed, ok := parseMonthStart(q)
		if !ok {
			writeError(w, http.StatusBadRequest, "from must be in YYYY-MM format")
			return
		}
		from = parsed
	}

	if from.After(to) {
		writeError(w, http.StatusBadRequest, "from must not be after to")
		return
	}
	if monthSpan(from, to) > maxTrendMonths {
		writeError(w, http.StatusBadRequest, "range too large (max 24 months)")
		return
	}

	trend, err := h.svc.BuildTrend(userID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not build trend")
		return
	}
	writeJSON(w, http.StatusOK, trend)
}
