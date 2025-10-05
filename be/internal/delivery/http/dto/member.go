package dto

import "time"

type MemberSummaryResponse struct {
	MemberID    int        `json:"member_id"`
	ArabicName  string     `json:"arabic_name"`
	EnglishName string     `json:"english_name"`
	Gender      string     `json:"gender"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	DateOfDeath *time.Time `json:"date_of_death"`
	HasPicture  bool       `json:"has_picture"`
	IsMarried   bool       `json:"is_married"`
}

type MemberResponse struct {
	MemberID        int          `json:"member_id"`
	ArabicName      string       `json:"arabic_name"`
	EnglishName     string       `json:"english_name"`
	ArabicFullName  string       `json:"arabic_full_name"`
	EnglishFullName string       `json:"english_full_name"`
	Gender          string       `json:"gender"`
	PictureURL      *string      `json:"picture_url"`
	DateOfBirth     *time.Time   `json:"date_of_birth"`
	DateOfDeath     *time.Time   `json:"date_of_death"`
	Age             *int         `json:"age"`
	FatherID        *int         `json:"father_id"`
	MotherID        *int         `json:"mother_id"`
	Spouses         []SpouseInfo `json:"spouses"`
	Nicknames       []string     `json:"nicknames"`
	Profession      *string      `json:"profession"`
	GenerationLevel int          `json:"generation_level"`
	Revision        int          `json:"revision"`
	DeletedAt       *time.Time   `json:"deleted_at"`
}

type MemberCreateRequest struct {
	ArabicName  string     `json:"arabic_name" binding:"required,max=255"`
	EnglishName string     `json:"english_name" binding:"required,max=255"`
	Gender      string     `json:"gender" binding:"required,oneof=M F"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	DateOfDeath *time.Time `json:"date_of_death"`
	FatherID    *int       `json:"father_id"`
	MotherID    *int       `json:"mother_id"`
	Nicknames   []string   `json:"nicknames"`
	Profession  *string    `json:"profession" binding:"omitempty,max=255"`
}

type MemberUpdateRequest struct {
	ArabicName  string     `json:"arabic_name" binding:"required,max=255"`
	EnglishName string     `json:"english_name" binding:"required,max=255"`
	Gender      string     `json:"gender" binding:"required,oneof=M F"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	DateOfDeath *time.Time `json:"date_of_death"`
	FatherID    *int       `json:"father_id"`
	MotherID    *int       `json:"mother_id"`
	Nicknames   []string   `json:"nicknames"`
	Profession  *string    `json:"profession" binding:"omitempty,max=255"`
	Revision    int        `json:"revision" binding:"required,min=1"`
}

type MemberPatchRequest struct {
	ArabicName  *string    `json:"arabic_name" binding:"omitempty,max=255"`
	EnglishName *string    `json:"english_name" binding:"omitempty,max=255"`
	Gender      *string    `json:"gender" binding:"omitempty,oneof=M F"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	DateOfDeath *time.Time `json:"date_of_death"`
	FatherID    *int       `json:"father_id"`
	MotherID    *int       `json:"mother_id"`
	Nicknames   []string   `json:"nicknames"`
	Profession  *string    `json:"profession" binding:"omitempty,max=255"`
	Revision    int        `json:"revision" binding:"required,min=1"`
}

type MemberListResponse struct {
	Members []MemberSummaryResponse `json:"members"`
}

type PictureUploadResponse struct {
	Message  string `json:"message"`
	Revision int    `json:"revision"`
}

type MemberPicturesResponse struct {
	Pictures []MemberPictureInfo `json:"pictures"`
}

type MemberPictureInfo struct {
	MemberID   int     `json:"member_id"`
	PictureURL *string `json:"picture_url"`
}

type RollbackRequest struct {
	Revision int `json:"revision" binding:"required,min=0"`
}

type SpouseInfo struct {
	MemberID     int        `json:"member_id"`
	Name         string     `json:"name"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}
