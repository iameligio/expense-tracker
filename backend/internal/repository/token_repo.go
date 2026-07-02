package repository

import (
	"errors"
	"time"

	"expense-tracker/backend/internal/models"

	"gorm.io/gorm"
)

// TokenRepository manages refresh tokens for revocable JWT sessions.
type TokenRepository struct{ db *gorm.DB }

// NewTokenRepository builds a TokenRepository.
func NewTokenRepository(db *gorm.DB) *TokenRepository { return &TokenRepository{db: db} }

// Store persists a hashed refresh token.
func (r *TokenRepository) Store(userID, tokenHash string, expiresAt time.Time) error {
	rt := &models.RefreshToken{UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt}
	return r.db.Create(rt).Error
}

// FindValid returns a non-revoked, non-expired token by its hash.
func (r *TokenRepository) FindValid(tokenHash string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Where("token_hash = ? AND revoked = ? AND expires_at > ?", tokenHash, false, time.Now()).
		First(&rt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &rt, err
}

// Revoke marks a single token (by hash) as revoked.
func (r *TokenRepository) Revoke(tokenHash string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("token_hash = ?", tokenHash).Update("revoked", true).Error
}

// RevokeAllForUser revokes every active token for a user.
func (r *TokenRepository) RevokeAllForUser(userID string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).Update("revoked", true).Error
}
