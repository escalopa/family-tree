package repository

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FamilyGraphRepository struct {
	db *pgxpool.Pool
}

func NewFamilyGraphRepository(db *pgxpool.Pool) *FamilyGraphRepository {
	return &FamilyGraphRepository{db: db}
}

func (r *FamilyGraphRepository) ListFamilyUnitsByTreeID(ctx context.Context, treeID int) ([]*domain.FamilyUnit, error) {
	query := `
		SELECT
			fu.family_unit_id,
			fu.tree_id,
			fu.relationship_type,
			fu.status,
			fu.start_date,
			fu.end_date,
			COALESCE(
				array_agg(DISTINCT fup.person_id ORDER BY fup.person_id)
					FILTER (WHERE fup.person_id IS NOT NULL),
				'{}'
			) AS partner_ids,
			COALESCE(
				array_agg(DISTINCT fuc.child_person_id ORDER BY fuc.child_person_id)
					FILTER (WHERE fuc.child_person_id IS NOT NULL),
				'{}'
			) AS child_ids
		FROM family_units fu
		LEFT JOIN family_unit_partners fup ON fup.family_unit_id = fu.family_unit_id
		LEFT JOIN family_unit_children fuc ON fuc.family_unit_id = fu.family_unit_id
		WHERE fu.tree_id = $1
		  AND fu.deleted_at IS NULL
		GROUP BY fu.family_unit_id, fu.tree_id, fu.relationship_type, fu.status, fu.start_date, fu.end_date
		ORDER BY
			COALESCE(fu.start_date, DATE '9999-12-31'),
			fu.family_unit_id
	`

	rows, err := r.db.Query(ctx, query, treeID)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer rows.Close()

	units := make([]*domain.FamilyUnit, 0)
	for rows.Next() {
		unit := &domain.FamilyUnit{}
		if err := rows.Scan(
			&unit.FamilyUnitID,
			&unit.TreeID,
			&unit.RelationshipType,
			&unit.Status,
			&unit.StartDate,
			&unit.EndDate,
			&unit.PartnerIDs,
			&unit.ChildIDs,
		); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		unit.ChildRelations = make(map[int]string)
		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	if len(units) == 0 {
		return units, nil
	}

	relationQuery := `
		SELECT family_unit_id, child_person_id, relation_type
		FROM family_unit_children
		WHERE family_unit_id = ANY($1)
	`
	unitIDs := make([]int, 0, len(units))
	unitByID := make(map[int]*domain.FamilyUnit, len(units))
	for _, unit := range units {
		unitIDs = append(unitIDs, unit.FamilyUnitID)
		unitByID[unit.FamilyUnitID] = unit
	}

	relationRows, err := r.db.Query(ctx, relationQuery, unitIDs)
	if err != nil {
		return nil, domain.NewDatabaseError(err)
	}
	defer relationRows.Close()

	for relationRows.Next() {
		var unitID, childID int
		var relationType string
		if err := relationRows.Scan(&unitID, &childID, &relationType); err != nil {
			return nil, domain.NewDatabaseError(err)
		}
		if unit := unitByID[unitID]; unit != nil {
			unit.ChildRelations[childID] = relationType
		}
	}
	if err := relationRows.Err(); err != nil {
		return nil, domain.NewDatabaseError(err)
	}

	return units, nil
}
