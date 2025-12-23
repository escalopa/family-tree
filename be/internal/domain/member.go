package domain

import (
	"time"
)

type Member struct {
	MemberID    int        `json:"member_id"`
	ArabicName  string     `json:"arabic_name"`
	EnglishName string     `json:"english_name"`
	Gender      string     `json:"gender"` // M, F
	Picture     *string    `json:"picture"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	DateOfDeath *time.Time `json:"date_of_death"`
	FatherID    *int       `json:"father_id"`
	MotherID    *int       `json:"mother_id"`
	Nicknames   []string   `json:"nicknames"`
	Profession  *string    `json:"profession"`
	Version     int        `json:"version"`
	DeletedAt   *time.Time `json:"deleted_at"`
	IsMarried   bool       `json:"is_married"` // Computed field for search results
}

// Computed fields
type MemberWithComputed struct {
	Member
	ArabicFullName  string                 `json:"arabic_full_name"`
	EnglishFullName string                 `json:"english_full_name"`
	Age             *int                   `json:"age"`
	GenerationLevel int                    `json:"generation_level"`
	IsMarried       bool                   `json:"is_married"`
	Spouses         []SpouseWithMemberInfo `json:"spouses,omitempty"`
}

type MemberTreeNode struct {
	MemberWithComputed
	Children    []*MemberTreeNode `json:"children,omitempty"`
	SpouseNodes []*MemberTreeNode `json:"spouse_nodes,omitempty"` // Spouse nodes at same level
	IsInPath    bool              `json:"is_in_path,omitempty"`   // For relation path highlighting
}
