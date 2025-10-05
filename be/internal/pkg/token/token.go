package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	authSecret    string
	refreshSecret string
	authExpiry    time.Duration
	refreshExpiry time.Duration
}

type Claims struct {
	UserID    int       `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	jwt.RegisteredClaims
}

func NewManager(authSecret, refreshSecret string, authExpiry, refreshExpiry time.Duration) *Manager {
	return &Manager{
		authSecret:    authSecret,
		refreshSecret: refreshSecret,
		authExpiry:    authExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func (m *Manager) generateToken(userID int, sessionID uuid.UUID, secret string, expiry time.Duration, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(expiry)
	claims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	return signed, expiresAt, err
}

func (m *Manager) GenerateAuthToken(userID int, sessionID uuid.UUID, now time.Time) (string, time.Time, error) {
	return m.generateToken(userID, sessionID, m.authSecret, m.authExpiry, now)
}

func (m *Manager) GenerateRefreshToken(userID int, sessionID uuid.UUID, now time.Time) (string, time.Time, error) {
	return m.generateToken(userID, sessionID, m.refreshSecret, m.refreshExpiry, now)
}

func (m *Manager) validateToken(tokenString string, secret string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsed.Claims.(*Claims); ok && parsed.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

func (m *Manager) ValidateAuthToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.authSecret)
}

func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.refreshSecret)
}
