package dto

import "time"

type CreateSpouseRequest struct {
	Member1ID    int        `json:"member1_id" binding:"required"`
	Member2ID    int        `json:"member2_id" binding:"required"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}

type UpdateSpouseRequest struct {
	Member1ID    int        `json:"member1_id" binding:"required"`
	Member2ID    int        `json:"member2_id" binding:"required"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}
