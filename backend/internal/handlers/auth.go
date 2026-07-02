package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"expense-tracker/backend/internal/auth"
	appmw "expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/models"
	"expense-tracker/backend/internal/repository"
	"expense-tracker/backend/internal/validation"
)

const refreshCookieName = "refresh_token"

// AuthHandler handles registration, login, token refresh, and logout.
type AuthHandler struct {
	users        *repository.UserRepository
	categories   *repository.CategoryRepository
	tokens       *repository.TokenRepository
	tm           *auth.TokenManager
	adminEmail   string
	cookieSecure bool
	loginLimiter *appmw.RateLimiter // per-email brute-force throttle
	dummyHash    string             // for constant-time login on unknown users
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(u *repository.UserRepository, c *repository.CategoryRepository, t *repository.TokenRepository, tm *auth.TokenManager, adminEmail string, cookieSecure bool) *AuthHandler {
	// Pre-compute a bcrypt hash so login spends the same time whether or not
	// the account exists — closing a user-enumeration timing side channel.
	dummy, _ := auth.HashPassword("timing-equalizer-not-a-real-password")
	return &AuthHandler{
		users: u, categories: c, tokens: t, tm: tm,
		adminEmail: adminEmail, cookieSecure: cookieSecure, dummyHash: dummy,
	}
}

// SetLoginLimiter attaches the per-email login throttle.
func (h *AuthHandler) SetLoginLimiter(l *appmw.RateLimiter) { h.loginLimiter = l }

type credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type authResponse struct {
	AccessToken string      `json:"accessToken"`
	ExpiresIn   int         `json:"expiresIn"` // seconds
	User        userPayload `json:"user"`
}

type userPayload struct {
	ID            string            `json:"id"`
	Email         string            `json:"email"`
	Role          models.Role       `json:"role"`
	Status        models.UserStatus `json:"status"`
	MonthlyIncome string            `json:"monthlyIncome"`
}

func toUserPayload(u *models.User) userPayload {
	return userPayload{ID: u.ID, Email: u.Email, Role: u.Role, Status: u.Status, MonthlyIncome: u.MonthlyIncome.String()}
}

// commonPasswords is a small blocklist of the most-guessed passwords. NIST
// SP 800-63B recommends screening against known-common values rather than
// imposing composition rules.
var commonPasswords = map[string]bool{
	"password": true, "password1": true, "password123": true, "12345678": true,
	"123456789": true, "1234567890": true, "qwerty123": true, "qwertyui": true,
	"11111111": true, "12341234": true, "letmein1": true, "iloveyou": true,
	"admin123": true, "welcome1": true, "abc12345": true, "87654321": true,
	"changeme": true, "passw0rd": true,
}

// passwordPolicyError validates a password at registration. Returns (msg, ok);
// ok=false means rejected with the given user-facing message.
func passwordPolicyError(email, password string) (string, bool) {
	if len(password) < 8 {
		return "password must be at least 8 characters", false
	}
	if len(password) > 72 {
		return "password must be at most 72 characters", false
	}
	lower := strings.ToLower(password)
	if commonPasswords[lower] {
		return "that password is too common — please choose a stronger one", false
	}
	// Password must not embed the email's local part (e.g. "juan" in juan@x.com).
	local := strings.ToLower(strings.TrimSpace(email))
	if i := strings.Index(local, "@"); i > 0 {
		local = local[:i]
	}
	if len(local) >= 4 && strings.Contains(lower, local) {
		return "password must not contain your email name", false
	}
	return "", true
}

// statusRejection returns a client-facing message if the account cannot log in.
func statusRejection(status models.UserStatus) (string, bool) {
	switch status {
	case models.StatusSuspended:
		return "Your account has been suspended. Please contact support.", true
	case models.StatusBanned:
		return "Your account has been banned.", true
	default:
		return "", false
	}
}

// Register creates a member account (open self-registration). The account
// matching the configured ADMIN_EMAIL is promoted to admin.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validation.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if msg, ok := passwordPolicyError(body.Email, body.Password); !ok {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not process password")
		return
	}

	email := strings.ToLower(strings.TrimSpace(body.Email))
	role := models.RoleMember
	if h.adminEmail != "" && email == h.adminEmail {
		role = models.RoleAdmin
	}

	user := &models.User{Email: email, PasswordHash: hash, Role: role}
	if err := h.users.Create(user); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			writeError(w, http.StatusConflict, "an account with that email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}

	// Seed starter categories; non-fatal on failure.
	_ = h.categories.SeedDefaults(user.ID)

	h.issueTokens(w, user, http.StatusCreated)
}

// Login authenticates a user and issues tokens.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.ToLower(strings.TrimSpace(body.Email))

	// Per-account throttle: slows credential-stuffing against one account even
	// from many IPs (the IP-based limiter covers the other direction).
	if h.loginLimiter != nil && email != "" && !h.loginLimiter.Allow(email) {
		writeError(w, http.StatusTooManyRequests, "too many login attempts, please wait a minute")
		return
	}

	user, err := h.users.FindByEmail(email)
	if err != nil {
		// Spend the same time as a real bcrypt compare so response timing does
		// not reveal whether the account exists.
		auth.CheckPassword(h.dummyHash, body.Password)
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.CheckPassword(user.PasswordHash, body.Password) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if msg, blocked := statusRejection(user.Status); blocked {
		writeError(w, http.StatusForbidden, msg)
		return
	}

	h.issueTokens(w, user, http.StatusOK)
}

// Refresh rotates the refresh token and issues a new access token.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	hash := auth.HashRefreshToken(cookie.Value)
	stored, err := h.tokens.FindValid(hash)
	if err != nil {
		h.clearRefreshCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	// Rotate: revoke the old token before issuing a new one.
	_ = h.tokens.Revoke(hash)

	user, err := h.users.FindByID(stored.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "account no longer exists")
		return
	}
	if _, blocked := statusRejection(user.Status); blocked {
		h.clearRefreshCookie(w)
		writeError(w, http.StatusForbidden, "account is not active")
		return
	}

	h.issueTokens(w, user, http.StatusOK)
}

// Logout revokes the current refresh token and clears the cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(refreshCookieName); err == nil && cookie.Value != "" {
		_ = h.tokens.Revoke(auth.HashRefreshToken(cookie.Value))
	}
	h.clearRefreshCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// issueTokens generates an access token + refresh token, stores the refresh
// hash, sets the cookie, and writes the auth response.
func (h *AuthHandler) issueTokens(w http.ResponseWriter, user *models.User, status int) {
	access, err := h.tm.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	plaintext, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	expiresAt := time.Now().Add(h.tm.RefreshTTL())
	if err := h.tokens.Store(user.ID, hash, expiresAt); err != nil {
		writeError(w, http.StatusInternalServerError, "could not persist session")
		return
	}

	h.setRefreshCookie(w, plaintext, expiresAt)
	writeJSON(w, status, authResponse{
		AccessToken: access,
		ExpiresIn:   int(h.tm.AccessTTL().Seconds()),
		User:        toUserPayload(user),
	})
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    value,
		Path:     "/api/auth",
		Expires:  expires,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}
