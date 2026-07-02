package repository

import (
	"errors"
	"time"

	"expense-tracker/backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ExpenseRepository provides ownership-scoped data access for expenses.
type ExpenseRepository struct{ db *gorm.DB }

// NewExpenseRepository builds an ExpenseRepository.
func NewExpenseRepository(db *gorm.DB) *ExpenseRepository { return &ExpenseRepository{db: db} }

// Create inserts an expense.
func (r *ExpenseRepository) Create(e *models.Expense) error {
	return r.db.Create(e).Error
}

// ListByUserMonth returns a user's expenses within [from, to) with categories loaded.
func (r *ExpenseRepository) ListByUserMonth(userID string, from, to time.Time) ([]models.Expense, error) {
	var expenses []models.Expense
	err := r.db.Preload("Category").
		Where("user_id = ? AND spent_on >= ? AND spent_on < ?", userID, from, to).
		Order("spent_on DESC, created_at DESC").
		Find(&expenses).Error
	return expenses, err
}

// FindOwned returns an expense only if it belongs to the user.
func (r *ExpenseRepository) FindOwned(userID, id string) (*models.Expense, error) {
	var e models.Expense
	err := r.db.Preload("Category").
		Where("id = ? AND user_id = ?", id, userID).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &e, err
}

// Update saves changes to an owned expense.
func (r *ExpenseRepository) Update(userID, id string, fields map[string]interface{}) error {
	res := r.db.Model(&models.Expense{}).
		Where("id = ? AND user_id = ?", id, userID).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes an owned expense.
func (r *ExpenseRepository) Delete(userID, id string) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Expense{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CategoryBreakdownRow is one grouped total for the pie chart.
type CategoryBreakdownRow struct {
	CategoryID string
	Name       string
	Type       models.CategoryType
	Total      decimal.Decimal
}

// CategoryBreakdown returns per-category spend totals for a user within a month.
func (r *ExpenseRepository) CategoryBreakdown(userID string, from, to time.Time) ([]CategoryBreakdownRow, error) {
	var rows []CategoryBreakdownRow
	err := r.db.Model(&models.Expense{}).
		Select("categories.id AS category_id, categories.name AS name, categories.type AS type, SUM(expenses.amount) AS total").
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Where("expenses.user_id = ? AND expenses.spent_on >= ? AND expenses.spent_on < ?", userID, from, to).
		Group("categories.id, categories.name, categories.type").
		Order("total DESC").
		Scan(&rows).Error
	return rows, err
}

// TotalForUserMonth returns the sum of a user's expenses within a month.
func (r *ExpenseRepository) TotalForUserMonth(userID string, from, to time.Time) (decimal.Decimal, error) {
	var result struct{ Total decimal.Decimal }
	err := r.db.Model(&models.Expense{}).
		Select("COALESCE(SUM(amount), 0) AS total").
		Where("user_id = ? AND spent_on >= ? AND spent_on < ?", userID, from, to).
		Scan(&result).Error
	return result.Total, err
}
