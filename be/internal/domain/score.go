package domain

import "time"

const (
	// Mandatory fields
	PointsArabicName  = 1
	PointsEnglishName = 1
	PointsGender      = 1

	// Optional fields
	PointsPicture     = 2
	PointsDateOfBirth = 5
	PointsDateOfDeath = 5
	PointsFather      = 3
	PointsMother      = 3
	PointsSpouse      = 3
	PointsNicknames   = 1
	PointsProfession  = 1
)

type Score struct {
	UserID        int       `json:"user_id"`
	MemberID      int       `json:"member_id"`
	FieldName     string    `json:"field_name"`
	Points        int       `json:"points"`
	MemberVersion int       `json:"member_version"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserScore struct {
	UserID     int     `json:"user_id"`
	FullName   string  `json:"full_name"`
	Avatar     *string `json:"avatar"`
	TotalScore int     `json:"total_score"`
	Rank       int     `json:"rank"`
}

type ScoreHistory struct {
	Score
	MemberArabicName  string `json:"member_arabic_name"`
	MemberEnglishName string `json:"member_english_name"`
}
