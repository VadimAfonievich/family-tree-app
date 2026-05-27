package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://familytree:familytree@postgres:5432/familytree?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Create a test user
	telegramID := int64(123456789)
	username := "testuser"

	var userID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO users (telegram_id, username) 
		VALUES ($1, $2) 
		ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username
		RETURNING id
	`, telegramID, username).Scan(&userID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created user:", userID)

	// Create a test tree
	var treeID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO trees (owner_id, title) VALUES ($1, 'My Family') RETURNING id
	`, userID).Scan(&treeID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created tree:", treeID)

	// Create test persons
	var grandfatherID, grandmotherID, fatherID, motherID, childID uuid.UUID

	pool.QueryRow(ctx, `
		INSERT INTO persons (tree_id, first_name, last_name, gender, birth_date)
		VALUES ($1, 'Ivan', 'Petrov', 'male', '1950-03-15') RETURNING id
	`, treeID).Scan(&grandfatherID)

	pool.QueryRow(ctx, `
		INSERT INTO persons (tree_id, first_name, last_name, gender, birth_date)
		VALUES ($1, 'Maria', 'Petrova', 'female', '1952-07-20') RETURNING id
	`, treeID).Scan(&grandmotherID)

	pool.QueryRow(ctx, `
		INSERT INTO persons (tree_id, first_name, last_name, gender, birth_date)
		VALUES ($1, 'Alexey', 'Petrov', 'male', '1975-01-10') RETURNING id
	`, treeID).Scan(&fatherID)

	pool.QueryRow(ctx, `
		INSERT INTO persons (tree_id, first_name, last_name, gender, birth_date)
		VALUES ($1, 'Elena', 'Petrova', 'female', '1978-05-25') RETURNING id
	`, treeID).Scan(&motherID)

	pool.QueryRow(ctx, `
		INSERT INTO persons (tree_id, first_name, last_name, gender, birth_date)
		VALUES ($1, 'Dmitry', 'Petrov', 'male', '2005-11-02') RETURNING id
	`, treeID).Scan(&childID)

	fmt.Println("Created 5 persons")

	// Create relations
	pool.Exec(ctx, `INSERT INTO relations (tree_id, person1_id, person2_id, relation_type) VALUES ($1, $2, $3, 'spouse')`,
		treeID, grandfatherID, grandmotherID)
	pool.Exec(ctx, `INSERT INTO relations (tree_id, person1_id, person2_id, relation_type) VALUES ($1, $2, $3, 'parent')`,
		treeID, grandfatherID, fatherID)
	pool.Exec(ctx, `INSERT INTO relations (tree_id, person1_id, person2_id, relation_type) VALUES ($1, $2, $3, 'parent')`,
		treeID, grandmotherID, fatherID)
	pool.Exec(ctx, `INSERT INTO relations (tree_id, person1_id, person2_id, relation_type) VALUES ($1, $2, $3, 'spouse')`,
		treeID, fatherID, motherID)
	pool.Exec(ctx, `INSERT INTO relations (tree_id, person1_id, person2_id, relation_type) VALUES ($1, $2, $3, 'parent')`,
		treeID, fatherID, childID)
	pool.Exec(ctx, `INSERT INTO relations (tree_id, person1_id, person2_id, relation_type) VALUES ($1, $2, $3, 'parent')`,
		treeID, motherID, childID)

	fmt.Println("Created 6 relations")
	fmt.Println("Seed completed successfully!")

	_ = time.Now() // unused import fix
}
