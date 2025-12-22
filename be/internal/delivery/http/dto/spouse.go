package dto

type CreateSpouseRequest struct {
	FatherID     int   `json:"father_id" binding:"required"`
	MotherID     int   `json:"mother_id" binding:"required"`
	MarriageDate *Date `json:"marriage_date"`
	DivorceDate  *Date `json:"divorce_date"`
}

type UpdateSpouseRequest struct {
	FatherID     int   `json:"father_id" binding:"required"`
	MotherID     int   `json:"mother_id" binding:"required"`
	MarriageDate *Date `json:"marriage_date"`
	DivorceDate  *Date `json:"divorce_date"`
}

type UpdateSpouseByMemberRequest struct {
	SpouseID     int   `json:"spouse_id" binding:"required"`
	MarriageDate *Date `json:"marriage_date"`
	DivorceDate  *Date `json:"divorce_date"`
}

type SpouseInfo struct {
	SpouseID     int     `json:"spouse_id"`
	MemberID     int     `json:"member_id"`
	ArabicName   string  `json:"arabic_name"`
	EnglishName  string  `json:"english_name"`
	Gender       string  `json:"gender"`
	Picture      *string `json:"picture"`
	MarriageDate *Date   `json:"marriage_date"`
	DivorceDate  *Date   `json:"divorce_date"`
}
