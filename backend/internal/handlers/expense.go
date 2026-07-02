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

// ExpenseHandler handles CRUD for the authenticated user's expenses.
type ExpenseHandler struct {
	expenses   *repository.ExpenseRepository
	categories *repository.CategoryRepository
}

// NewExpenseHandler builds an ExpenseHandler.
func NewExpenseHandler(e *repository.ExpenseRepository, c *repository.CategoryRepository) *ExpenseHandler {
	return &ExpenseHandler{expenses: e, categories: c}
}

type expenseRequest struct {
	Amount     string  `json:"amount"`
	Note       *string `json:"note"`
	SpentOn    string  `json:"spentOn"` // YYYY-MM-DD
	CategoryID string  `json:"categoryId"`
}

// List returns the user's expenses for a month (?month=YYYY-MM, default current).
func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	_, from, to, ok := monthRange(r.URL.Query().Get("month"))
	if !ok {
		writeError(w, http.StatusBadRequest, "month must be in YYYY-MM format")
		return
	}
	expenses, err := h.expenses.ListByUserMonth(userID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load expenses")
		return
	}
	writeJSON(w, http.StatusOK, expenses)
}

// parse validates an expense request into concrete values.
func (h *ExpenseHandler) parse(userID string, body expenseRequest) (decimal.Decimal, time.Time, string, *string, error) {
	amount, err := decimal.NewFromString(strings.TrimSpace(body.Amount))
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, time.Time{}, "", nil, errors.New("amount must be a number greater than 0")
	}
	if amount.GreaterThan(decimal.NewFromInt(1_000_000_000)) {
		return decimal.Zero, time.Time{}, "", nil, errors.New("amount is unreasonably large")
	}
	spentOn, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(body.SpentOn), time.Local)
	if err != nil {
		return decimal.Zero, time.Time{}, "", nil, errors.New("spentOn must be a valid date (YYYY-MM-DD)")
	}
	catID := strings.TrimSpace(body.CategoryID)
	if catID == "" {
		return decimal.Zero, time.Time{}, "", nil, errors.New("categoryId is required")
	}
	// Ownership check: the category must belong to this user.
	if _, err := h.categories.FindOwned(userID, catID); err != nil {
		return decimal.Zero, time.Time{}, "", nil, errors.New("category not found")
	}
	var note *string
	if body.Note != nil {
		trimmed := strings.TrimSpace(*body.Note)
		if trimmed != "" {
			note = &trimmed
		}
	}
	return amount, spentOn, catID, note, nil
}

// Create adds an expense.
func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	var body expenseRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	amount, spentOn, catID, note, err := h.parse(userID, body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	exp := &models.Expense{
		UserID: userID, Amount: amount, SpentOn: spentOn, CategoryID: catID, Note: note,
	}
	if err := h.expenses.Create(exp); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create expense")
		return
	}
	created, err := h.expenses.FindOwned(userID, exp.ID)
	if err != nil {
		writeJSON(w, http.StatusCreated, exp)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Update edits an owned expense.
func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var body expenseRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	amount, spentOn, catID, note, err := h.parse(userID, body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]interface{}{
		"amount": amount, "spent_on": spentOn, "category_id": catID, "note": note,
	}
	if err := h.expenses.Update(userID, id, fields); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update expense")
		return
	}
	updated, err := h.expenses.FindOwned(userID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load updated expense")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete removes an owned expense.
func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.expenses.Delete(userID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete expense")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
