package service

import (
	"context"

	"family-tree-backend/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TreeService struct {
	queries *db.Queries
}

func NewTreeService(pool *pgxpool.Pool) *TreeService {
	return &TreeService{queries: db.New(pool)}
}

func (s *TreeService) CreateTree(ctx context.Context, ownerID uuid.UUID, title string) (db.Tree, error) {
	return s.queries.CreateTree(ctx, db.CreateTreeParams{
		OwnerID: ownerID,
		Title:   title,
	})
}

func (s *TreeService) ListTrees(ctx context.Context, ownerID uuid.UUID) ([]db.Tree, error) {
	return s.queries.ListTrees(ctx, ownerID)
}

func (s *TreeService) GetTree(ctx context.Context, treeID, ownerID uuid.UUID) (db.Tree, error) {
	return s.queries.GetTree(ctx, db.GetTreeParams{
		ID:      treeID,
		OwnerID: ownerID,
	})
}

func (s *TreeService) GetTreeByID(ctx context.Context, treeID uuid.UUID) (db.Tree, error) {
	return s.queries.GetTreeByID(ctx, treeID)
}

func (s *TreeService) DeleteTree(ctx context.Context, treeID, ownerID uuid.UUID) error {
	return s.queries.DeleteTree(ctx, db.DeleteTreeParams{
		ID:      treeID,
		OwnerID: ownerID,
	})
}
