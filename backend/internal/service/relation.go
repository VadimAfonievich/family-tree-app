package service

import (
	"context"
	"fmt"

	"family-tree-backend/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RelationService struct {
	queries *db.Queries
}

func NewRelationService(pool *pgxpool.Pool) *RelationService {
	return &RelationService{queries: db.New(pool)}
}

func (s *RelationService) CreateRelation(ctx context.Context, treeID uuid.UUID, params CreateRelationParams) (db.Relation, error) {
	// Verify tree exists
	_, err := s.queries.GetTreeByID(ctx, treeID)
	if err != nil {
		return db.Relation{}, fmt.Errorf("tree not found: %w", err)
	}

	// Verify both persons exist and belong to the tree
	_, err = s.queries.GetPersonInTree(ctx, db.GetPersonInTreeParams{
		ID:     params.Person1ID,
		TreeID: treeID,
	})
	if err != nil {
		return db.Relation{}, fmt.Errorf("person1 not found in tree: %w", err)
	}

	_, err = s.queries.GetPersonInTree(ctx, db.GetPersonInTreeParams{
		ID:     params.Person2ID,
		TreeID: treeID,
	})
	if err != nil {
		return db.Relation{}, fmt.Errorf("person2 not found in tree: %w", err)
	}

	return s.queries.CreateRelation(ctx, db.CreateRelationParams{
		TreeID:       treeID,
		Person1ID:    params.Person1ID,
		Person2ID:    params.Person2ID,
		RelationType: params.RelationType,
	})
}

func (s *RelationService) ListRelations(ctx context.Context, treeID uuid.UUID) ([]db.Relation, error) {
	return s.queries.ListRelationsByTree(ctx, treeID)
}

func (s *RelationService) GetRelation(ctx context.Context, id uuid.UUID) (db.Relation, error) {
	return s.queries.GetRelation(ctx, id)
}

func (s *RelationService) DeleteRelation(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteRelation(ctx, id)
}

type CreateRelationParams struct {
	Person1ID    uuid.UUID
	Person2ID    uuid.UUID
	RelationType string
}
