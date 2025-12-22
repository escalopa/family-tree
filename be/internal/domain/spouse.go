package domain

import "time"

type Spouse struct {
	SpouseID     int        `json:"spouse_id"`
	FatherID     int        `json:"father_id"`
	MotherID     int        `json:"mother_id"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type SpouseWithMemberInfo struct {
	SpouseID     int        `json:"spouse_id"`
	MemberID     int        `json:"member_id"`
	ArabicName   string     `json:"arabic_name"`
	EnglishName  string     `json:"english_name"`
	Gender       string     `json:"gender"`
	Picture      *string    `json:"picture"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}
