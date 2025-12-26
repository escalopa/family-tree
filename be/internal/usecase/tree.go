package usecase

import (
	"context"
	"sort"

	"github.com/escalopa/family-tree/internal/domain"
)

type (
	treeUseCaseRepo struct {
		member MemberRepository
		spouse SpouseRepository
	}

	treeUseCase struct {
		repo treeUseCaseRepo
	}
)

func NewTreeUseCase(
	memberRepo MemberRepository,
	spouseRepo SpouseRepository,
) *treeUseCase {
	return &treeUseCase{
		repo: treeUseCaseRepo{
			member: memberRepo,
			spouse: spouseRepo,
		},
	}
}

func (uc *treeUseCase) Get(ctx context.Context, rootID *int, userRole int) (*domain.MemberTreeNode, error) {
	members, err := uc.repo.member.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, nil
	}

	spouseMap, err := uc.repo.spouse.GetAllSpouses(ctx)
	if err != nil {
		return nil, err
	}

	memberMap := make(map[int]*domain.Member)
	for _, m := range members {
		memberMap[m.MemberID] = m
	}

	if rootID != nil {
		if _, exists := memberMap[*rootID]; !exists {
			return nil, domain.NewNotFoundError("member")
		}

		// Build tree - generation starts at 1
		visited := make(map[int]bool)
		tree := uc.buildTree(memberMap, spouseMap, *rootID, userRole, visited, nil, 1)
		return tree, nil
	}

	// Find all roots (members with no parents)
	roots := uc.findAllRoots(members)
	if len(roots) == 0 {
		return nil, nil
	}

	// Return the first root directly (single tree) - generation starts at 1
	visited := make(map[int]bool)
	tree := uc.buildTree(memberMap, spouseMap, roots[0].MemberID, userRole, visited, nil, 1)
	return tree, nil
}

func (uc *treeUseCase) List(ctx context.Context, rootID *int, userRole int) ([]*domain.MemberWithComputed, error) {
	members, err := uc.repo.member.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	spouseMap, err := uc.repo.spouse.GetAllSpouses(ctx)
	if err != nil {
		return nil, err
	}

	memberMap := make(map[int]*domain.Member)
	for _, m := range members {
		memberMap[m.MemberID] = m
	}

	var result []*domain.MemberWithComputed
	for _, m := range members {
		spouseIDs := spouseMap[m.MemberID]
		spouses := uc.convertSpouseIDsToInfo(spouseIDs, memberMap)

		computed := &domain.MemberWithComputed{
			Member:    *m,
			IsMarried: len(spouseIDs) > 0,
			Spouses:   spouses,
		}

		// Apply privacy rules
		if m.Gender == "F" && userRole < domain.RoleSuperAdmin {
			computed.DateOfBirth = nil
			computed.DateOfDeath = nil
		}
		if m.Gender == "F" && userRole < domain.RoleAdmin {
			computed.Picture = nil
		}

		result = append(result, computed)
	}

	return result, nil
}

func (uc *treeUseCase) GetRelation(ctx context.Context, member1ID, member2ID int, userRole int) (*domain.MemberTreeNode, error) {
	members, err := uc.repo.member.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, nil
	}

	// Get spouse relationships
	spouseMap, err := uc.repo.spouse.GetAllSpouses(ctx)
	if err != nil {
		return nil, err
	}

	// Create member map
	memberMap := make(map[int]*domain.Member)
	for _, m := range members {
		memberMap[m.MemberID] = m
	}

	// Validate members exist
	if _, exists := memberMap[member1ID]; !exists {
		return nil, domain.NewNotFoundError("member")
	}
	if _, exists := memberMap[member2ID]; !exists {
		return nil, domain.NewNotFoundError("member")
	}

	// Find path between members
	pathMemberIDs := uc.findPath(memberMap, member1ID, member2ID)
	if pathMemberIDs == nil {
		return nil, domain.NewValidationError("error.member.no_relation")
	}

	// Create a map for quick lookup of path members
	pathMembers := make(map[int]bool)
	for _, id := range pathMemberIDs {
		pathMembers[id] = true
	}

	// Find common root (oldest ancestor)
	root := uc.findCommonRoot(memberMap, member1ID, member2ID)
	if root == nil {
		root = uc.findOldestRoot(members)
	}

	// Build tree with path highlighting - generation starts at 1
	visited := make(map[int]bool)
	tree := uc.buildTree(memberMap, spouseMap, root.MemberID, userRole, visited, pathMembers, 1)
	return tree, nil
}

