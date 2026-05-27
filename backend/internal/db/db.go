package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

type User struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
}

type Tree struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   uuid.UUID `json:"owner_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type Person struct {
	ID        uuid.UUID  `json:"id"`
	TreeID    uuid.UUID  `json:"tree_id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Gender    string     `json:"gender"`
	BirthDate *time.Time `json:"birth_date"`
	DeathDate *time.Time `json:"death_date"`
	PhotoURL  string     `json:"photo_url"`
	CreatedAt time.Time  `json:"created_at"`
}

type Relation struct {
	ID           uuid.UUID `json:"id"`
	TreeID       uuid.UUID `json:"tree_id"`
	Person1ID    uuid.UUID `json:"person1_id"`
	Person2ID    uuid.UUID `json:"person2_id"`
	RelationType string    `json:"relation_type"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateUserParams struct {
	TelegramID int64
	Username   string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	return scanUser(q.pool.QueryRow(ctx, `
		INSERT INTO users (telegram_id, username)
		VALUES ($1, $2)
		ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username
		RETURNING id, telegram_id, COALESCE(username, ''), created_at
	`, arg.TelegramID, arg.Username))
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	return scanUser(q.pool.QueryRow(ctx, `
		SELECT id, telegram_id, COALESCE(username, ''), created_at FROM users WHERE id = $1
	`, id))
}

func scanUser(row pgx.Row) (User, error) {
	var user User
	err := row.Scan(&user.ID, &user.TelegramID, &user.Username, &user.CreatedAt)
	return user, err
}

type CreateTreeParams struct {
	OwnerID uuid.UUID
	Title   string
}

type GetTreeParams struct {
	ID      uuid.UUID
	OwnerID uuid.UUID
}

type DeleteTreeParams struct {
	ID      uuid.UUID
	OwnerID uuid.UUID
}

func (q *Queries) CreateTree(ctx context.Context, arg CreateTreeParams) (Tree, error) {
	return scanTree(q.pool.QueryRow(ctx, `
		INSERT INTO trees (owner_id, title)
		VALUES ($1, $2)
		RETURNING id, owner_id, title, created_at
	`, arg.OwnerID, arg.Title))
}

func (q *Queries) ListTrees(ctx context.Context, ownerID uuid.UUID) ([]Tree, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, owner_id, title, created_at
		FROM trees
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trees []Tree
	for rows.Next() {
		tree, err := scanTree(rows)
		if err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}
	return trees, rows.Err()
}

func (q *Queries) GetTree(ctx context.Context, arg GetTreeParams) (Tree, error) {
	return scanTree(q.pool.QueryRow(ctx, `
		SELECT id, owner_id, title, created_at FROM trees WHERE id = $1 AND owner_id = $2
	`, arg.ID, arg.OwnerID))
}

func (q *Queries) GetTreeByID(ctx context.Context, id uuid.UUID) (Tree, error) {
	return scanTree(q.pool.QueryRow(ctx, `
		SELECT id, owner_id, title, created_at FROM trees WHERE id = $1
	`, id))
}

func (q *Queries) DeleteTree(ctx context.Context, arg DeleteTreeParams) error {
	tag, err := q.pool.Exec(ctx, `DELETE FROM trees WHERE id = $1 AND owner_id = $2`, arg.ID, arg.OwnerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanTree(row pgx.Row) (Tree, error) {
	var tree Tree
	err := row.Scan(&tree.ID, &tree.OwnerID, &tree.Title, &tree.CreatedAt)
	return tree, err
}

type CreatePersonParams struct {
	TreeID    uuid.UUID
	FirstName string
	LastName  string
	Gender    string
	BirthDate sql.NullTime
	DeathDate sql.NullTime
	PhotoURL  string
}

type GetPersonInTreeParams struct {
	ID     uuid.UUID
	TreeID uuid.UUID
}

type UpdatePersonParams struct {
	ID        uuid.UUID
	FirstName string
	LastName  sql.NullString
	Gender    string
	BirthDate sql.NullTime
	DeathDate sql.NullTime
	PhotoURL  sql.NullString
}

func (q *Queries) CreatePerson(ctx context.Context, arg CreatePersonParams) (Person, error) {
	return scanPerson(q.pool.QueryRow(ctx, `
		INSERT INTO persons (tree_id, first_name, last_name, gender, birth_date, death_date, photo_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tree_id, first_name, last_name, gender, birth_date, death_date, photo_url, created_at
	`, arg.TreeID, arg.FirstName, nullableString(arg.LastName), arg.Gender, arg.BirthDate, arg.DeathDate, nullableString(arg.PhotoURL)))
}

func (q *Queries) GetPerson(ctx context.Context, id uuid.UUID) (Person, error) {
	return scanPerson(q.pool.QueryRow(ctx, `
		SELECT id, tree_id, first_name, last_name, gender, birth_date, death_date, photo_url, created_at
		FROM persons WHERE id = $1
	`, id))
}

func (q *Queries) GetPersonInTree(ctx context.Context, arg GetPersonInTreeParams) (Person, error) {
	return scanPerson(q.pool.QueryRow(ctx, `
		SELECT id, tree_id, first_name, last_name, gender, birth_date, death_date, photo_url, created_at
		FROM persons WHERE id = $1 AND tree_id = $2
	`, arg.ID, arg.TreeID))
}

func (q *Queries) ListPersonsByTree(ctx context.Context, treeID uuid.UUID) ([]Person, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, tree_id, first_name, last_name, gender, birth_date, death_date, photo_url, created_at
		FROM persons WHERE tree_id = $1 ORDER BY created_at
	`, treeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		person, err := scanPerson(rows)
		if err != nil {
			return nil, err
		}
		persons = append(persons, person)
	}
	return persons, rows.Err()
}

