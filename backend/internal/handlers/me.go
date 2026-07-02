package handlers

import (
	"net/http"

	"expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/repository"

	"github.com/shopspring/decimal"
)

// MeHandler serves the authenticated user's own profile.
type MeHandler struct {
	users *repository.UserRepository
}

// NewMeHandler builds a MeHandler.
func NewMeHandler(u *repository.UserRepository) *MeHandler { return &MeHandler{users: u} }

// Get returns the current user's profile.
func (h *MeHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	user, err := h.users.FindByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, toUserPayload(user))
}

type updateMeRequest struct {
	MonthlyIncome string `json:"monthlyIncome"`
}

// UpdateIncome sets the current user's monthly income.
func (h *MeHandler) UpdateIncome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var body updateMeRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	income, err := decimal.NewFromString(body.MonthlyIncome)
	if err != nil || income.IsNegative() {
		writeError(w, http.StatusBadRequest, "monthlyIncome must be a non-negative number")
		return
	}

	if err := h.users.UpdateIncome(userID, income); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update income")
		return
	}
	user, err := h.users.FindByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load updated profile")
		return
	}
	writeJSON(w, http.StatusOK, toUserPayload(user))
}