func (uc *treeUseCase) findCommonRoot(memberMap map[int]*domain.Member, member1ID, member2ID int) *domain.Member {
	// Get all ancestors of member1
	ancestors1 := uc.getAncestors(memberMap, member1ID)

	// For member2, walk up the tree and find first common ancestor
	current := member2ID
	for current != 0 {
		if ancestors1[current] {
			return memberMap[current]
		}
		member := memberMap[current]
		if member.FatherID != nil {
			current = *member.FatherID
		} else if member.MotherID != nil {
			current = *member.MotherID
		} else {
			break
		}
	}

	// If no common ancestor, return oldest root
	return nil
}

func (uc *treeUseCase) getAncestors(memberMap map[int]*domain.Member, memberID int) map[int]bool {
	ancestors := make(map[int]bool)
	ancestors[memberID] = true

	current := memberID
	for current != 0 {
		member := memberMap[current]
		if member == nil {
			break
		}
		if member.FatherID != nil {
			ancestors[*member.FatherID] = true
			current = *member.FatherID
		} else if member.MotherID != nil {
			ancestors[*member.MotherID] = true
			current = *member.MotherID
		} else {
			break
		}
	}

	return ancestors
}

func (uc *treeUseCase) findAllRoots(members []*domain.Member) []*domain.Member {
	// Find all members with no parents
	var roots []*domain.Member
	for _, m := range members {
		if m.FatherID == nil && m.MotherID == nil {
			roots = append(roots, m)
		}
	}

	// Sort by birth date (oldest first)
	sort.Slice(roots, func(i, j int) bool {
		dateI := roots[i].DateOfBirth
		dateJ := roots[j].DateOfBirth

		// Handle nil dates (put them at the end)
		if dateI == nil && dateJ == nil {
			return roots[i].MemberID < roots[j].MemberID
		}
		if dateI == nil {
			return false
		}
		if dateJ == nil {
			return true
		}

		// Compare dates
		if dateI.Equal(*dateJ) {
			return roots[i].MemberID < roots[j].MemberID
		}
		return dateI.Before(*dateJ)
	})

	return roots
}

func (uc *treeUseCase) findOldestRoot(members []*domain.Member) *domain.Member {
	roots := uc.findAllRoots(members)
	if len(roots) > 0 {
		return roots[0]
	}
	// Fallback: return first member
	return members[0]
}

