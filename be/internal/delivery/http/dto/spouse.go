package dto

import "time"

func CalculateMarriedYears(marriageDate, divorceDate *time.Time) *int {
	if marriageDate == nil {
		return nil
	}

	endDate := time.Now()
	if divorceDate != nil {
		endDate = *divorceDate
	}

	years := int(endDate.Sub(*marriageDate).Hours() / 24 / 365.25)
	years = max(years, 0)

	return &years
}

type CreateSpouseRequest struct {
	FatherID     int   `json:"father_id" binding:"required"`
	MotherID     int   `json:"mother_id" binding:"required"`
	MarriageDate *Date `json:"marriage_date"`
	DivorceDate  *Date `json:"divorce_date"`
}

type UpdateSpouseRequest struct {
	SpouseID     int   `json:"spouse_id" binding:"required"`
	MarriageDate *Date `json:"marriage_date"`
	DivorceDate  *Date `json:"divorce_date"`
}

type DeleteSpouseRequest struct {
	SpouseID int `json:"spouse_id" binding:"required"`
}

type SpouseInfo struct {
	SpouseID     int     `json:"spouse_id"`
	MemberID     int     `json:"member_id"`
	Name         string  `json:"name"`
	Gender       string  `json:"gender"`
	Picture      *string `json:"picture"`
	MarriageDate *Date   `json:"marriage_date"`
	DivorceDate  *Date   `json:"divorce_date"`
	MarriedYears *int    `json:"married_years"`
}
