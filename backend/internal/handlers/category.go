package handlers

import (
	"errors"
	"net/http"
	"strings"

	"expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/models"
	"expense-tracker/backend/internal/repository"

	"github.com/go-chi/chi/v5"
)

// CategoryHandler handles CRUD for the authenticated user's categories.
type CategoryHandler struct {
	categories *repository.CategoryRepository
}

// NewCategoryHandler builds a CategoryHandler.
func NewCategoryHandler(c *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categories: c}
}

type categoryRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (b categoryRequest) normalized() (string, models.CategoryType, bool) {
	name := strings.TrimSpace(b.Name)
	catType := models.CategoryType(strings.ToLower(strings.TrimSpace(b.Type)))
	if name == "" || len(name) > 100 || !models.ValidCategoryType(catType) {
		return "", "", false
	}
	return name, catType, true
}

// List returns the user's categories.
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	cats, err := h.categories.ListByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load categories")
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

// Create adds a category.
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	var body categoryRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	name, catType, ok := body.normalized()
	if !ok {
		writeError(w, http.StatusBadRequest, "name is required (max 100 chars) and type must be one of: fixed, variable, wants, debts")
		return
	}
	cat := &models.Category{UserID: userID, Name: name, Type: catType}
	if err := h.categories.Create(cat); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			writeError(w, http.StatusConflict, "you already have a category with that name")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create category")
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

// Update edits an owned category.
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var body categoryRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	name, catType, ok := body.normalized()
	if !ok {
		writeError(w, http.StatusBadRequest, "name is required (max 100 chars) and type must be one of: fixed, variable, wants, debts")
		return
	}
	if err := h.categories.Update(userID, id, name, catType); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			writeError(w, http.StatusNotFound, "category not found")
		case errors.Is(err, repository.ErrDuplicate):
			writeError(w, http.StatusConflict, "you already have a category with that name")
		default:
			writeError(w, http.StatusInternalServerError, "could not update category")
		}
		return
	}
	cat, err := h.categories.FindOwned(userID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load updated category")
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

// Delete removes an owned category (fails if expenses reference it).
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.categories.Delete(userID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "category not found")
			return
		}
		writeError(w, http.StatusConflict, "cannot delete a category that still has expenses")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
