package handlers

import (
	"errors"
	"net/http"
	"strings"

	"expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/models"
	"expense-tracker/backend/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

// AdminHandler serves admin-only endpoints (user moderation, savings policy).
type AdminHandler struct {
	users    *repository.UserRepository
	settings *repository.SettingRepository
	tokens   *repository.TokenRepository
}

// NewAdminHandler builds an AdminHandler.
func NewAdminHandler(u *repository.UserRepository, s *repository.SettingRepository, t *repository.TokenRepository) *AdminHandler {
	return &AdminHandler{users: u, settings: s, tokens: t}
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

type statusRequest struct {
	Status string `json:"status"`
}

// UpdateUserStatus suspends, bans, or reactivates a user (admin only).
// Suspending/banning also revokes the target's active sessions.
func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.UserIDFromContext(r.Context())
	targetID := chi.URLParam(r, "id")
	if targetID == adminID {
		writeError(w, http.StatusBadRequest, "you cannot change your own account status")
		return
	}

	var body statusRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	status := models.UserStatus(strings.ToLower(strings.TrimSpace(body.Status)))
	if !models.ValidUserStatus(status) {
		writeError(w, http.StatusBadRequest, "status must be one of: active, suspended, banned")
		return
	}

	if err := h.users.UpdateStatus(targetID, status); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update user status")
		return
	}

	// Kill active sessions when blocking the account.
	if status != models.StatusActive {
		_ = h.tokens.RevokeAllForUser(targetID)
	}

	user, err := h.users.FindByID(targetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load updated user")
		return
	}
	writeJSON(w, http.StatusOK, toUserPayload(user))
}

// DeleteUser permanently removes a user and all their data (admin only).
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.UserIDFromContext(r.Context())
	targetID := chi.URLParam(r, "id")
	if targetID == adminID {
		writeError(w, http.StatusBadRequest, "you cannot delete your own account")
		return
	}

	if err := h.users.Delete(targetID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete user")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
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
