package dto

type CreateMemberRequest struct {
	Names       map[string]string `json:"names" binding:"required,min=1"` // language_code -> name
	Gender      string            `json:"gender" binding:"required,oneof=M F"`
	DateOfBirth *Date             `json:"date_of_birth"`
	DateOfDeath *Date             `json:"date_of_death"`
	FatherID    *int              `json:"father_id"`
	MotherID    *int              `json:"mother_id"`
	Nicknames   []string          `json:"nicknames"`
	Profession  *string           `json:"profession"`
}

type UpdateMemberRequest struct {
	Names       map[string]string `json:"names" binding:"required,min=1"` // language_code -> name
	Gender      string            `json:"gender" binding:"required,oneof=M F"`
	DateOfBirth *Date             `json:"date_of_birth"`
	DateOfDeath *Date             `json:"date_of_death"`
	FatherID    *int              `json:"father_id"`
	MotherID    *int              `json:"mother_id"`
	Nicknames   []string          `json:"nicknames"`
	Profession  *string           `json:"profession"`
	Version     int               `json:"version" binding:"required,min=1"`
}

type MemberSearchQuery struct {
	Name    *string `form:"name"`
	Gender  *string `form:"gender"`
	Married *int    `form:"married"` // 0 = no, 1 = yes
	Cursor  *string `form:"cursor"`
	Limit   int     `form:"limit,default=20" binding:"omitempty,min=1,max=100"`
}

type MemberListItem struct {
	MemberID    int     `json:"member_id"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Picture     *string `json:"picture"`
	DateOfBirth *Date   `json:"date_of_birth"`
	DateOfDeath *Date   `json:"date_of_death"`
	IsMarried   bool    `json:"is_married"`
}

type PaginatedMembersResponse struct {
	Members    []MemberListItem `json:"members"`
	NextCursor *string          `json:"next_cursor,omitempty"`
}

type MemberInfo struct {
	MemberID int     `json:"member_id"`
	Name     string  `json:"name"`
	Picture  *string `json:"picture"`
	Gender   string  `json:"gender"`
}

type MemberResponse struct {
	MemberID        int               `json:"member_id"`
	Name            string            `json:"name"`                 // Name in user's preferred language
	Names           map[string]string `json:"names"`                // language_code -> name (for editing)
	FullName        string            `json:"full_name,omitempty"`  // Full name in user's preferred language
	FullNames       map[string]string `json:"full_names,omitempty"` // language_code -> full_name (for editing)
	Gender          string            `json:"gender"`
	Picture         *string           `json:"picture"`
	DateOfBirth     *Date             `json:"date_of_birth"`
	DateOfDeath     *Date             `json:"date_of_death"`
	FatherID        *int              `json:"father_id"`
	MotherID        *int              `json:"mother_id"`
	Father          *MemberInfo       `json:"father,omitempty"`
	Mother          *MemberInfo       `json:"mother,omitempty"`
	Nicknames       []string          `json:"nicknames"`
	Profession      *string           `json:"profession"`
	Version         int               `json:"version"`
	Age             *int              `json:"age,omitempty"`
	GenerationLevel int               `json:"generation_level,omitempty"`
	IsMarried       bool              `json:"is_married"`
	Spouses         []SpouseInfo      `json:"spouses,omitempty"`
	Children        []MemberInfo      `json:"children,omitempty"`
	Siblings        []MemberInfo      `json:"siblings,omitempty"`
}
