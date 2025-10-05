package dto

import "time"

type TreeNodeResponse struct {
	MemberID        int                `json:"member_id"`
	ArabicName      string             `json:"arabic_name"`
	EnglishName     string             `json:"english_name"`
	Gender          string             `json:"gender"`
	DateOfBirth     *time.Time         `json:"date_of_birth"`
	DateOfDeath     *time.Time         `json:"date_of_death"`
	Age             *int               `json:"age"`
	GenerationLevel int                `json:"generation_level"`
	HasPicture      bool               `json:"has_picture"`
	Profession      *string            `json:"profession"`
	Nicknames       []string           `json:"nicknames"`
	FatherID        *int               `json:"father_id"`
	MotherID        *int               `json:"mother_id"`
	Spouse          *TreeSpouseInfo    `json:"spouse"`
	Children        []TreeNodeResponse `json:"children"`
}

type TreeSpouseInfo struct {
	MemberID     int        `json:"member_id"`
	Name         string     `json:"name"`
	MarriageDate *time.Time `json:"marriage_date"`
	DivorceDate  *time.Time `json:"divorce_date"`
}

type TreeMetadataResponse struct {
	TotalMembers  int `json:"total_members"`
	MaxGeneration int `json:"max_generation"`
	RootMemberID  int `json:"root_member_id"`
}

type TreeResponse struct {
	Root     TreeNodeResponse     `json:"root"`
	Metadata TreeMetadataResponse `json:"metadata"`
}

type RelationMemberInfo struct {
	MemberID int    `json:"member_id"`
	Name     string `json:"name"`
}

type RelationPathItem struct {
	MemberID     int    `json:"member_id"`
	Name         string `json:"name"`
	RelationType string `json:"relation_type"`
}

type RelationResponse struct {
	Member1      RelationMemberInfo `json:"member1"`
	Member2      RelationMemberInfo `json:"member2"`
	Relationship string             `json:"relationship"`
	Path         []RelationPathItem `json:"path"`
}
