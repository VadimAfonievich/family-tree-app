package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

var migrationVersionPattern = regexp.MustCompile(`^(\d+)_.*\.sql$`)

func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string, log zerolog.Logger) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	if err := adoptGooseVersions(ctx, pool); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		version, ok := migrationVersion(file)
		if !ok {
			continue
		}

		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %d: %w", version, err)
		}
		if exists {
			continue
		}

		raw, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}
		upSQL := gooseUpSQL(string(raw))
		if strings.TrimSpace(upSQL) == "" {
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", version, err)
		}
		if _, err = tx.Exec(ctx, upSQL); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d: %w", version, err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %d: %w", version, err)
		}
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}
		log.Info().Int64("version", version).Str("file", filepath.Base(file)).Msg("migration applied")
	}

	return nil
}

func adoptGooseVersions(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		return fmt.Errorf("count schema_migrations: %w", err)
	}
	if count > 0 {
		return nil
	}

	var gooseExists bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&gooseExists); err != nil {
		return fmt.Errorf("check goose_db_version: %w", err)
	}
	if !gooseExists {
		return nil
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO schema_migrations (version, applied_at)
		SELECT version_id, NOW()
		FROM goose_db_version
		WHERE is_applied = true
		ON CONFLICT (version) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("adopt goose migrations: %w", err)
	}
	return nil
}

func migrationVersion(file string) (int64, bool) {
	matches := migrationVersionPattern.FindStringSubmatch(filepath.Base(file))
	if len(matches) != 2 {
		return 0, false
	}
	version, err := strconv.ParseInt(matches[1], 10, 64)
	return version, err == nil
}

func gooseUpSQL(raw string) string {
	afterUp := raw
	if parts := strings.SplitN(raw, "-- +goose Up", 2); len(parts) == 2 {
		afterUp = parts[1]
	}
	if parts := strings.SplitN(afterUp, "-- +goose Down", 2); len(parts) == 2 {
		return parts[0]
	}
	return afterUp
}
