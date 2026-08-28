package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aurora-shop/aurora-shop/backend/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateIdentity(context.Context, CreateIdentityParams) (User, error)
	FindCredentialByEmail(context.Context, string) (UserCredential, error)
	CreateSession(context.Context, uuid.UUID, uuid.UUID, []byte, time.Time) error
	FindUserBySession(context.Context, []byte, time.Time) (User, error)
	RevokeSession(context.Context, []byte, time.Time) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateIdentity(ctx context.Context, params CreateIdentityParams) (user User, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin identity transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := db.New(tx)
	row, err := queries.CreateUser(ctx, db.CreateUserParams{
		ID: params.UserID, Email: params.Email, NormalizedEmail: params.NormalizedEmail,
		PasswordHash: params.PasswordHash, DisplayName: params.DisplayName,
	})
	if err != nil {
		return User{}, mapIdentityWriteError(err)
	}
	if _, err := queries.CreateSession(ctx, db.CreateSessionParams{
		ID: params.SessionID, UserID: params.UserID, TokenHash: params.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: params.ExpiresAt, Valid: true},
	}); err != nil {
		return User{}, mapIdentityWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit identity transaction: %w", err)
	}
	return userFromDB(row.ID, row.Email, row.DisplayName, row.Status, row.CreatedAt, row.UpdatedAt), nil
}

func (r *PostgresRepository) FindCredentialByEmail(ctx context.Context, normalizedEmail string) (UserCredential, error) {
	row, err := db.New(r.pool).GetUserCredentialByNormalizedEmail(ctx, normalizedEmail)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserCredential{}, ErrInvalidCredentials
	}
	if err != nil {
		return UserCredential{}, fmt.Errorf("find identity credential: %w", err)
	}
	return UserCredential{
		User:            userFromDB(row.ID, row.Email, row.DisplayName, row.Status, row.CreatedAt, row.UpdatedAt),
		NormalizedEmail: row.NormalizedEmail,
		PasswordHash:    row.PasswordHash,
	}, nil
}

func (r *PostgresRepository) CreateSession(ctx context.Context, sessionID, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	_, err := db.New(r.pool).CreateSession(ctx, db.CreateSessionParams{
		ID: sessionID, UserID: userID, TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return mapIdentityWriteError(err)
	}
	return nil
}

func (r *PostgresRepository) FindUserBySession(ctx context.Context, tokenHash []byte, now time.Time) (User, error) {
	row, err := db.New(r.pool).GetUserByActiveSession(ctx, db.GetUserByActiveSessionParams{
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUnauthenticated
	}
	if err != nil {
		return User{}, fmt.Errorf("find active identity session: %w", err)
	}
	return userFromDB(row.ID, row.Email, row.DisplayName, row.Status, row.CreatedAt, row.UpdatedAt), nil
}

func (r *PostgresRepository) RevokeSession(ctx context.Context, tokenHash []byte, revokedAt time.Time) error {
	if err := db.New(r.pool).RevokeSession(ctx, db.RevokeSessionParams{
		TokenHash: tokenHash,
		RevokedAt: pgtype.Timestamptz{Time: revokedAt, Valid: true},
	}); err != nil {
		return fmt.Errorf("revoke identity session: %w", err)
	}
	return nil
}

func userFromDB(id uuid.UUID, email, displayName, status string, createdAt, updatedAt pgtype.Timestamptz) User {
	return User{
		ID: id, Email: email, DisplayName: displayName, Status: status,
		CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
	}
}

func mapIdentityWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return fmt.Errorf("write identity data: %w", err)
}
