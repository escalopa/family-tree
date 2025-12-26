package domain

import (
	"time"
)

type Member struct {
	MemberID    int               `json:"member_id"`
	Names       map[string]string `json:"names"` // language_code -> name
	Gender      string            `json:"gender"`
	Picture     *string           `json:"picture"`
	DateOfBirth *time.Time        `json:"date_of_birth"`
	DateOfDeath *time.Time        `json:"date_of_death"`
	FatherID    *int              `json:"father_id"`
	MotherID    *int              `json:"mother_id"`
	Nicknames   []string          `json:"nicknames"`
	Profession  *string           `json:"profession"`
	Version     int               `json:"version"`
	DeletedAt   *time.Time        `json:"deleted_at"`
	IsMarried   bool              `json:"is_married"`
}

type MemberWithComputed struct {
	Member
	FullNames       map[string]string      `json:"full_names"` // language_code -> full_name
	Age             *int                   `json:"age"`
	GenerationLevel int                    `json:"generation_level"`
	IsMarried       bool                   `json:"is_married"`
	Spouses         []SpouseWithMemberInfo `json:"spouses,omitempty"`
}

type MemberTreeNode struct {
	MemberWithComputed
	Children []*MemberTreeNode `json:"children,omitempty"`
	IsInPath bool              `json:"is_in_path,omitempty"`
}
