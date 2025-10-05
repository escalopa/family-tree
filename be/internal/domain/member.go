package domain

import "time"

type Member struct {
	MemberID    int    `json:"member_id" db:"member_id"`
	ArabicName  string `json:"arabic_name" db:"arabic_name"`
	EnglishName string `json:"english_name" db:"english_name"`
	Gender      string `json:"gender" db:"gender"`
}

type MemberInfo struct {
	Picture     []byte     `json:"-" db:"picture"`
	DateOfBirth *time.Time `json:"date_of_birth" db:"date_of_birth"`
	DateOfDeath *time.Time `json:"date_of_death" db:"date_of_death"`
	FatherID    *int       `json:"father_id" db:"father_id"`
	MotherID    *int       `json:"mother_id" db:"mother_id"`
	Nicknames   []string   `json:"nicknames" db:"nicknames"`
	Profession  *string    `json:"profession" db:"profession"`
	Revision    int        `json:"revision" db:"revision"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
}

type MemberWithDetails struct {
	MemberInfo
	ArabicFullName  string       `json:"arabic_full_name"`
	EnglishFullName string       `json:"english_full_name"`
	Age             *int         `json:"age"`
	GenerationLevel int          `json:"generation_level"`
	HasPicture      bool         `json:"has_picture"`
	IsMarried       bool         `json:"is_married"`
	Spouses         []SpouseInfo `json:"spouses"`
}

type SpouseInfo struct {
	Member       Member     `json:"member"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}
