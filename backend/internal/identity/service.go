package identity

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const (
	minimumPasswordRunes = 15
	maximumPasswordRunes = 128
	maximumPasswordBytes = 1024
)

type Service struct {
	repository Repository
	hasher     PasswordHasher
	tokens     SessionTokenGenerator
	sessionTTL time.Duration
	dummyHash  string
	now        func() time.Time
}

func NewService(repository Repository, hasher PasswordHasher, tokens SessionTokenGenerator, sessionTTL time.Duration) (*Service, error) {
	if sessionTTL <= 0 {
		return nil, errors.New("session TTL must be positive")
	}
	dummyHash, err := hasher.Hash("dummy-credential-never-used-for-login")
	if err != nil {
		return nil, fmt.Errorf("create dummy credential hash: %w", err)
	}
	return &Service{
		repository: repository, hasher: hasher, tokens: tokens,
		sessionTTL: sessionTTL, dummyHash: dummyHash, now: time.Now,
	}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	email, normalizedEmail := normalizeEmail(input.Email)
	displayName := strings.TrimSpace(input.DisplayName)
	password := norm.NFC.String(input.Password)
	if validationErr := validateRegistration(email, displayName, password); validationErr != nil {
		return AuthResult{}, validationErr
	}

	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash identity password: %w", err)
	}
	rawToken, tokenHash, err := s.tokens.New()
	if err != nil {
		return AuthResult{}, err
	}
	now := s.now().UTC()
	expiresAt := now.Add(s.sessionTTL)
	user, err := s.repository.CreateIdentity(ctx, CreateIdentityParams{
		UserID: uuid.New(), Email: email, NormalizedEmail: normalizedEmail,
		PasswordHash: passwordHash, DisplayName: displayName,
		SessionID: uuid.New(), TokenHash: tokenHash, ExpiresAt: expiresAt,
	})
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{User: user, Token: rawToken, ExpiresAt: expiresAt}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	_, normalizedEmail := normalizeEmail(input.Email)
	password := norm.NFC.String(input.Password)
	if normalizedEmail == "" || password == "" || len(password) > maximumPasswordBytes {
		return AuthResult{}, ErrInvalidCredentials
	}

	credential, err := s.repository.FindCredentialByEmail(ctx, normalizedEmail)
	if errors.Is(err, ErrInvalidCredentials) {
		_, _ = s.hasher.Verify(s.dummyHash, password)
		return AuthResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return AuthResult{}, err
	}
	valid, err := s.hasher.Verify(credential.PasswordHash, password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("verify identity password: %w", err)
	}
	if !valid || credential.Status != "active" {
		return AuthResult{}, ErrInvalidCredentials
	}

	rawToken, tokenHash, err := s.tokens.New()
	if err != nil {
		return AuthResult{}, err
	}
	now := s.now().UTC()
	expiresAt := now.Add(s.sessionTTL)
	if err := s.repository.CreateSession(ctx, uuid.New(), credential.ID, tokenHash, expiresAt); err != nil {
		return AuthResult{}, err
	}
	return AuthResult{User: credential.User, Token: rawToken, ExpiresAt: expiresAt}, nil
}

func (s *Service) Authenticate(ctx context.Context, rawToken string) (User, error) {
	tokenHash, err := s.tokens.Hash(rawToken)
	if err != nil {
		return User{}, ErrUnauthenticated
	}
	return s.repository.FindUserBySession(ctx, tokenHash, s.now().UTC())
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	tokenHash, err := s.tokens.Hash(rawToken)
	if err != nil {
		return nil
	}
	return s.repository.RevokeSession(ctx, tokenHash, s.now().UTC())
}

func normalizeEmail(rawEmail string) (string, string) {
	email := strings.TrimSpace(rawEmail)
	return email, strings.ToLower(email)
}

func validateRegistration(email, displayName, password string) error {
	fields := map[string]string{}
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || len(email) > 254 || !strings.Contains(email, "@") {
		fields["email"] = "must be a valid email address of at most 254 characters"
	}
	if count := utf8.RuneCountInString(displayName); count < 1 || count > 100 {
		fields["display_name"] = "must contain between 1 and 100 characters"
	}
	passwordRunes := utf8.RuneCountInString(password)
	if passwordRunes < minimumPasswordRunes || passwordRunes > maximumPasswordRunes || len(password) > maximumPasswordBytes {
		fields["password"] = "must contain between 15 and 128 characters"
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}
