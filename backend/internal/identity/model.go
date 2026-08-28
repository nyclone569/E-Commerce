package identity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RegisterInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	User      User
	Token     string
	ExpiresAt time.Time
}

type UserCredential struct {
	User
	NormalizedEmail string
	PasswordHash    string
}

type CreateIdentityParams struct {
	UserID          uuid.UUID
	Email           string
	NormalizedEmail string
	PasswordHash    string
	DisplayName     string
	SessionID       uuid.UUID
	TokenHash       []byte
	ExpiresAt       time.Time
}
