package dto

type CreateMemberRequest struct {
	ArabicName  string   `json:"arabic_name" binding:"required"`
	EnglishName string   `json:"english_name" binding:"required"`
	Gender      string   `json:"gender" binding:"required,oneof=M F"`
	DateOfBirth *Date    `json:"date_of_birth"`
	DateOfDeath *Date    `json:"date_of_death"`
	FatherID    *int     `json:"father_id"`
	MotherID    *int     `json:"mother_id"`
	Nicknames   []string `json:"nicknames"`
	Profession  *string  `json:"profession"`
}

type UpdateMemberRequest struct {
	ArabicName  string   `json:"arabic_name" binding:"required"`
	EnglishName string   `json:"english_name" binding:"required"`
	Gender      string   `json:"gender" binding:"required,oneof=M F"`
	DateOfBirth *Date    `json:"date_of_birth"`
	DateOfDeath *Date    `json:"date_of_death"`
	FatherID    *int     `json:"father_id"`
	MotherID    *int     `json:"mother_id"`
	Nicknames   []string `json:"nicknames"`
	Profession  *string  `json:"profession"`
	Version     int      `json:"version" binding:"required,min=1"`
}

type MemberSearchQuery struct {
	Name    *string `form:"name"`
	Gender  *string `form:"gender"`
	Married *int    `form:"married"` // 0 = no, 1 = yes
	Cursor  *string `form:"cursor"`
	Limit   int     `form:"limit" binding:"omitempty,min=1,max=100"`
}

type PaginatedMembersResponse struct {
	Members    []MemberResponse `json:"members"`
	NextCursor *string          `json:"next_cursor,omitempty"`
}

type ParentInfo struct {
	MemberID    int     `json:"member_id"`
	ArabicName  string  `json:"arabic_name"`
	EnglishName string  `json:"english_name"`
	Picture     *string `json:"picture"`
}

type MemberResponse struct {
	MemberID        int          `json:"member_id"`
	ArabicName      string       `json:"arabic_name"`
	EnglishName     string       `json:"english_name"`
	Gender          string       `json:"gender"`
	Picture         *string      `json:"picture"`
	DateOfBirth     *Date        `json:"date_of_birth"`
	DateOfDeath     *Date        `json:"date_of_death"`
	FatherID        *int         `json:"father_id"`
	MotherID        *int         `json:"mother_id"`
	Father          *ParentInfo  `json:"father,omitempty"`
	Mother          *ParentInfo  `json:"mother,omitempty"`
	Nicknames       []string     `json:"nicknames"`
	Profession      *string      `json:"profession"`
	Version         int          `json:"version"`
	ArabicFullName  string       `json:"arabic_full_name,omitempty"`
	EnglishFullName string       `json:"english_full_name,omitempty"`
	Age             *int         `json:"age,omitempty"`
	GenerationLevel int          `json:"generation_level,omitempty"`
	IsMarried       bool         `json:"is_married"`
	Spouses         []SpouseInfo `json:"spouses,omitempty"`
}

type ParentSearchQuery struct {
	Query  string `form:"q" binding:"required,min=1"`
	Gender string `form:"gender" binding:"required,oneof=M F"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

type ParentOption struct {
	MemberID    int     `json:"member_id"`
	ArabicName  string  `json:"arabic_name"`
	EnglishName string  `json:"english_name"`
	Picture     *string `json:"picture"`
	Gender      string  `json:"gender"`
}
