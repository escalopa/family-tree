package usecase

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/escalopa/family-tree-api/internal/repository"
)

type TreeNode struct {
	Member   *domain.MemberWithDetails
	Children []*TreeNode
}

type TreeResponse struct {
	Root     *TreeNode
	Metadata TreeMetadata
}

type TreeMetadata struct {
	TotalMembers  int
	MaxGeneration int
	RootMemberID  int
}

type RelationPath struct {
	MemberID     int
	Name         string
	RelationType string
}

type RelationResponse struct {
	Member1      MemberInfo
	Member2      MemberInfo
	Relationship string
	Path         []RelationPath
}

type MemberInfo struct {
	MemberID int
	Name     string
}

type TreeUseCase interface {
	GetTreeFromRoot(ctx context.Context, rootID *int) (*TreeResponse, error)
	GetRelationship(ctx context.Context, member1ID, member2ID int) (*RelationResponse, error)
}

type treeUseCase struct {
	memberRepo repository.MemberRepository
	spouseRepo repository.SpouseRepository
}

func NewTreeUseCase(
	memberRepo repository.MemberRepository,
	spouseRepo repository.SpouseRepository,
) TreeUseCase {
	return &treeUseCase{
		memberRepo: memberRepo,
		spouseRepo: spouseRepo,
	}
}

func (uc *treeUseCase) GetTreeFromRoot(ctx context.Context, rootID *int) (*TreeResponse, error) {
	// TODO: Implementation
	return nil, nil
}

func (uc *treeUseCase) GetRelationship(ctx context.Context, member1ID, member2ID int) (*RelationResponse, error) {
	// TODO: Implementation
	return nil, nil
}
