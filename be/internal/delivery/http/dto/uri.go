package dto

type MemberIDUri struct {
	MemberID int `uri:"member_id" binding:"required,min=1"`
}

type SpouseIDUri struct {
	SpouseID int `uri:"spouse_id" binding:"required,min=1"`
}

type UserIDUri struct {
	UserID int `uri:"user_id" binding:"required,min=1"`
}

type CodeUri struct {
	Code string `uri:"code" binding:"required,min=2,max=10"`
}

type ProviderUri struct {
	Provider string `uri:"provider" binding:"required,min=2,max=20"`
}