func (q *Queries) UpdatePerson(ctx context.Context, arg UpdatePersonParams) (Person, error) {
	return scanPerson(q.pool.QueryRow(ctx, `
		UPDATE persons
		SET first_name = COALESCE(NULLIF($2, ''), first_name),
			last_name = COALESCE($3, last_name),
			gender = COALESCE(NULLIF($4, ''), gender),
			birth_date = COALESCE($5, birth_date),
			death_date = COALESCE($6, death_date),
			photo_url = COALESCE($7, photo_url)
		WHERE id = $1
		RETURNING id, tree_id, first_name, last_name, gender, birth_date, death_date, photo_url, created_at
	`, arg.ID, arg.FirstName, arg.LastName, arg.Gender, arg.BirthDate, arg.DeathDate, arg.PhotoURL))
}

func (q *Queries) DeletePerson(ctx context.Context, id uuid.UUID) error {
	tag, err := q.pool.Exec(ctx, `DELETE FROM persons WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanPerson(row pgx.Row) (Person, error) {
	var person Person
	var lastName, photoURL sql.NullString
	var birthDate, deathDate sql.NullTime
	err := row.Scan(
		&person.ID,
		&person.TreeID,
		&person.FirstName,
		&lastName,
		&person.Gender,
		&birthDate,
		&deathDate,
		&photoURL,
		&person.CreatedAt,
	)
	if err != nil {
		return person, err
	}
	if lastName.Valid {
		person.LastName = lastName.String
	}
	if photoURL.Valid {
		person.PhotoURL = photoURL.String
	}
	if birthDate.Valid {
		t := birthDate.Time
		person.BirthDate = &t
	}
	if deathDate.Valid {
		t := deathDate.Time
		person.DeathDate = &t
	}
	return person, nil
}

type CreateRelationParams struct {
	TreeID       uuid.UUID
	Person1ID    uuid.UUID
	Person2ID    uuid.UUID
	RelationType string
}

func (q *Queries) CreateRelation(ctx context.Context, arg CreateRelationParams) (Relation, error) {
	return scanRelation(q.pool.QueryRow(ctx, `
		INSERT INTO relations (tree_id, person1_id, person2_id, relation_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tree_id, person1_id, person2_id, relation_type, created_at
	`, arg.TreeID, arg.Person1ID, arg.Person2ID, arg.RelationType))
}

func (q *Queries) ListRelationsByTree(ctx context.Context, treeID uuid.UUID) ([]Relation, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, tree_id, person1_id, person2_id, relation_type, created_at
		FROM relations WHERE tree_id = $1 ORDER BY created_at
	`, treeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []Relation
	for rows.Next() {
		relation, err := scanRelation(rows)
		if err != nil {
			return nil, err
		}
		relations = append(relations, relation)
	}
	return relations, rows.Err()
}

func (q *Queries) GetRelation(ctx context.Context, id uuid.UUID) (Relation, error) {
	return scanRelation(q.pool.QueryRow(ctx, `
		SELECT id, tree_id, person1_id, person2_id, relation_type, created_at
		FROM relations WHERE id = $1
	`, id))
}

func (q *Queries) DeleteRelation(ctx context.Context, id uuid.UUID) error {
	tag, err := q.pool.Exec(ctx, `DELETE FROM relations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanRelation(row pgx.Row) (Relation, error) {
	var relation Relation
	err := row.Scan(&relation.ID, &relation.TreeID, &relation.Person1ID, &relation.Person2ID, &relation.RelationType, &relation.CreatedAt)
	return relation, err
}

func nullableString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
