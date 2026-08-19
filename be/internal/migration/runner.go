package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Up applies every pending migration in its own transaction. It is safe to call
// both from the deploy migration job and during API startup.
func Up(ctx context.Context, pool *pgxpool.Pool) error {
	dir, err := findDirectory()
	if err != nil {
		return err
	}
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT NOW())`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}
	sort.Strings(files)
	for _, file := range files {
		version := strings.TrimSuffix(filepath.Base(file), ".up.sql")
		var applied bool
		if err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			continue
		}
		sql, readErr := os.ReadFile(file)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", file, readErr)
		}
		tx, beginErr := pool.Begin(ctx)
		if beginErr != nil {
			return fmt.Errorf("begin %s: %w", file, beginErr)
		}
		if _, err = tx.Exec(ctx, string(sql)); err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, version)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", file, err)
		}
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", file, err)
		}
	}
	var schemaReady bool
	if err = pool.QueryRow(ctx, `SELECT to_regclass('public.users') IS NOT NULL AND to_regclass('public.spirits') IS NOT NULL`).Scan(&schemaReady); err != nil {
		return fmt.Errorf("verify schema: %w", err)
	}
	if !schemaReady {
		return fmt.Errorf("migration history is inconsistent: baseline is recorded but core tables are missing")
	}
	return nil
}

func findDirectory() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); configured != "" {
		return configured, nil
	}
	executable, _ := os.Executable()
	for _, candidate := range []string{"migrations", filepath.Join("be", "migrations"), filepath.Join(filepath.Dir(executable), "migrations")} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("migration directory not found; set MIGRATIONS_DIR")
}
