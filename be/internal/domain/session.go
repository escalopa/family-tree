package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	SessionID uuid.UUID `json:"session_id" db:"session_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	IssuedAt  time.Time `json:"issued_at" db:"issued_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}
