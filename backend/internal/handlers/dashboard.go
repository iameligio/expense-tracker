package handlers

import (
	"net/http"

	"expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/services"
)

// DashboardHandler serves the aggregated dashboard payload.
type DashboardHandler struct {
	svc *services.DashboardService
}

// NewDashboardHandler builds a DashboardHandler.
func NewDashboardHandler(s *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: s}
}

// Get returns KPIs + category/type breakdowns for ?month=YYYY-MM (default current).
func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	month, from, to, ok := monthRange(r.URL.Query().Get("month"))
	if !ok {
		writeError(w, http.StatusBadRequest, "month must be in YYYY-MM format")
		return
	}
	dash, err := h.svc.Build(userID, month, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not build dashboard")
		return
	}
	writeJSON(w, http.StatusOK, dash)
}
