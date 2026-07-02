package repository

import (
	"errors"
	"strings"

	"expense-tracker/backend/internal/models"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a record does not exist (or is not owned by the caller).
var ErrNotFound = errors.New("record not found")

// ErrDuplicate is returned on a unique-constraint violation.
var ErrDuplicate = errors.New("record already exists")

// UserRepository provides data access for users.
type UserRepository struct{ db *gorm.DB }

// NewUserRepository builds a UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

// Create inserts a new user. Returns ErrDuplicate if the email is taken.
func (r *UserRepository) Create(u *models.User) error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if err := r.db.Create(u).Error; err != nil {
		if isDuplicate(err) {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

// FindByEmail looks up a user by email (case-insensitive).
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.Where("email = ?", strings.ToLower(strings.TrimSpace(email))).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

// FindByID looks up a user by id.
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var u models.User
	err := r.db.First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

// UpdateIncome sets a user's monthly income.
func (r *UserRepository) UpdateIncome(userID string, income interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("monthly_income", income).Error
}

// UpdateStatus sets a user's moderation status (active/suspended/banned).
func (r *UserRepository) UpdateStatus(userID string, status models.UserStatus) error {
	res := r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// IsActive reports whether the user exists and is in the active status.
func (r *UserRepository) IsActive(userID string) (bool, error) {
	var status models.UserStatus
	err := r.db.Model(&models.User{}).Select("status").
		Where("id = ?", userID).Scan(&status).Error
	if err != nil {
		return false, err
	}
	return status == models.StatusActive, nil
}

// Delete hard-deletes a user and all their data in a single transaction.
// Expenses are removed first (their FK to categories is RESTRICT), then
// refresh tokens and categories, then the user row itself.
func (r *UserRepository) Delete(userID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.Expense{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.Category{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", userID).Delete(&models.User{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// List returns all users (admin use).
func (r *UserRepository) List() ([]models.User, error) {
	var users []models.User
	err := r.db.Order("created_at ASC").Find(&users).Error
	return users, err
}

// Count returns the total number of users.
func (r *UserRepository) Count() (int64, error) {
	var n int64
	err := r.db.Model(&models.User{}).Count(&n).Error
	return n, err
}

func isDuplicate(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
