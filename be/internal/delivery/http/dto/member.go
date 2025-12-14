package dto

import "time"

type CreateMemberRequest struct {
	ArabicName  string     `json:"arabic_name" binding:"required"`
	EnglishName string     `json:"english_name" binding:"required"`
	Gender      string     `json:"gender" binding:"required,oneof=M F N"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	DateOfDeath *time.Time `json:"date_of_death"`
	FatherID    *int       `json:"father_id"`
	MotherID    *int       `json:"mother_id"`
	Nicknames   []string   `json:"nicknames"`
	Profession  *string    `json:"profession"`
}

type UpdateMemberRequest struct {
	ArabicName  string     `json:"arabic_name" binding:"required"`
	EnglishName string     `json:"english_name" binding:"required"`
	Gender      string     `json:"gender" binding:"required,oneof=M F N"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	DateOfDeath *time.Time `json:"date_of_death"`
	FatherID    *int       `json:"father_id"`
	MotherID    *int       `json:"mother_id"`
	Nicknames   []string   `json:"nicknames"`
	Profession  *string    `json:"profession"`
	Version     int        `json:"version" binding:"required"`
}

type MemberSearchQuery struct {
	ArabicName  *string `form:"arabic_name"`
	EnglishName *string `form:"english_name"`
	Gender      *string `form:"gender"`
	Married     *int    `form:"married"` // 0 = no, 1 = yes
	Cursor      *string `form:"cursor"`
	Limit       int     `form:"limit"`
}

type PaginatedMembersResponse struct {
	Members    []MemberResponse `json:"members"`
	NextCursor *string          `json:"next_cursor,omitempty"`
}

type MemberResponse struct {
	MemberID        int        `json:"member_id"`
	ArabicName      string     `json:"arabic_name"`
	EnglishName     string     `json:"english_name"`
	Gender          string     `json:"gender"`
	Picture         *string    `json:"picture"`
	DateOfBirth     *time.Time `json:"date_of_birth"`
	DateOfDeath     *time.Time `json:"date_of_death"`
	FatherID        *int       `json:"father_id"`
	MotherID        *int       `json:"mother_id"`
	Nicknames       []string   `json:"nicknames"`
	Profession      *string    `json:"profession"`
	Version         int        `json:"version"`
	ArabicFullName  string     `json:"arabic_full_name,omitempty"`
	EnglishFullName string     `json:"english_full_name,omitempty"`
	Age             *int       `json:"age,omitempty"`
	GenerationLevel int        `json:"generation_level,omitempty"`
	IsMarried       bool       `json:"is_married"`
	Spouses         []int      `json:"spouses,omitempty"`
}


