package repository

import (
	"errors"

	"expense-tracker/backend/internal/models"

	"gorm.io/gorm"
)

// CategoryRepository provides ownership-scoped data access for categories.
type CategoryRepository struct{ db *gorm.DB }

// NewCategoryRepository builds a CategoryRepository.
func NewCategoryRepository(db *gorm.DB) *CategoryRepository { return &CategoryRepository{db: db} }

// Create inserts a category. Returns ErrDuplicate if the user already has one
// with the same name.
func (r *CategoryRepository) Create(c *models.Category) error {
	if err := r.db.Create(c).Error; err != nil {
		if isDuplicate(err) {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

// ListByUser returns all categories owned by the user.
func (r *CategoryRepository) ListByUser(userID string) ([]models.Category, error) {
	var cats []models.Category
	err := r.db.Where("user_id = ?", userID).Order("type ASC, name ASC").Find(&cats).Error
	return cats, err
}

// FindOwned returns a category only if it belongs to the user.
func (r *CategoryRepository) FindOwned(userID, id string) (*models.Category, error) {
	var c models.Category
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

// Update saves name/type changes for an owned category.
func (r *CategoryRepository) Update(userID, id, name string, catType models.CategoryType) error {
	res := r.db.Model(&models.Category{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{"name": name, "type": catType})
	if res.Error != nil {
		if isDuplicate(res.Error) {
			return ErrDuplicate
		}
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes an owned category. Fails if expenses still reference it
// (FK RESTRICT) — surfaced to the caller as an error.
func (r *CategoryRepository) Delete(userID, id string) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Category{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountForUser returns how many categories a user has.
func (r *CategoryRepository) CountForUser(userID string) (int64, error) {
	var n int64
	err := r.db.Model(&models.Category{}).Where("user_id = ?", userID).Count(&n).Error
	return n, err
}

// SeedDefaults creates a starter set of categories for a new user.
func (r *CategoryRepository) SeedDefaults(userID string) error {
	defaults := []models.Category{
		{UserID: userID, Name: "Rent", Type: models.CategoryFixed},
		{UserID: userID, Name: "Utilities", Type: models.CategoryFixed},
		{UserID: userID, Name: "Groceries", Type: models.CategoryVariable},
		{UserID: userID, Name: "Transportation", Type: models.CategoryVariable},
		{UserID: userID, Name: "Dining Out", Type: models.CategoryWants},
		{UserID: userID, Name: "Entertainment", Type: models.CategoryWants},
		{UserID: userID, Name: "Loan Payment", Type: models.CategoryDebts},
	}
	return r.db.Create(&defaults).Error
}
