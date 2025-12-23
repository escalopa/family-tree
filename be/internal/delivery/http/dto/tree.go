package dto

type TreeQuery struct {
	RootID *int   `form:"root"`
	Style  string `form:"style"` // "tree" or "list"
}

type RelationQuery struct {
	Member1ID int `form:"member1" binding:"required"`
	Member2ID int `form:"member2" binding:"required"`
}

type TreeNodeResponse struct {
	Member       MemberResponse      `json:"member"`
	Children     []*TreeNodeResponse `json:"children,omitempty"`
	SpouseNodes  []*TreeNodeResponse `json:"spouse_nodes,omitempty"`
	SiblingNodes []*TreeNodeResponse `json:"sibling_nodes,omitempty"`
	IsInPath     bool                `json:"is_in_path,omitempty"`
}
