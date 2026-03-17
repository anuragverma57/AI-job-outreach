package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/config"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidInput       = errors.New("invalid input")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenRevoked       = errors.New("token has been revoked")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	cfg       *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, *TokenPair, error) {
	if err := validateRegisterInput(req); err != nil {
		return nil, nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, req.Email, string(hash), req.Name)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokens(ctx, user.ID, "", "")
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest, userAgent, ipAddress string) (*model.User, *TokenPair, error) {
	if req.Email == "" || req.Password == "" {
		return nil, nil, ErrInvalidInput
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, user.ID, userAgent, ipAddress)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken, userAgent, ipAddress string) (*model.User, *TokenPair, error) {
	if rawRefreshToken == "" {
		return nil, nil, ErrInvalidInput
	}

	tokenHash := hashToken(rawRefreshToken)

	stored, err := s.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Reuse detection: if the token was already revoked, someone stole it.
	// Revoke ALL tokens for this user as a precaution.
	if stored.Revoked {
		_ = s.tokenRepo.RevokeAllForUser(ctx, stored.UserID)
		return nil, nil, ErrTokenRevoked
	}

	if time.Now().After(stored.ExpiresAt) {
		_ = s.tokenRepo.Revoke(ctx, tokenHash)
		return nil, nil, ErrTokenExpired
	}

	// Rotate: revoke the old token, issue new pair
	if err := s.tokenRepo.Revoke(ctx, tokenHash); err != nil {
		return nil, nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokens(ctx, user.ID, userAgent, ipAddress)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}

	tokenHash := hashToken(rawRefreshToken)
	return s.tokenRepo.Revoke(ctx, tokenHash)
}

func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *AuthService) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return "", errors.New("invalid user id in token")
	}

	return userID, nil
}

// --- private helpers ---

func (s *AuthService) issueTokens(ctx context.Context, userID, userAgent, ipAddress string) (*TokenPair, error) {
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	rawRefresh, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshRecord := &model.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshTokenExpiry),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	if err := s.tokenRepo.Create(ctx, refreshRecord); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}

func (s *AuthService) generateAccessToken(userID string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(s.cfg.JWT.AccessTokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// generateRefreshToken creates a 32-byte cryptographically random token
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken uses SHA-256 to hash the refresh token before storage.
// SHA-256 is safe here because the input is high-entropy (32 random bytes).
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func validateRegisterInput(req model.RegisterRequest) error {
	if req.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if req.Password == "" || len(req.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidInput)
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return fmt.Errorf("%w: invalid email format", ErrInvalidInput)
	}
	return nil
}
