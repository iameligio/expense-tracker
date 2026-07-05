package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/models"
	"expense-tracker/backend/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

// IncomeHandler handles CRUD for the authenticated user's income entries.
type IncomeHandler struct {
	incomes *repository.IncomeRepository
}

// NewIncomeHandler builds an IncomeHandler.
func NewIncomeHandler(i *repository.IncomeRepository) *IncomeHandler {
	return &IncomeHandler{incomes: i}
}

type incomeRequest struct {
	Amount     string  `json:"amount"`
	Note       *string `json:"note"`
	ReceivedOn string  `json:"receivedOn"` // YYYY-MM-DD
	Source     string  `json:"source"`
}

// List returns the user's income for a month (?month=YYYY-MM, default current).
func (h *IncomeHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	_, from, to, ok := monthRange(r.URL.Query().Get("month"))
	if !ok {
		writeError(w, http.StatusBadRequest, "month must be in YYYY-MM format")
		return
	}
	incomes, err := h.incomes.ListByUserMonth(userID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load income")
		return
	}
	writeJSON(w, http.StatusOK, incomes)
}

// parse validates an income request into concrete values.
func (h *IncomeHandler) parse(body incomeRequest) (decimal.Decimal, time.Time, models.IncomeSource, *string, error) {
	amount, err := decimal.NewFromString(strings.TrimSpace(body.Amount))
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, time.Time{}, "", nil, errors.New("amount must be a number greater than 0")
	}
	if amount.GreaterThan(decimal.NewFromInt(1_000_000_000)) {
		return decimal.Zero, time.Time{}, "", nil, errors.New("amount is unreasonably large")
	}
	receivedOn, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(body.ReceivedOn), time.Local)
	if err != nil {
		return decimal.Zero, time.Time{}, "", nil, errors.New("receivedOn must be a valid date (YYYY-MM-DD)")
	}
	source := models.IncomeSource(strings.TrimSpace(body.Source))
	if !models.ValidIncomeSource(source) {
		return decimal.Zero, time.Time{}, "", nil, errors.New("source must be one of: salary, side_project, other")
	}
	var note *string
	if body.Note != nil {
		trimmed := strings.TrimSpace(*body.Note)
		if trimmed != "" {
			note = &trimmed
		}
	}
	return amount, receivedOn, source, note, nil
}

// Create adds an income entry.
func (h *IncomeHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	var body incomeRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	amount, receivedOn, source, note, err := h.parse(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	inc := &models.Income{
		UserID: userID, Amount: amount, ReceivedOn: receivedOn, Source: source, Note: note,
	}
	if err := h.incomes.Create(inc); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create income")
		return
	}
	created, err := h.incomes.FindOwned(userID, inc.ID)
	if err != nil {
		writeJSON(w, http.StatusCreated, inc)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Update edits an owned income entry.
func (h *IncomeHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var body incomeRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	amount, receivedOn, source, note, err := h.parse(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]interface{}{
		"amount": amount, "received_on": receivedOn, "source": source, "note": note,
	}
	if err := h.incomes.Update(userID, id, fields); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "income not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update income")
		return
	}
	updated, err := h.incomes.FindOwned(userID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load updated income")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete removes an owned income entry.
func (h *IncomeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.incomes.Delete(userID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "income not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete income")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
