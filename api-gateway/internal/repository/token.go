package repository

import (
	"context"
	"errors"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTokenNotFound = errors.New("refresh token not found")

type TokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_address)
		 VALUES ($1, $2, $3, $4, $5)`,
		token.UserID, token.TokenHash, token.ExpiresAt, token.UserAgent, token.IPAddress,
	)
	return err
}

func (r *TokenRepository) FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	token := &model.RefreshToken{}

	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked, created_at, user_agent, ip_address
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt,
		&token.Revoked, &token.CreatedAt, &token.UserAgent, &token.IPAddress,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	return token, nil
}

func (r *TokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`,
		tokenHash,
	)
	return err
}

// RevokeAllForUser revokes every refresh token for a user.
// Called when token reuse is detected (indicates theft).
func (r *TokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`,
		userID,
	)
	return err
}
