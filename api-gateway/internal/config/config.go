package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	CORSOrigins string
	UploadDir   string
	JWT         JWTConfig
	Cookie      CookieConfig
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type CookieConfig struct {
	Secure   bool
	SameSite string
	Domain   string
}

func Load() *Config {
	port := getEnv("API_PORT", "8080")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("POSTGRES_USER", "outreach"),
			getEnv("POSTGRES_PASSWORD", "outreach_secret"),
			getEnv("POSTGRES_HOST", "localhost"),
			getEnv("POSTGRES_PORT", "5432"),
			getEnv("POSTGRES_DB", "outreach"),
		)
	}

	accessMins := getEnvInt("JWT_ACCESS_EXPIRY_MINUTES", 15)
	refreshDays := getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7)
	secureCookie := getEnv("COOKIE_SECURE", "false") == "true"

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
			AccessTokenExpiry:  time.Duration(accessMins) * time.Minute,
			RefreshTokenExpiry: time.Duration(refreshDays) * 24 * time.Hour,
		},
		Cookie: CookieConfig{
			Secure:   secureCookie,
			SameSite: getEnv("COOKIE_SAMESITE", "Lax"),
			Domain:   getEnv("COOKIE_DOMAIN", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
