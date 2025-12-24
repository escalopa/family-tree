package domain

import "time"

type Language struct {
	LanguageCode string    `json:"language_code"`
	LanguageName string    `json:"language_name"`
	IsActive     bool      `json:"is_active"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

type MemberName struct {
	MemberNameID int       `json:"member_name_id"`
	MemberID     int       `json:"member_id"`
	LanguageCode string    `json:"language_code"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserLanguagePreference struct {
	UserID            int       `json:"user_id"`
	PreferredLanguage string    `json:"preferred_language"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type LanguageFilter struct {
	IsActive *bool
}
