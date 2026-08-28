//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aurora-shop/aurora-shop/backend/internal/identity"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIdentityRepositorySessionLifecycle(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("aurora"), postgres.WithUsername("aurora"), postgres.WithPassword("aurora"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("start PostgreSQL container: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })

	databaseURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("database connection string: %v", err)
	}
	applyMigrations(t, databaseURL)
	pool := openPool(t, ctx, databaseURL)

	repository := identity.NewPostgresRepository(pool)
	hasher := identity.NewArgon2idHasher(identity.DefaultArgon2Parameters())
	tokens := identity.NewCryptoSessionTokenGenerator()
	service, err := identity.NewService(repository, hasher, tokens, time.Hour)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	password := "correct horse battery staple"
	registered, err := service.Register(ctx, identity.RegisterInput{
		Email: "Learner@Example.com", Password: password, DisplayName: "Aurora Learner",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	var storedPasswordHash string
	var storedTokenHash []byte
	if err := pool.QueryRow(ctx, `
		SELECT u.password_hash, s.token_hash
		FROM users u JOIN sessions s ON s.user_id = u.id
		WHERE u.id = $1`, registered.User.ID).Scan(&storedPasswordHash, &storedTokenHash); err != nil {
		t.Fatalf("inspect stored credentials: %v", err)
	}
	expectedTokenHash, err := tokens.Hash(registered.Token)
	if err != nil {
		t.Fatalf("Hash(token) error = %v", err)
	}
	if storedPasswordHash == password || !bytes.Equal(storedTokenHash, expectedTokenHash) || bytes.Equal(storedTokenHash, []byte(registered.Token)) {
		t.Fatal("database credential representation is unsafe or incorrect")
	}

	currentUser, err := service.Authenticate(ctx, registered.Token)
	if err != nil || currentUser.ID != registered.User.ID {
		t.Fatalf("Authenticate() = %#v, %v", currentUser, err)
	}
	if _, err := service.Login(ctx, identity.LoginInput{Email: "learner@example.com", Password: "wrong password"}); !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Fatalf("Login(wrong) error = %v", err)
	}
	loggedIn, err := service.Login(ctx, identity.LoginInput{Email: "LEARNER@example.com", Password: password})
	if err != nil {
		t.Fatalf("Login(correct) error = %v", err)
	}
	var sessionCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM sessions WHERE user_id = $1", registered.User.ID).Scan(&sessionCount); err != nil || sessionCount != 2 {
		t.Fatalf("independent session count = %d, error = %v", sessionCount, err)
	}

	if err := service.Logout(ctx, registered.Token); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err := service.Authenticate(ctx, registered.Token); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("Authenticate(revoked) error = %v", err)
	}
	loginHash, err := tokens.Hash(loggedIn.Token)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindUserBySession(ctx, loginHash, loggedIn.ExpiresAt.Add(time.Second)); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("FindUserBySession(expired) error = %v", err)
	}

	if _, err := service.Register(ctx, identity.RegisterInput{
		Email: "learner@example.com", Password: "another correct horse phrase", DisplayName: "Duplicate",
	}); !errors.Is(err, identity.ErrConflict) {
		t.Fatalf("Register(duplicate) error = %v", err)
	}

	rollbackEmail := "rollback@example.com"
	_, err = repository.CreateIdentity(ctx, identity.CreateIdentityParams{
		UserID: uuid.New(), Email: rollbackEmail, NormalizedEmail: rollbackEmail,
		PasswordHash: storedPasswordHash, DisplayName: "Rollback Candidate",
		SessionID: uuid.New(), TokenHash: storedTokenHash, ExpiresAt: time.Now().Add(time.Hour),
	})
	if !errors.Is(err, identity.ErrConflict) {
		t.Fatalf("CreateIdentity(duplicate token) error = %v", err)
	}
	var rollbackUserCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE normalized_email = $1", rollbackEmail).Scan(&rollbackUserCount); err != nil || rollbackUserCount != 0 {
		t.Fatalf("failed session insert left user behind: count = %d, error = %v", rollbackUserCount, err)
	}
}
