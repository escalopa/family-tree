package domain

import "time"

type Spouse struct {
	Member1ID    int        `json:"member1_id"`
	Member2ID    int        `json:"member2_id"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}


