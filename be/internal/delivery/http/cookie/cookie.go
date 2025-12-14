package cookie

import (
	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/domain"
)

const (
	AccessTokenCookie  = "auth_token"
	RefreshTokenCookie = "refresh_token"
	SessionIDCookie    = "session_id"
)

type Manager struct {
	accessTokenMaxAge  int
	refreshTokenMaxAge int
	sessionIDMaxAge    int
	path               string
	domain             string
	secure             bool
	httpOnly           bool
}

func NewManager(cfg *config.CookieConfig) *Manager {
	return &Manager{
		accessTokenMaxAge:  cfg.AccessTokenMaxAge,
		refreshTokenMaxAge: cfg.RefreshTokenMaxAge,
		sessionIDMaxAge:    cfg.SessionIDMaxAge,
		path:               cfg.Path,
		domain:             cfg.Domain,
		secure:             cfg.Secure,
		httpOnly:           cfg.HttpOnly,
	}
}

func (m *Manager) SetAuthCookies(c domain.CookieContext, accessToken, refreshToken, sessionID string) {
	c.SetCookie(AccessTokenCookie, accessToken, m.accessTokenMaxAge, m.path, m.domain, m.secure, m.httpOnly)
	c.SetCookie(RefreshTokenCookie, refreshToken, m.refreshTokenMaxAge, m.path, m.domain, m.secure, m.httpOnly)
	c.SetCookie(SessionIDCookie, sessionID, m.sessionIDMaxAge, m.path, m.domain, m.secure, m.httpOnly)
}

func (m *Manager) SetTokenCookies(c domain.CookieContext, accessToken, refreshToken string) {
	c.SetCookie(AccessTokenCookie, accessToken, m.accessTokenMaxAge, m.path, m.domain, m.secure, m.httpOnly)
	c.SetCookie(RefreshTokenCookie, refreshToken, m.refreshTokenMaxAge, m.path, m.domain, m.secure, m.httpOnly)
}

func (m *Manager) ClearAuthCookies(c domain.CookieContext) {
	c.SetCookie(AccessTokenCookie, "", -1, m.path, m.domain, m.secure, m.httpOnly)
	c.SetCookie(RefreshTokenCookie, "", -1, m.path, m.domain, m.secure, m.httpOnly)
	c.SetCookie(SessionIDCookie, "", -1, m.path, m.domain, m.secure, m.httpOnly)
}

func (m *Manager) GetAccessToken(c domain.CookieContext) (string, error) {
	return c.Cookie(AccessTokenCookie)
}

func (m *Manager) GetRefreshToken(c domain.CookieContext) (string, error) {
	return c.Cookie(RefreshTokenCookie)
}

func (m *Manager) GetSessionID(c domain.CookieContext) (string, error) {
	return c.Cookie(SessionIDCookie)
}
