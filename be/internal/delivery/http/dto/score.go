package dto

import "time"

type UserScoreResponse struct {
	MemberID      int       `json:"member_id"`
	MemberName    string    `json:"member_name"`
	FieldName     string    `json:"field_name"`
	Points        int       `json:"points"`
	MemberVersion int       `json:"member_version"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserScoresResponse struct {
	UserID        int                 `json:"user_id"`
	TotalScore    int                 `json:"total_score"`
	Contributions []UserScoreResponse `json:"contributions"`
}

type LeaderboardEntryResponse struct {
	Rank               int     `json:"rank"`
	UserID             int     `json:"user_id"`
	FullName           string  `json:"full_name"`
	Avatar             *string `json:"avatar"`
	TotalScore         int     `json:"total_score"`
	ContributionsCount int     `json:"contributions_count"`
}

type LeaderboardResponse struct {
	Leaderboard []LeaderboardEntryResponse `json:"leaderboard"`
	Pagination  PaginationResponse         `json:"pagination"`
}
