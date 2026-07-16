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

type FamilyGraphPersonResponse struct {
	Member               MemberResponse `json:"member"`
	ParentFamilyUnitIDs  []int          `json:"parent_family_unit_ids,omitempty"`
	PartnerFamilyUnitIDs []int          `json:"partner_family_unit_ids,omitempty"`
	IsReferenceCandidate bool           `json:"is_reference_candidate"`
	IsInPath             bool           `json:"is_in_path"`
}

type FamilyGraphUnitResponse struct {
	FamilyUnitID     int    `json:"family_unit_id"`
	TreeID           int    `json:"tree_id"`
	RelationshipType string `json:"relationship_type"`
	Status           string `json:"status"`
	StartDate        *Date  `json:"start_date"`
	EndDate          *Date  `json:"end_date"`
	PartnerIDs       []int  `json:"partner_ids"`
	ChildIDs         []int  `json:"child_ids"`
}

type FamilyGraphEdgeResponse struct {
	EdgeID       string `json:"edge_id"`
	SourceID     string `json:"source_id"`
	TargetID     string `json:"target_id"`
	Type         string `json:"type"`
	RelationType string `json:"relation_type,omitempty"`
	Status       string `json:"status,omitempty"`
	IsInPath     bool   `json:"is_in_path"`
}

type FamilyGraphReferenceResponse struct {
	ReferenceID  string `json:"reference_id"`
	PersonID     int    `json:"person_id"`
	FamilyUnitID int    `json:"family_unit_id"`
	Reason       string `json:"reason"`
}

type FamilyGraphResponse struct {
	People            []FamilyGraphPersonResponse    `json:"people"`
	FamilyUnits       []FamilyGraphUnitResponse      `json:"family_units"`
	Edges             []FamilyGraphEdgeResponse      `json:"edges"`
	References        []FamilyGraphReferenceResponse `json:"references"`
	PathPersonIDs     []int                          `json:"path_person_ids,omitempty"`
	PathFamilyUnitIDs []int                          `json:"path_family_unit_ids,omitempty"`
}
