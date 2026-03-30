package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	CORSOrigins    string
	UploadDir      string
	AIServiceURL   string
	JWT            JWTConfig
	Cookie         CookieConfig
	SMTP           SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
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

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = fmt.Sprintf("redis://%s:%s/0",
			getEnv("REDIS_HOST", "localhost"),
			getEnv("REDIS_PORT", "6379"),
		)
	}

	smtpFrom := getEnv("SMTP_FROM", "")
	if smtpFrom == "" {
		smtpFrom = getEnv("SMTP_USER", "")
	}

	return &Config{
		Port:         port,
		DatabaseURL:  dbURL,
		RedisURL:     redisURL,
		CORSOrigins:  getEnv("CORS_ORIGINS", "http://localhost:3000"),
		UploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
		AIServiceURL: getEnv("AI_SERVICE_URL", "http://localhost:8000"),
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
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnv("SMTP_PORT", "587"),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     smtpFrom,
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
