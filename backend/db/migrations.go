package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// This file owns the application's deterministic database migration path.
// Migrations are ordered by their numeric filename prefix, recorded in
// schema_migrations, and protected by content checksums. Concurrent
// application startups are serialized with a PostgreSQL advisory lock.
var migrationSeqPattern = regexp.MustCompile(`^(\\d+)_`)

const migrationLockKey = "skill-sifter:migrations"

// migrationsDir locates the authoritative migrations directory. It tries a
// path relative to the backend module root first, then relative to the repo
// root, matching how the application is run in local development and CI.
func migrationsDir() (string, error) {
	candidates := []string{"database/migrations", "backend/database/migrations"}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("could not locate migrations directory (tried: %v)", candidates)
}

// ensureMigrationsTable creates the migration history table and upgrades the
// pre-checksum form created by the earlier deterministic runner. Existing
// rows with a NULL checksum are baselined against the current migration file
// on their first run after this upgrade.
func ensureMigrationsTable() error {
	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			checksum   VARCHAR(64),
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("could not create schema_migrations table: %w", err)
	}

	if _, err := DB.Exec(`
		ALTER TABLE schema_migrations
		ADD COLUMN IF NOT EXISTS checksum VARCHAR(64)
	`); err != nil {
		return fmt.Errorf("could not add migration checksum column: %w", err)
	}

	return nil
}

type migrationFile struct {
	version int
	name    string
	path    string
}

type appliedMigration struct {
	version  int
	name     string
	checksum sql.NullString
}

// discoverMigrations reads every SQL migration, validates its numeric
// sequence, and returns the files in ascending execution order.
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

func migrationChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func appliedMigrations() (map[int]appliedMigration, error) {
	rows, err := DB.Query(`
		SELECT version, name, checksum
		FROM schema_migrations
		ORDER BY version
	`)
	if err != nil {
		return nil, fmt.Errorf("could not read schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[int]appliedMigration{}
	for rows.Next() {
		var m appliedMigration
		if err := rows.Scan(&m.version, &m.name, &m.checksum); err != nil {
			return nil, fmt.Errorf("could not scan schema_migrations row: %w", err)
		}
		applied[m.version] = m
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not read schema_migrations rows: %w", err)
	}
	return applied, nil
}

// appliedVersions is retained as a small compatibility helper for the
// migration tests and callers that only need to know whether a version ran.
func appliedVersions() (map[int]bool, error) {
	migrations, err := appliedMigrations()
	if err != nil {
		return nil, err
	}

	applied := make(map[int]bool, len(migrations))
	for version := range migrations {
		applied[version] = true
	}
	return applied, nil
}

// withMigrationLock serializes migration execution across application
// instances. The lock is session-scoped and therefore must be acquired and
// released on the same dedicated database connection.
func withMigrationLock(fn func() error) error {
	conn, err := DB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("could not acquire migration database connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(context.Background(),
		`SELECT pg_advisory_lock(hashtext($1)::bigint)`,
		migrationLockKey,
	); err != nil {
		return fmt.Errorf("could not acquire migration lock: %w", err)
	}

	defer func() {
		_, _ = conn.ExecContext(context.Background(),
			`SELECT pg_advisory_unlock(hashtext($1)::bigint)`,
			migrationLockKey,
		)
	}()

	return fn()
}

// ApplyMigrations discovers every migration file in the authoritative
// migrations directory, validates applied migration history, executes pending
// migrations in ascending order, and records each success atomically.
func ApplyMigrations() error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}
	return applyMigrationsFromDir(dir)
}

// applyMigrationsFromDir is parameterized by directory so migration behavior
// can be tested with scratch migrations without modifying repository SQL.
func applyMigrationsFromDir(dir string) error {
	return withMigrationLock(func() error {
		if err := ensureMigrationsTable(); err != nil {
			return err
		}

		files, err := discoverMigrations(dir)
		if err != nil {
			return err
		}

		applied, err := appliedMigrations()
		if err != nil {
			return err
		}

		filesByVersion := make(map[int]migrationFile, len(files))
		for _, f := range files {
			filesByVersion[f.version] = f
		}

		// Every recorded migration must still exist and retain its original
		// filename and content. A missing or modified migration is unsafe to
		// silently ignore because it changes the database history contract.
		for version, recorded := range applied {
			f, exists := filesByVersion[version]
			if !exists {
				return fmt.Errorf("applied migration %d (%s) is missing from the migrations directory", version, recorded.name)
			}
			if f.name != recorded.name {
				return fmt.Errorf("migration %d filename mismatch: database records %q, filesystem contains %q", version, recorded.name, f.name)
			}

			checksum, err := migrationChecksum(f.path)
			if err != nil {
				return fmt.Errorf("migration %d (%s): could not calculate checksum: %w", version, f.name, err)
			}

			if recorded.checksum.Valid {
				if recorded.checksum.String != checksum {
					return fmt.Errorf("migration %d (%s) checksum mismatch: database=%s filesystem=%s", version, f.name, recorded.checksum.String, checksum)
				}
				continue
			}

			// Compatibility bridge for databases created by the earlier runner,
			// which tracked version/name but not checksums. The current file is
			// explicitly baselined once; future changes are then detected.
			if _, err := DB.Exec(
				`UPDATE schema_migrations SET checksum = $1 WHERE version = $2 AND checksum IS NULL`,
				checksum, version,
			); err != nil {
				return fmt.Errorf("migration %d (%s): could not baseline checksum: %w", version, f.name, err)
			}
		}

		for _, f := range files {
			if _, alreadyApplied := applied[f.version]; alreadyApplied {
				continue
			}
			if err := applyOne(f); err != nil {
				return err
			}
			fmt.Printf("Applied migration %d (%s)\n", f.version, f.name)
		}

		return nil
	})
}

func applyOne(f migrationFile) error {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return fmt.Errorf("migration %d (%s): could not read file: %w", f.version, f.name, err)
	}

	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("migration %d (%s): could not start transaction: %w", f.version, f.name, err)
	}

	if _, err := tx.Exec(string(data)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("migration %d (%s) failed, stopping before later migrations: %w", f.version, f.name, err)
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)`,
		f.version, f.name, checksum,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("migration %d (%s): could not record success: %w", f.version, f.name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migration %d (%s): could not commit: %w", f.version, f.name, err)
	}

	return nil
}
