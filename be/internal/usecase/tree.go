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
	members, err := uc.memberRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("no members found")
	}

	spouseMap, err := uc.spouseRepo.GetAllSpouses(ctx)
	if err != nil {
		return nil, fmt.Errorf("get spouses: %w", err)
	}

	// Create member map
	memberMap := make(map[int]*domain.Member)
	for _, m := range members {
		memberMap[m.MemberID] = m
	}

	// If root is specified, build tree from that root
	if rootID != nil {
		// Validate root exists
		if _, exists := memberMap[*rootID]; !exists {
			return nil, fmt.Errorf("root member not found")
		}

		// Build tree with spouse support
		visited := make(map[int]bool)
		tree := uc.buildTreeWithSpouses(memberMap, spouseMap, *rootID, userRole, visited, nil, 0)
		return tree, nil
	}

	// Find all roots (members with no parents)
	roots := uc.findAllRoots(members)
	if len(roots) == 0 {
		return nil, fmt.Errorf("no root members found")
	}

	// If only one root, return it directly
	if len(roots) == 1 {
		visited := make(map[int]bool)
		tree := uc.buildTreeWithSpouses(memberMap, spouseMap, roots[0].MemberID, userRole, visited, nil, 0)
		return tree, nil
	}

	// Multiple roots: create a virtual root node containing all trees
	virtualRoot := &domain.MemberTreeNode{
		MemberWithComputed: domain.MemberWithComputed{
			Member: domain.Member{
				MemberID:    0, // Virtual root ID
				ArabicName:  "All Trees",
				EnglishName: "All Trees",
				Gender:      "M",
			},
			GenerationLevel: -1,
		},
		Children:     []*domain.MemberTreeNode{},
		SpouseNodes:  []*domain.MemberTreeNode{},
		SiblingNodes: []*domain.MemberTreeNode{},
	}

	// Build each disconnected tree and add as children
	for _, root := range roots {
		visited := make(map[int]bool)
		tree := uc.buildTreeWithSpouses(memberMap, spouseMap, root.MemberID, userRole, visited, nil, 0)
		if tree != nil {
			virtualRoot.Children = append(virtualRoot.Children, tree)
		}
	}

	return virtualRoot, nil
}

