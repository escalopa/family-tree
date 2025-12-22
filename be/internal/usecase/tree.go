package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/escalopa/family-tree/internal/domain"
)

type treeUseCase struct {
	memberRepo MemberRepository
	spouseRepo SpouseRepository
}

func NewTreeUseCase(
	memberRepo MemberRepository,
	spouseRepo SpouseRepository,
) *treeUseCase {
	return &treeUseCase{
		memberRepo: memberRepo,
		spouseRepo: spouseRepo,
	}
}

func (uc *treeUseCase) GetTree(ctx context.Context, rootID *int, userRole int) (*domain.MemberTreeNode, error) {
	// Get all members
	members, err := uc.memberRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("no members found")
	}

	// Get spouse relationships
	spouseMap, err := uc.spouseRepo.GetAllSpouses(ctx)
	if err != nil {
		return nil, fmt.Errorf("get spouses: %w", err)
	}

	// Create member map
	memberMap := make(map[int]*domain.Member)
	for _, m := range members {
		memberMap[m.MemberID] = m
	}

	// Find root if not specified
	if rootID == nil {
		root := uc.findOldestRoot(members)
		rootID = &root.MemberID
	}

	// Validate root exists
	if _, exists := memberMap[*rootID]; !exists {
		return nil, fmt.Errorf("root member not found")
	}

	// Build tree
	tree := uc.buildTree(memberMap, spouseMap, *rootID, userRole)
	return tree, nil
}

func (uc *treeUseCase) GetListView(ctx context.Context, rootID *int, userRole int) ([]*domain.MemberWithComputed, error) {
	// Get all members
	members, err := uc.memberRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	// Get spouse relationships
	spouseMap, err := uc.spouseRepo.GetAllSpouses(ctx)
	if err != nil {
		return nil, fmt.Errorf("get spouses: %w", err)
	}

	// Create member map for spouse lookup
	memberMap := make(map[int]*domain.Member)
	for _, m := range members {
		memberMap[m.MemberID] = m
	}

	// Convert to MemberWithComputed
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

	// Sort by date of birth (oldest first)
	sort.Slice(result, func(i, j int) bool {
		dateI := result[i].DateOfBirth
		dateJ := result[j].DateOfBirth

		// Handle nil dates (put them at the end)
		if dateI == nil && dateJ == nil {
			return result[i].MemberID < result[j].MemberID
		}
		if dateI == nil {
			return false
		}
		if dateJ == nil {
			return true
		}

		// Compare dates
		if dateI.Equal(*dateJ) {
			return result[i].MemberID < result[j].MemberID
		}
		return dateI.Before(*dateJ)
	})

	return result, nil
}

func (uc *treeUseCase) GetRelationPath(ctx context.Context, member1ID, member2ID int, userRole int) ([]*domain.MemberWithComputed, error) {
	// Get all members
	members, err := uc.memberRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	// Create member map
	memberMap := make(map[int]*domain.Member)
	for _, m := range members {
		memberMap[m.MemberID] = m
	}

	// Find path (simplified BFS approach)
	path := uc.findPath(memberMap, member1ID, member2ID)
	if path == nil {
		return nil, fmt.Errorf("no relation found between members")
	}

	// Convert to MemberWithComputed
	var result []*domain.MemberWithComputed
	spouseMap, _ := uc.spouseRepo.GetAllSpouses(ctx)
	for _, memberID := range path {
		m := memberMap[memberID]
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

func (uc *treeUseCase) findOldestRoot(members []*domain.Member) *domain.Member {
	// Find member with no parents and oldest birth date
	var oldest *domain.Member
	for _, m := range members {
		if m.FatherID == nil && m.MotherID == nil {
			if oldest == nil {
				oldest = m
			} else if m.DateOfBirth != nil && (oldest.DateOfBirth == nil || m.DateOfBirth.Before(*oldest.DateOfBirth)) {
				oldest = m
			}
		}
	}
	if oldest != nil {
		return oldest
	}
	// Fallback: return first member
	return members[0]
}

func (uc *treeUseCase) buildTree(memberMap map[int]*domain.Member, spouseMap map[int][]int, rootID int, userRole int) *domain.MemberTreeNode {
	root := memberMap[rootID]
	if root == nil {
		return nil
	}

	spouseIDs := spouseMap[rootID]
	spouses := uc.convertSpouseIDsToInfo(spouseIDs, memberMap)

	node := &domain.MemberTreeNode{
		MemberWithComputed: domain.MemberWithComputed{
			Member:    *root,
			IsMarried: len(spouseIDs) > 0,
			Spouses:   spouses,
		},
		Children: []*domain.MemberTreeNode{},
	}

	// Apply privacy rules
	if root.Gender == "F" && userRole < domain.RoleSuperAdmin {
		node.DateOfBirth = nil
		node.DateOfDeath = nil
	}
	if root.Gender == "F" && userRole < domain.RoleAdmin {
		node.Picture = nil
	}

	// Find children
	for _, m := range memberMap {
		if (m.FatherID != nil && *m.FatherID == rootID) || (m.MotherID != nil && *m.MotherID == rootID) {
			child := uc.buildTree(memberMap, spouseMap, m.MemberID, userRole)
			if child != nil {
				node.Children = append(node.Children, child)
			}
		}
	}

	// Sort children by birth date (oldest first)
	sort.Slice(node.Children, func(i, j int) bool {
		dateI := node.Children[i].DateOfBirth
		dateJ := node.Children[j].DateOfBirth

		// Handle nil dates (put them at the end)
		if dateI == nil && dateJ == nil {
			return node.Children[i].MemberID < node.Children[j].MemberID
		}
		if dateI == nil {
			return false
		}
		if dateJ == nil {
			return true
		}

		// Compare dates
		if dateI.Equal(*dateJ) {
			return node.Children[i].MemberID < node.Children[j].MemberID
		}
		return dateI.Before(*dateJ)
	})

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
				ArabicName:   spouse.ArabicName,
				EnglishName:  spouse.EnglishName,
				Gender:       spouse.Gender,
				Picture:      spouse.Picture,
				MarriageDate: nil, // Will be populated from spouse relationship if needed
				DivorceDate:  nil, // Will be populated from spouse relationship if needed
			})
		}
	}
	return spouses
}
