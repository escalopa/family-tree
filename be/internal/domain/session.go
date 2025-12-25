package domain

import "time"

type Session struct {
	SessionID string    `json:"session_id"`
	UserID    int       `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}







