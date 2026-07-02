package repository

import (
	"errors"

	"expense-tracker/backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SettingRepository provides access to the global AppSetting singleton.
type SettingRepository struct{ db *gorm.DB }

// NewSettingRepository builds a SettingRepository.
func NewSettingRepository(db *gorm.DB) *SettingRepository { return &SettingRepository{db: db} }

// Get returns the global app setting.
func (r *SettingRepository) Get() (*models.AppSetting, error) {
	var s models.AppSetting
	err := r.db.First(&s, "id = ?", models.GlobalSettingID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

// Update saves the savings-target policy and records who changed it.
func (r *SettingRepository) Update(targetType models.SavingsTargetType, value decimal.Decimal, updatedBy string) (*models.AppSetting, error) {
	err := r.db.Model(&models.AppSetting{}).
		Where("id = ?", models.GlobalSettingID).
		Updates(map[string]interface{}{
			"savings_target_type":  targetType,
			"savings_target_value": value,
			"updated_by":           updatedBy,
		}).Error
	if err != nil {
		return nil, err
	}
	return r.Get()
}
