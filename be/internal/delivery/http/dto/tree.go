package dto

type TreeQuery struct {
	RootID *int   `form:"root"`
	Style  string `form:"style" binding:"required,oneof=tree list"`
}

type RelationQuery struct {
	Member1ID int `form:"member1" binding:"required,min=1"`
	Member2ID int `form:"member2" binding:"required,min=1"`
}

type TreeNodeResponse struct {
	Member   MemberResponse      `json:"member"`
	Children []*TreeNodeResponse `json:"children,omitempty"`
	IsInPath bool                `json:"is_in_path,omitempty"`
}

type TreeResponse struct {
	Roots []*TreeNodeResponse `json:"roots"`
}
