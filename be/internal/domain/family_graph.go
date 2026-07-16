package domain

import "time"

type FamilyUnit struct {
	FamilyUnitID     int        `json:"family_unit_id"`
	TreeID           int        `json:"tree_id"`
	RelationshipType string     `json:"relationship_type"`
	Status           string     `json:"status"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	PartnerIDs       []int      `json:"partner_ids"`
	ChildIDs         []int      `json:"child_ids"`
	ChildRelations   map[int]string
}

type FamilyGraphPerson struct {
	MemberWithComputed
	ParentFamilyUnitIDs  []int `json:"parent_family_unit_ids,omitempty"`
	PartnerFamilyUnitIDs []int `json:"partner_family_unit_ids,omitempty"`
	IsReferenceCandidate bool  `json:"is_reference_candidate"`
	IsInPath             bool  `json:"is_in_path"`
}

type FamilyGraphEdge struct {
	EdgeID       string `json:"edge_id"`
	SourceID     string `json:"source_id"`
	TargetID     string `json:"target_id"`
	Type         string `json:"type"`
	RelationType string `json:"relation_type,omitempty"`
	Status       string `json:"status,omitempty"`
	IsInPath     bool   `json:"is_in_path"`
}

type FamilyGraphReference struct {
	ReferenceID  string `json:"reference_id"`
	PersonID     int    `json:"person_id"`
	FamilyUnitID int    `json:"family_unit_id"`
	Reason       string `json:"reason"`
}

type FamilyGraph struct {
	People            []*FamilyGraphPerson   `json:"people"`
	FamilyUnits       []*FamilyUnit          `json:"family_units"`
	Edges             []FamilyGraphEdge      `json:"edges"`
	References        []FamilyGraphReference `json:"references"`
	PathPersonIDs     []int                  `json:"path_person_ids,omitempty"`
	PathFamilyUnitIDs []int                  `json:"path_family_unit_ids,omitempty"`
}
