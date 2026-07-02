package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"expense-tracker/backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when an access token cannot be validated.
var ErrInvalidToken = errors.New("invalid or expired token")

// Claims are the custom JWT claims embedded in an access token.
type Claims struct {
	Role models.Role `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies JWT access tokens and generates refresh tokens.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenManager builds a TokenManager.
func NewTokenManager(secret []byte, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// AccessTTL exposes the access-token lifetime (used to hint the SPA).
func (m *TokenManager) AccessTTL() time.Duration { return m.accessTTL }

// RefreshTTL exposes the refresh-token lifetime.
func (m *TokenManager) RefreshTTL() time.Duration { return m.refreshTTL }

// GenerateAccessToken signs a short-lived access token for the user.
func (m *TokenManager) GenerateAccessToken(userID string, role models.Role) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseAccessToken validates a token string and returns its claims.
func (m *TokenManager) ParseAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GenerateRefreshToken returns a new random opaque refresh token (plaintext)
// and its SHA-256 hash for storage.
func GenerateRefreshToken() (plaintext, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	plaintext = base64.RawURLEncoding.EncodeToString(buf)
	hash = HashRefreshToken(plaintext)
	return plaintext, hash, nil
}

// HashRefreshToken returns the hex SHA-256 of a refresh token for lookup/storage.
func HashRefreshToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}
