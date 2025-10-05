package dto

import "time"

type SpouseCreateRequest struct {
	Member1ID    int        `json:"member1_id" binding:"required"`
	Member2ID    int        `json:"member2_id" binding:"required"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}

type SpousePatchRequest struct {
	Member1ID    int        `json:"member1_id" binding:"required"`
	Member2ID    int        `json:"member2_id" binding:"required"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}

type SpouseDeleteRequest struct {
	Member1ID int `json:"member1_id" binding:"required"`
	Member2ID int `json:"member2_id" binding:"required"`
}

type SpouseResponse struct {
	Member1ID    int        `json:"member1_id"`
	Member2ID    int        `json:"member2_id"`
	Member1Name  string     `json:"member1_name"`
	Member2Name  string     `json:"member2_name"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}
