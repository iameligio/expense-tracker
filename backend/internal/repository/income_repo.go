package repository

import (
	"errors"
	"time"

	"expense-tracker/backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// IncomeRepository provides ownership-scoped data access for income entries.
type IncomeRepository struct{ db *gorm.DB }

// NewIncomeRepository builds an IncomeRepository.
func NewIncomeRepository(db *gorm.DB) *IncomeRepository { return &IncomeRepository{db: db} }

// Create inserts an income entry.
func (r *IncomeRepository) Create(i *models.Income) error {
	return r.db.Create(i).Error
}

// ListByUserMonth returns a user's income entries within [from, to).
func (r *IncomeRepository) ListByUserMonth(userID string, from, to time.Time) ([]models.Income, error) {
	var incomes []models.Income
	err := r.db.
		Where("user_id = ? AND received_on >= ? AND received_on < ?", userID, from, to).
		Order("received_on DESC, created_at DESC").
		Find(&incomes).Error
	return incomes, err
}

// FindOwned returns an income entry only if it belongs to the user.
func (r *IncomeRepository) FindOwned(userID, id string) (*models.Income, error) {
	var i models.Income
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&i).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &i, err
}

// Update saves changes to an owned income entry.
func (r *IncomeRepository) Update(userID, id string, fields map[string]interface{}) error {
	res := r.db.Model(&models.Income{}).
		Where("id = ? AND user_id = ?", id, userID).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes an owned income entry.
func (r *IncomeRepository) Delete(userID, id string) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Income{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// SourceBreakdownRow is one grouped total for the income pie chart.
type SourceBreakdownRow struct {
	Source models.IncomeSource
	Total  decimal.Decimal
}

// SourceBreakdown returns per-source income totals for a user within a month.
func (r *IncomeRepository) SourceBreakdown(userID string, from, to time.Time) ([]SourceBreakdownRow, error) {
	var rows []SourceBreakdownRow
	err := r.db.Model(&models.Income{}).
		Select("source, SUM(amount) AS total").
		Where("user_id = ? AND received_on >= ? AND received_on < ?", userID, from, to).
		Group("source").
		Order("total DESC").
		Scan(&rows).Error
	return rows, err
}

// TotalForUserMonth returns the sum of a user's income within a month.
func (r *IncomeRepository) TotalForUserMonth(userID string, from, to time.Time) (decimal.Decimal, error) {
	var result struct{ Total decimal.Decimal }
	err := r.db.Model(&models.Income{}).
		Select("COALESCE(SUM(amount), 0) AS total").
		Where("user_id = ? AND received_on >= ? AND received_on < ?", userID, from, to).
		Scan(&result).Error
	return result.Total, err
}
