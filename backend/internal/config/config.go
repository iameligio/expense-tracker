package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	Port         string
	CORSOrigins  []string
	IsProduction bool
	CookieSecure bool

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret       []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	AdminEmail string
}

// Load reads configuration from an environment-specific .env file and the
// process environment. APP_ENV selects the file: "local" (default) loads
// .env.local, "production" loads .env.production; a bare .env is used as a
// shared fallback. Real process env vars always win (nothing is overridden).
func Load() (*Config, error) {
	appEnv := getEnv("APP_ENV", "local")
	// godotenv.Load does not override already-set vars, and earlier files win
	// over later ones — so .env.<APP_ENV> takes precedence over the shared .env.
	_ = godotenv.Load(".env."+appEnv, ".env")

	accessMin := getIntEnv("ACCESS_TOKEN_TTL_MINUTES", 15)
	refreshHrs := getIntEnv("REFRESH_TOKEN_TTL_HOURS", 168)

	secret := getEnv("JWT_SECRET", "")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	isProd := strings.EqualFold(getEnv("ENV", "development"), "production")

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		CORSOrigins:     splitAndTrim(getEnv("CORS_ORIGINS", "http://localhost:5173")),
		IsProduction:    isProd,
		// Secure cookies default to on in production, but can be forced on in
		// dev when served over TLS (e.g. behind Valet at https://*.test).
		CookieSecure:    getBoolEnv("COOKIE_SECURE", isProd),
		DBHost:          getEnv("DB_HOST", "127.0.0.1"),
		DBPort:          getEnv("DB_PORT", "3306"),
		DBUser:          getEnv("DB_USER", "root"),
		DBPassword:      getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "expense_tracker"),
		JWTSecret:       []byte(secret),
		AccessTokenTTL:  time.Duration(accessMin) * time.Minute,
		RefreshTokenTTL: time.Duration(refreshHrs) * time.Hour,
		AdminEmail:      strings.ToLower(strings.TrimSpace(getEnv("ADMIN_EMAIL", ""))),
	}
	return cfg, nil
}

// DSN builds the MySQL/MariaDB data source name for GORM.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
