package identity

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubRepository struct {
	createParams       CreateIdentityParams
	createdUser        User
	createErr          error
	credential         UserCredential
	credentialErr      error
	createdSessionUser uuid.UUID
	createdSessionHash []byte
	createdExpiresAt   time.Time
	activeUser         User
	activeHash         []byte
	activeNow          time.Time
	activeErr          error
	revokedHash        []byte
	revokedAt          time.Time
}

func (r *stubRepository) CreateIdentity(_ context.Context, params CreateIdentityParams) (User, error) {
	r.createParams = params
	return r.createdUser, r.createErr
}

func (r *stubRepository) FindCredentialByEmail(_ context.Context, _ string) (UserCredential, error) {
	return r.credential, r.credentialErr
}

func (r *stubRepository) CreateSession(_ context.Context, _ uuid.UUID, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	r.createdSessionUser = userID
	r.createdSessionHash = tokenHash
	r.createdExpiresAt = expiresAt
	return nil
}

func (r *stubRepository) FindUserBySession(_ context.Context, tokenHash []byte, now time.Time) (User, error) {
	r.activeHash = tokenHash
	r.activeNow = now
	return r.activeUser, r.activeErr
}

func (r *stubRepository) RevokeSession(_ context.Context, tokenHash []byte, revokedAt time.Time) error {
	r.revokedHash = tokenHash
	r.revokedAt = revokedAt
	return nil
}

type stubHasher struct {
	verifiedHash     string
	verifiedPassword string
	verifyResult     bool
}

func (*stubHasher) Hash(password string) (string, error) { return "hash:" + password, nil }

func (h *stubHasher) Verify(encodedHash, password string) (bool, error) {
	h.verifiedHash = encodedHash
	h.verifiedPassword = password
	return h.verifyResult, nil
}

type stubTokens struct{}

func (*stubTokens) New() (string, []byte, error) {
	return "raw-token", bytes.Repeat([]byte{7}, 32), nil
}

func (*stubTokens) Hash(rawToken string) ([]byte, error) {
	if rawToken != "raw-token" {
		return nil, errors.New("invalid token")
	}
	return bytes.Repeat([]byte{7}, 32), nil
}

func newTestService(t *testing.T, repository Repository, hasher PasswordHasher) *Service {
	t.Helper()
	service, err := NewService(repository, hasher, &stubTokens{}, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	service.now = func() time.Time { return time.Date(2026, time.August, 28, 10, 0, 0, 0, time.UTC) }
	return service
}

func TestRegisterNormalizesAndCreatesIdentity(t *testing.T) {
	userID := uuid.New()
	repository := &stubRepository{createdUser: User{ID: userID, Email: "Learner@Example.com"}}
	service := newTestService(t, repository, &stubHasher{})

	result, err := service.Register(context.Background(), RegisterInput{
		Email: "  Learner@Example.com ", Password: "a sufficiently long passphrase", DisplayName: "  Aurora Learner  ",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repository.createParams.Email != "Learner@Example.com" || repository.createParams.NormalizedEmail != "learner@example.com" || repository.createParams.DisplayName != "Aurora Learner" {
		t.Fatalf("unexpected normalized params: %#v", repository.createParams)
	}
	if !stringsHasPrefix(repository.createParams.PasswordHash, "hash:") || result.Token != "raw-token" || result.User.ID != userID {
		t.Fatalf("unexpected result: %#v", result)
	}
	if got, want := repository.createParams.ExpiresAt, service.now().Add(24*time.Hour); !got.Equal(want) {
		t.Fatalf("expiresAt = %v, want %v", got, want)
	}
}

func TestRegisterRejectsInvalidFieldsBeforeRepository(t *testing.T) {
	repository := &stubRepository{}
	service := newTestService(t, repository, &stubHasher{})
	_, err := service.Register(context.Background(), RegisterInput{Email: "bad", Password: "short", DisplayName: ""})
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Register() error = %v, want ValidationError", err)
	}
	for _, field := range []string{"email", "password", "display_name"} {
		if validationErr.Fields[field] == "" {
			t.Errorf("missing validation for %s", field)
		}
	}
	if repository.createParams.UserID != uuid.Nil {
		t.Fatal("repository was called for invalid registration")
	}
}

func TestLoginUsesGenericInvalidCredentialPath(t *testing.T) {
	repository := &stubRepository{credentialErr: ErrInvalidCredentials}
	hasher := &stubHasher{}
	service := newTestService(t, repository, hasher)
	_, err := service.Login(context.Background(), LoginInput{Email: "missing@example.com", Password: "some submitted password"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v", err)
	}
	if hasher.verifiedHash != service.dummyHash || hasher.verifiedPassword != "some submitted password" {
		t.Fatal("unknown account did not execute dummy password verification")
	}
}

func TestLoginCreatesIndependentSession(t *testing.T) {
	userID := uuid.New()
	repository := &stubRepository{credential: UserCredential{
		User: User{ID: userID, Status: "active"}, PasswordHash: "stored-password-hash",
	}}
	hasher := &stubHasher{verifyResult: true}
	service := newTestService(t, repository, hasher)
	result, err := service.Login(context.Background(), LoginInput{Email: "learner@example.com", Password: "correct password phrase"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Token != "raw-token" || repository.createdSessionUser != userID || len(repository.createdSessionHash) != 32 {
		t.Fatalf("session was not created correctly: %#v", result)
	}
}

func TestAuthenticateAndLogoutUseHashedToken(t *testing.T) {
	repository := &stubRepository{activeUser: User{ID: uuid.New()}}
	service := newTestService(t, repository, &stubHasher{})
	user, err := service.Authenticate(context.Background(), "raw-token")
	if err != nil || user.ID == uuid.Nil || len(repository.activeHash) != 32 {
		t.Fatalf("Authenticate() = %#v, %v", user, err)
	}
	if err := service.Logout(context.Background(), "raw-token"); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if len(repository.revokedHash) != 32 || !repository.revokedAt.Equal(service.now()) {
		t.Fatal("Logout() did not revoke the hashed token")
	}
	if err := service.Logout(context.Background(), "malformed"); err != nil {
		t.Fatalf("Logout(malformed) error = %v, want idempotent nil", err)
	}
}

func stringsHasPrefix(value, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}
