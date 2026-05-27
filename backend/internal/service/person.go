package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"family-tree-backend/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PersonService struct {
	queries *db.Queries
}

func NewPersonService(pool *pgxpool.Pool) *PersonService {
	return &PersonService{queries: db.New(pool)}
}

func (s *PersonService) CreatePerson(ctx context.Context, treeID uuid.UUID, params CreatePersonParams) (db.Person, error) {
	// Verify tree exists
	_, err := s.queries.GetTreeByID(ctx, treeID)
	if err != nil {
		return db.Person{}, fmt.Errorf("tree not found: %w", err)
	}

	var birthDate, deathDate sql.NullTime
	if !params.BirthDate.IsZero() {
		birthDate = sql.NullTime{Time: params.BirthDate, Valid: true}
	}
	if !params.DeathDate.IsZero() {
		deathDate = sql.NullTime{Time: params.DeathDate, Valid: true}
	}

	return s.queries.CreatePerson(ctx, db.CreatePersonParams{
		TreeID:    treeID,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		Gender:    params.Gender,
		BirthDate: birthDate,
		DeathDate: deathDate,
		PhotoURL:  params.PhotoURL,
	})
}

func (s *PersonService) GetPerson(ctx context.Context, id uuid.UUID) (db.Person, error) {
	return s.queries.GetPerson(ctx, id)
}

func (s *PersonService) ListPersons(ctx context.Context, treeID uuid.UUID) ([]db.Person, error) {
	return s.queries.ListPersonsByTree(ctx, treeID)
}

func (s *PersonService) UpdatePerson(ctx context.Context, id uuid.UUID, params UpdatePersonParams) (db.Person, error) {
	var lastName, photoURL sql.NullString
	if params.LastName != "" {
		lastName = sql.NullString{String: params.LastName, Valid: true}
	}
	if params.PhotoURL != "" {
		photoURL = sql.NullString{String: params.PhotoURL, Valid: true}
	}

	var birthDate, deathDate sql.NullTime
	if params.BirthDate != nil {
		birthDate = sql.NullTime{Time: *params.BirthDate, Valid: true}
	}
	if params.DeathDate != nil {
		deathDate = sql.NullTime{Time: *params.DeathDate, Valid: true}
	}

	return s.queries.UpdatePerson(ctx, db.UpdatePersonParams{
		ID:        id,
		FirstName: params.FirstName,
		LastName:  lastName,
		Gender:    params.Gender,
		BirthDate: birthDate,
		DeathDate: deathDate,
		PhotoURL:  photoURL,
	})
}

func (s *PersonService) DeletePerson(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeletePerson(ctx, id)
}

type CreatePersonParams struct {
	FirstName string
	LastName  string
	Gender    string
	BirthDate time.Time
	DeathDate time.Time
	PhotoURL  string
}

type UpdatePersonParams struct {
	FirstName string
	LastName  string
	Gender    string
	BirthDate *time.Time
	DeathDate *time.Time
	PhotoURL  string
}
