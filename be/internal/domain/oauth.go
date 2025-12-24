package domain

type OAuthUserInfo struct {
	ID      string
	Email   string
	Name    string
	Picture string
}

type CookieContext interface {
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
	Cookie(name string) (string, error)
}
