package db

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// This file implements the MINIMAL migration-runner prerequisite required to
// safely apply Issue #33's tenant-identity migration (see ADR 0001). It
// evolves the previous one-off "always run 002_ai_reporting.sql" behaviour
// into a deterministic, ordered runner per ADR 0007, but intentionally does
// NOT implement the full ADR 0007 / Issue #40 scope. Deliberately excluded
// as out of scope here: checksum verification of applied migrations,
// down/rollback migrations, and concurrent-startup locking. Those remain
// tracked under Issue #40.

var migrationSeqPattern = regexp.MustCompile(`^(\d+)_`)

// migrationsDir locates the authoritative migrations directory. It tries a
// path relative to the backend module root first, then relative to the repo
// root, matching how the previous implementation located files regardless
// of the process's working directory.
func migrationsDir() (string, error) {
	candidates := []string{"database/migrations", "backend/database/migrations"}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("could not locate migrations directory (tried: %v)", candidates)
}

// ensureMigrationsTable creates the migration-tracking table if it does not
// already exist. This is intentionally a separate table from
// schema_version (which was the legacy bootstrap tracking table used by the
// old InitializeSchema approach, now removed). ADR 0007 requires a dedicated
// migration-history table.
func ensureMigrationsTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("could not create schema_migrations table: %w", err)
	}
	return nil
}

type migrationFile struct {
	version int
	name    string
	path    string
}

// discoverMigrations reads the migrations directory and returns every
// *.sql file, sorted in ascending order by the numeric prefix in its
// filename. Two files claiming the same sequence number is a deterministic
// error (ADR 0007: "must stop execution rather than being selected
// arbitrarily") rather than an arbitrary pick.
func discoverMigrations(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read migrations directory %q: %w", dir, err)
	}

	var files []migrationFile
	seen := map[int]string{}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
			continue
		}

		m := migrationSeqPattern.FindStringSubmatch(e.Name())
		if m == nil {
			return nil, fmt.Errorf("migration file %q does not start with a numeric sequence prefix", e.Name())
		}

		version, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("migration file %q has an unparseable sequence prefix: %w", e.Name(), err)
		}

		if prior, exists := seen[version]; exists {
			return nil, fmt.Errorf("duplicate migration sequence %d: both %q and %q claim it; rename one before the runner can proceed", version, prior, e.Name())
		}
		seen[version] = e.Name()

		files = append(files, migrationFile{
			version: version,
			name:    e.Name(),
			path:    filepath.Join(dir, e.Name()),
		})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	return files, nil
}

// appliedVersions returns the set of migration versions already recorded as
// successfully applied, so they are never re-executed.
func appliedVersions() (map[int]bool, error) {
	rows, err := DB.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("could not read schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("could not scan schema_migrations row: %w", err)
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// ApplyMigrations discovers every migration file in the authoritative
// migrations directory, executes the ones not yet recorded as applied in
// ascending numeric order, and records each success in schema_migrations.
// Execution stops immediately at the first failure; a failed migration is
// never recorded as applied and later migrations do not run (ADR 0007).
// The same function runs identically on a fresh database (applies every
// migration) and on an upgrade (applies only pending ones).
func ApplyMigrations() error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}
	return applyMigrationsFromDir(dir)
}

// applyMigrationsFromDir contains the actual runner logic, parameterized by
// directory so it can be exercised against a scratch migrations directory in
// tests without touching the repository's real migration set.
func applyMigrationsFromDir(dir string) error {
	if err := ensureMigrationsTable(); err != nil {
		return err
	}

	files, err := discoverMigrations(dir)
	if err != nil {
		return err
	}

	applied, err := appliedVersions()
	if err != nil {
		return err
	}

	for _, f := range files {
		if applied[f.version] {
			continue
		}
		if err := applyOne(f); err != nil {
			return err
		}
		fmt.Printf("Applied migration %d (%s)\n", f.version, f.name)
	}

	return nil
}

func applyOne(f migrationFile) error {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return fmt.Errorf("migration %d (%s): could not read file: %w", f.version, f.name, err)
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("migration %d (%s): could not start transaction: %w", f.version, f.name, err)
	}

	if _, err := tx.Exec(string(data)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("migration %d (%s) failed, stopping before later migrations: %w", f.version, f.name, err)
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		f.version, f.name,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("migration %d (%s): could not record success: %w", f.version, f.name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migration %d (%s): could not commit: %w", f.version, f.name, err)
	}

	return nil
}
