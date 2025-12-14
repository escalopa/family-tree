package domain

type OAuthUserInfo struct {
	ID            string
	Email         string
	VerifiedEmail bool
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
	Locale        string
}

type CookieContext interface {
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
	Cookie(name string) (string, error)
}