func (uc *treeUseCase) buildTree(memberMap map[int]*domain.Member, spouseMap map[int][]int, rootID int, userRole int, visited map[int]bool, pathMembers map[int]bool, generationLevel int) *domain.MemberTreeNode {
	// Avoid circular references
	if visited[rootID] {
		return nil
	}
	visited[rootID] = true

	root := memberMap[rootID]
	if root == nil {
		return nil
	}

	spouseIDs := spouseMap[rootID]
	spouses := uc.convertSpouseIDsToInfo(spouseIDs, memberMap)

	node := &domain.MemberTreeNode{
		MemberWithComputed: domain.MemberWithComputed{
			Member:          *root,
			IsMarried:       len(spouseIDs) > 0,
			Spouses:         spouses,
			GenerationLevel: generationLevel,
		},
		Children: []*domain.MemberTreeNode{},
		IsInPath: pathMembers != nil && pathMembers[rootID],
	}

	// Apply privacy rules
	if root.Gender == "F" && userRole < domain.RoleSuperAdmin {
		node.DateOfBirth = nil
		node.DateOfDeath = nil
	}
	if root.Gender == "F" && userRole < domain.RoleAdmin {
		node.Picture = nil
	}

	// Find ALL children of this member (including children with spouses)
	// Group children by their other parent (spouse) to handle multiple marriages
	childrenBySpouse := make(map[int][]*domain.Member) // spouseID -> children

	for _, m := range memberMap {
		// Check if this member is a child
		var isChild bool
		var otherParentID int

		if root.Gender == "M" {
			// Root is father - check if m is his child
			if m.FatherID != nil && *m.FatherID == rootID {
				isChild = true
				if m.MotherID != nil {
					otherParentID = *m.MotherID
				}
			}
		} else {
			// Root is mother - check if m is her child
			if m.MotherID != nil && *m.MotherID == rootID {
				isChild = true
				if m.FatherID != nil {
					otherParentID = *m.FatherID
				}
			}
		}

		if isChild {
			childrenBySpouse[otherParentID] = append(childrenBySpouse[otherParentID], m)
		}
	}

	// Build children nodes (all children appear as direct children)
	var allChildren []*domain.Member
	for _, children := range childrenBySpouse {
		allChildren = append(allChildren, children...)
	}

	// Sort children by birth date
	sort.Slice(allChildren, func(i, j int) bool {
		dateI := allChildren[i].DateOfBirth
		dateJ := allChildren[j].DateOfBirth

		if dateI == nil && dateJ == nil {
			return allChildren[i].MemberID < allChildren[j].MemberID
		}
		if dateI == nil {
			return false
		}
		if dateJ == nil {
			return true
		}

		if dateI.Equal(*dateJ) {
			return allChildren[i].MemberID < allChildren[j].MemberID
		}
		return dateI.Before(*dateJ)
	})

	// Recursively build child nodes
	for _, childMember := range allChildren {
		child := uc.buildTree(memberMap, spouseMap, childMember.MemberID, userRole, visited, pathMembers, generationLevel+1)
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	return node
}

func (uc *treeUseCase) findPath(memberMap map[int]*domain.Member, startID, endID int) []int {
	// Simple BFS to find path
	if startID == endID {
		return []int{startID}
	}

	visited := make(map[int]bool)
	parent := make(map[int]int)
	queue := []int{startID}
	visited[startID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		member := memberMap[current]
		if member == nil {
			continue
		}

		// Check neighbors (parents and children)
		neighbors := []int{}
		if member.FatherID != nil {
			neighbors = append(neighbors, *member.FatherID)
		}
		if member.MotherID != nil {
			neighbors = append(neighbors, *member.MotherID)
		}
		// Find children
		for _, m := range memberMap {
			if (m.FatherID != nil && *m.FatherID == current) || (m.MotherID != nil && *m.MotherID == current) {
				neighbors = append(neighbors, m.MemberID)
			}
		}

		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)

				if neighbor == endID {
					// Reconstruct path
					path := []int{}
					for n := endID; n != startID; n = parent[n] {
						path = append([]int{n}, path...)
					}
					path = append([]int{startID}, path...)
					return path
				}
			}
		}
	}

	return nil
}

func (uc *treeUseCase) convertSpouseIDsToInfo(spouseIDs []int, memberMap map[int]*domain.Member) []domain.SpouseWithMemberInfo {
	if len(spouseIDs) == 0 {
		return nil
	}

	spouses := make([]domain.SpouseWithMemberInfo, 0, len(spouseIDs))
	for _, spouseID := range spouseIDs {
		if spouse, exists := memberMap[spouseID]; exists {
			spouses = append(spouses, domain.SpouseWithMemberInfo{
				MemberID:     spouse.MemberID,
				Names:        spouse.Names,
				Gender:       spouse.Gender,
				Picture:      spouse.Picture,
				MarriageDate: nil, // Will be populated from spouse relationship if needed
				DivorceDate:  nil, // Will be populated from spouse relationship if needed
			})
		}
	}
	return spouses
}
