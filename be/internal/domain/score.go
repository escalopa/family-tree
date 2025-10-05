package domain

import "time"

type UserScore struct {
	UserID        int       `json:"user_id" db:"user_id"`
	MemberID      int       `json:"member_id" db:"member_id"`
	FieldName     string    `json:"field_name" db:"field_name"`
	Points        int       `json:"points" db:"points"`
	MemberVersion int       `json:"member_version" db:"member_version"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type UserScoreWithDetails struct {
	UserScore
	MemberName string `json:"member_name"`
}

type LeaderboardEntry struct {
	Rank               int     `json:"rank"`
	UserID             int     `json:"user_id"`
	FullName           string  `json:"full_name"`
	Avatar             *string `json:"avatar"`
	TotalScore         int     `json:"total_score"`
	ContributionsCount int     `json:"contributions_count"`
}

const (
	PointsArabicName  = 1
	PointsEnglishName = 1
	PointsGender      = 1
	PointsPicture     = 2
	PointsDateOfBirth = 5
	PointsDateOfDeath = 5
	PointsFather      = 3
	PointsMother      = 3
	PointsSpouse      = 3
	PointsNicknames   = 1
	PointsProfession  = 1
)
