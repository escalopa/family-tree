package domain

import "time"

type MemberSpouse struct {
	MemberID     int        `json:"member_id" db:"member_id"`
	MarriageDate *time.Time `json:"marriage_date" db:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date" db:"divorce_date"`
}
