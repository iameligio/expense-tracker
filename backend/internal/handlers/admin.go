package handlers

import (
	"net/http"
	"strings"

	"expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/models"
	"expense-tracker/backend/internal/repository"

	"github.com/shopspring/decimal"
)

// AdminHandler serves admin-only endpoints (users list, savings-target policy).
type AdminHandler struct {
	users    *repository.UserRepository
	settings *repository.SettingRepository
}

// NewAdminHandler builds an AdminHandler.
func NewAdminHandler(u *repository.UserRepository, s *repository.SettingRepository) *AdminHandler {
	return &AdminHandler{users: u, settings: s}
}

// ListUsers returns all users (admin only).
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load users")
		return
	}
	payload := make([]userPayload, 0, len(users))
	for i := range users {
		payload = append(payload, toUserPayload(&users[i]))
	}
	writeJSON(w, http.StatusOK, payload)
}

// GetSettings returns the global app setting.
func (h *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.settings.Get()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load settings")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

type settingsRequest struct {
	SavingsTargetType  string `json:"savingsTargetType"`
	SavingsTargetValue string `json:"savingsTargetValue"`
}

// UpdateSettings changes the global savings-target policy (admin only).
func (h *AdminHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.UserIDFromContext(r.Context())
	var body settingsRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	targetType := models.SavingsTargetType(strings.ToLower(strings.TrimSpace(body.SavingsTargetType)))
	if targetType != models.SavingsPercent && targetType != models.SavingsFixed {
		writeError(w, http.StatusBadRequest, "savingsTargetType must be 'percent' or 'fixed'")
		return
	}
	value, err := decimal.NewFromString(strings.TrimSpace(body.SavingsTargetValue))
	if err != nil || value.IsNegative() {
		writeError(w, http.StatusBadRequest, "savingsTargetValue must be a non-negative number")
		return
	}
	if targetType == models.SavingsPercent && value.GreaterThan(decimal.NewFromInt(100)) {
		writeError(w, http.StatusBadRequest, "percent target cannot exceed 100")
		return
	}

	updated, err := h.settings.Update(targetType, value, adminID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update settings")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