func (uc *treeUseCase) GetListView(ctx context.Context, rootID *int, userRole int) ([]*domain.MemberWithComputed, error) {
	members, err := uc.memberRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	spouseMap, err := uc.spouseRepo.GetAllSpouses(ctx)
	if err != nil {
		return nil, fmt.Errorf("get spouses: %w", err)
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

func (uc *treeUseCase) GetRelationTree(ctx context.Context, member1ID, member2ID int, userRole int) (*domain.MemberTreeNode, error) {
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

	// Validate members exist
	if _, exists := memberMap[member1ID]; !exists {
		return nil, fmt.Errorf("member1 not found")
	}
	if _, exists := memberMap[member2ID]; !exists {
		return nil, fmt.Errorf("member2 not found")
	}

	// Find path between members
	pathMemberIDs := uc.findPath(memberMap, member1ID, member2ID)
	if pathMemberIDs == nil {
		return nil, fmt.Errorf("no relation found between members")
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

	// Build tree with path highlighting
	visited := make(map[int]bool)
	tree := uc.buildTreeWithSpouses(memberMap, spouseMap, root.MemberID, userRole, visited, pathMembers, 0)
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

func (uc *treeUseCase) buildTreeWithSpouses(memberMap map[int]*domain.Member, spouseMap map[int][]int, rootID int, userRole int, visited map[int]bool, pathMembers map[int]bool, generationLevel int) *domain.MemberTreeNode {
	// Avoid circular references (spouse relationships)
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
		Children:     []*domain.MemberTreeNode{},
		SpouseNodes:  []*domain.MemberTreeNode{},
		SiblingNodes: []*domain.MemberTreeNode{},
		IsInPath:     pathMembers != nil && pathMembers[rootID],
	}

	// Apply privacy rules
	if root.Gender == "F" && userRole < domain.RoleSuperAdmin {
		node.DateOfBirth = nil
		node.DateOfDeath = nil
	}
	if root.Gender == "F" && userRole < domain.RoleAdmin {
		node.Picture = nil
	}

	// Add spouse nodes (only those under the current root)
	for _, spouseID := range spouseIDs {
		if !visited[spouseID] {
			spouseNode := uc.buildSpouseNode(memberMap, spouseMap, spouseID, userRole, pathMembers, generationLevel)
			if spouseNode != nil {
				node.SpouseNodes = append(node.SpouseNodes, spouseNode)
			}
		}
	}

	// Add sibling nodes (if this is not the tree root, i.e., generationLevel > 0)
	if generationLevel > 0 {
		for _, m := range memberMap {
			// Skip self and already visited
			if m.MemberID == rootID || visited[m.MemberID] {
				continue
			}
			// Check if sibling (same parents)
			isSibling := false
			if root.FatherID != nil && m.FatherID != nil && *root.FatherID == *m.FatherID {
				isSibling = true
			} else if root.MotherID != nil && m.MotherID != nil && *root.MotherID == *m.MotherID {
				isSibling = true
			}

			if isSibling {
				// Mark as visited to avoid duplicates
				visited[m.MemberID] = true

				// Build sibling node with their spouses
				siblingSpouseIDs := spouseMap[m.MemberID]
				siblingSpouses := uc.convertSpouseIDsToInfo(siblingSpouseIDs, memberMap)

				siblingNode := &domain.MemberTreeNode{
					MemberWithComputed: domain.MemberWithComputed{
						Member:          *m,
						IsMarried:       len(siblingSpouseIDs) > 0,
						Spouses:         siblingSpouses,
						GenerationLevel: generationLevel,
					},
					Children:     []*domain.MemberTreeNode{},
					SpouseNodes:  []*domain.MemberTreeNode{},
					SiblingNodes: []*domain.MemberTreeNode{},
					IsInPath:     pathMembers != nil && pathMembers[m.MemberID],
				}

				// Apply privacy rules for sibling
				if m.Gender == "F" && userRole < domain.RoleSuperAdmin {
					siblingNode.DateOfBirth = nil
					siblingNode.DateOfDeath = nil
				}
				if m.Gender == "F" && userRole < domain.RoleAdmin {
					siblingNode.Picture = nil
				}

				// Add sibling's spouses
				for _, siblingSpouseID := range siblingSpouseIDs {
					if !visited[siblingSpouseID] {
						siblingSpouseNode := uc.buildSpouseNode(memberMap, spouseMap, siblingSpouseID, userRole, pathMembers, generationLevel)
						if siblingSpouseNode != nil {
							siblingNode.SpouseNodes = append(siblingNode.SpouseNodes, siblingSpouseNode)
							visited[siblingSpouseID] = true
						}
					}
				}

				node.SiblingNodes = append(node.SiblingNodes, siblingNode)
			}
		}
	}

	// Find children (only from father to avoid duplication)
	for _, m := range memberMap {
		// Children branch from father node only, or mother if no father
		if m.FatherID != nil && *m.FatherID == rootID {
			child := uc.buildTreeWithSpouses(memberMap, spouseMap, m.MemberID, userRole, visited, pathMembers, generationLevel+1)
			if child != nil {
				node.Children = append(node.Children, child)
			}
		} else if m.FatherID == nil && m.MotherID != nil && *m.MotherID == rootID {
			// Only add if mother and father is unknown
			child := uc.buildTreeWithSpouses(memberMap, spouseMap, m.MemberID, userRole, visited, pathMembers, generationLevel+1)
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

// buildSpouseNode creates a spouse node without recursing into their children (to avoid duplication)
func (uc *treeUseCase) buildSpouseNode(memberMap map[int]*domain.Member, spouseMap map[int][]int, spouseID int, userRole int, pathMembers map[int]bool, generationLevel int) *domain.MemberTreeNode {
	spouse := memberMap[spouseID]
	if spouse == nil {
		return nil
	}

	spouseSpouseIDs := spouseMap[spouseID]
	spouseSpouses := uc.convertSpouseIDsToInfo(spouseSpouseIDs, memberMap)

	node := &domain.MemberTreeNode{
		MemberWithComputed: domain.MemberWithComputed{
			Member:          *spouse,
			IsMarried:       len(spouseSpouseIDs) > 0,
			Spouses:         spouseSpouses,
			GenerationLevel: generationLevel,
		},
		Children:    []*domain.MemberTreeNode{}, // Spouses don't show children (they're shown under main node)
		SpouseNodes: []*domain.MemberTreeNode{},
		IsInPath:    pathMembers != nil && pathMembers[spouseID],
	}

	// Apply privacy rules
	if spouse.Gender == "F" && userRole < domain.RoleSuperAdmin {
		node.DateOfBirth = nil
		node.DateOfDeath = nil
	}
	if spouse.Gender == "F" && userRole < domain.RoleAdmin {
		node.Picture = nil
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
