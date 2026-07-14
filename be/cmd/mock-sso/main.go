package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type mockUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type server struct {
	users map[string]mockUser
}

var loginTemplate = template.Must(template.New("login").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Mock SSO</title>
  <style>
    body { font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; min-height: 100vh; display: grid; place-items: center; background: #f8fafc; color: #0f172a; }
    main { width: min(520px, calc(100vw - 32px)); background: white; border: 1px solid #e2e8f0; border-radius: 8px; padding: 24px; box-shadow: 0 12px 30px rgba(15, 23, 42, 0.12); }
    h1 { margin: 0 0 8px; font-size: 24px; }
    p { margin: 0 0 20px; color: #475569; }
    a { display: block; padding: 12px 14px; border: 1px solid #cbd5e1; border-radius: 6px; color: #0f172a; text-decoration: none; margin-top: 10px; }
    a:hover { border-color: #6366f1; background: #eef2ff; }
    small { display: block; color: #64748b; margin-top: 3px; }
  </style>
</head>
<body>
  <main>
    <h1>Mock SSO</h1>
    <p>Select a test user. This service is intended for the Docker testing stack only.</p>
    {{range .Users}}
      <a href="{{$.BaseURL}}&user={{.ID}}">
        {{.Name}}
        <small>{{.Email}}</small>
      </a>
    {{end}}
  </main>
</body>
</html>`))

func main() {
	port := getenv("MOCK_SSO_PORT", "8090")
	srv := &server{users: loadUsers()}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/authorize", srv.authorize)
	mux.HandleFunc("/userinfo", srv.userInfo)

	log.Printf("mock SSO listening on :%s with %d users", port, len(srv.users))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func (s *server) authorize(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	if redirectURI == "" || state == "" {
		http.Error(w, "redirect_uri and state are required", http.StatusBadRequest)
		return
	}

	selectedUser := r.URL.Query().Get("user")
	if selectedUser == "" {
		base := *r.URL
		query := base.Query()
		query.Del("user")
		base.RawQuery = query.Encode()

		users := make([]mockUser, 0, len(s.users))
		for _, user := range s.users {
			users = append(users, user)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := loginTemplate.Execute(w, map[string]any{
			"BaseURL": base.String(),
			"Users":   users,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if _, ok := s.users[selectedUser]; !ok {
		http.Error(w, "unknown mock user", http.StatusBadRequest)
		return
	}

	callback, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}

	query := callback.Query()
	query.Set("code", selectedUser)
	query.Set("state", state)
	callback.RawQuery = query.Encode()

	http.Redirect(w, r, callback.String(), http.StatusFound)
}

func (s *server) userInfo(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	user, ok := s.users[code]
	if !ok {
		http.Error(w, "unknown mock authorization code", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadUsers() map[string]mockUser {
	users := defaultUsers()
	raw := os.Getenv("MOCK_SSO_USERS")
	if strings.TrimSpace(raw) == "" {
		return users
	}

	var configured []mockUser
	if err := json.Unmarshal([]byte(raw), &configured); err != nil {
		log.Printf("invalid MOCK_SSO_USERS JSON, using defaults: %v", err)
		return users
	}

	users = make(map[string]mockUser, len(configured))
	for _, user := range configured {
		if user.ID == "" || user.Email == "" {
			continue
		}
		users[user.ID] = user
	}
	if len(users) == 0 {
		return defaultUsers()
	}
	return users
}

func defaultUsers() map[string]mockUser {
	return map[string]mockUser{
		"mock-superadmin": {
			ID:    "mock-superadmin",
			Email: "superadmin.mock@example.test",
			Name:  "Mock Super Admin",
		},
		"mock-admin": {
			ID:    "mock-admin",
			Email: "admin.mock@example.test",
			Name:  "Mock Admin",
		},
		"mock-guest": {
			ID:    "mock-guest",
			Email: "guest.mock@example.test",
			Name:  "Mock Guest",
		},
	}
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return fmt.Sprint(value)
}
